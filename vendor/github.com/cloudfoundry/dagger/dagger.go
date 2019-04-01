package dagger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/pkg/errors"
)

const (
	CFLINUXFS3          = "org.cloudfoundry.stacks.cflinuxfs3"
	BIONIC              = "io.buildpacks.stacks.bionic"
	DEFAULT_BUILD_IMAGE = "cfbuildpacks/cflinuxfs3-cnb-experimental:build"
	DEFAULT_RUN_IMAGE   = "cfbuildpacks/cflinuxfs3-cnb-experimental:run"
)

var downloadCache sync.Map

func init() {
	rand.Seed(time.Now().UnixNano())
	downloadCache = sync.Map{}
}

func PackageBuildpack() (string, error) {
	cmd := exec.Command("./scripts/package.sh")
	cmd.Dir = "../"
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	r := regexp.MustCompile("Buildpack packaged into: (.*)")
	bpDir := r.FindStringSubmatch(string(out))[1]
	return bpDir, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetClient(ctx context.Context) *github.Client {
	git_token := os.Getenv("GIT_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: git_token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(http.DefaultClient)
	if git_token == "" {
		fmt.Println("Using unauthorized github api, consider setting the GIT_TOKEN environment variable")
		fmt.Println("More info on Github tokens here: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line")
		client = github.NewClient(tc)
	}

	return client
}

func TempBuildpackPath(name string) string {
	return filepath.Join("/tmp", name+"-"+RandStringRunes(16))
}

func PackageCachedBuildpack(bpPath string) (string, string, error) {
	tarFile := TempBuildpackPath(filepath.Base(bpPath)) // + ".tgz"
	cmd := exec.Command("./.bin/packager", tarFile)
	cmd.Dir = bpPath
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return tarFile, string(out), err
}

func PackageLocalBuildpack(name, path string) (string, error) {
	cmd := exec.Command("./scripts/package.sh")
	cmd.Dir = path
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	r := regexp.MustCompile("Buildpack packaged into: (.*)")
	bpDir := r.FindStringSubmatch(string(out))[1]
	return bpDir, nil
}

func GetLatestBuildpack(name string) (string, error) {
	uri := fmt.Sprintf("https://api.github.com/repos/cloudfoundry/%s/releases/latest", name)
	ctx := context.Background()
	client := GetClient(ctx)

	release := struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}{}
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}
	if _, err := client.Do(ctx, request, &release); err != nil {
		return "", err
	}
	if len(release.Assets) == 0 {
		return "", fmt.Errorf("there are no releases for %s", name)
	}

	contents, found := downloadCache.Load(name + release.TagName)
	if !found {
		buildpackResp, err := http.Get(release.Assets[0].BrowserDownloadURL)
		if err != nil {
			return "", err
		}

		defer buildpackResp.Body.Close()

		contents, err = ioutil.ReadAll(buildpackResp.Body)
		if err != nil {
			return "", err
		}

		if buildpackResp.StatusCode != http.StatusOK {
			return "", errors.Errorf("Erroring Getting buildpack : status %d : %s", buildpackResp.StatusCode, contents)
		}

		downloadCache.Store(name+release.TagName, contents)
	}

	downloadFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(downloadFile.Name())

	_, err = io.Copy(downloadFile, bytes.NewReader(contents.([]byte)))
	if err != nil {
		return "", err
	}

	dest, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	return dest, helper.ExtractTarGz(downloadFile.Name(), dest, 0)
}

// This returns the build logs as part of the error case
func PackBuild(appDir string, buildpacks ...string) (*App, error) {
	return PackBuildNamedImage(randomString(16), appDir, buildpacks...)
}

// This pack builds an app from appDir into appImageName, to allow specifying an image name in a test
func PackBuildNamedImage(appImageName, appDir string, buildpacks ...string) (*App, error) {
	buildLogs := &bytes.Buffer{}

	cmd := exec.Command("pack", "build", appImageName, "--builder", "cfbuildpacks/cflinuxfs3-cnb-test-builder")
	for _, bp := range buildpacks {
		cmd.Args = append(cmd.Args, "--buildpack", bp)
	}
	cmd.Dir = appDir
	cmd.Stdout = io.MultiWriter(os.Stdout, buildLogs)
	cmd.Stderr = io.MultiWriter(os.Stderr, buildLogs)
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, buildLogs.String())
	}

	app := &App{
		buildLogs:   buildLogs,
		Env:         make(map[string]string),
		imageName:   appImageName,
		fixtureName: appDir,
	}
	return app, nil
}

type App struct {
	Memory      string
	buildLogs   *bytes.Buffer
	Env         map[string]string
	logProc     *exec.Cmd
	imageName   string
	containerId string
	port        string
	fixtureName string
	healthCheck HealthCheck
}

type HealthCheck struct {
	command  string
	interval string
	timeout  string
}

func (a *App) BuildLogs() string {
	return stripColor(a.buildLogs.String())
}

func (a *App) SetHealthCheck(command, interval, timeout string) {
	a.healthCheck = HealthCheck{
		command:  command,
		interval: interval,
		timeout:  timeout,
	}
}

func (a *App) Start() error {
	buf := &bytes.Buffer{}

	args := []string{"run", "-d", "-P"}
	if a.Memory != "" {
		args = append(args, "--memory", a.Memory)
	}

	if a.healthCheck.command != "" {
		args = append(args, "--health-cmd", a.healthCheck.command)
	}

	if a.healthCheck.interval != "" {
		args = append(args, "--health-interval", a.healthCheck.interval)
	}

	if a.healthCheck.timeout != "" {
		args = append(args, "--health-timeout", a.healthCheck.timeout)
	}

	envTemplate := "%s=%s"
	for k, v := range a.Env {
		envString := fmt.Sprintf(envTemplate, k, v)
		args = append(args, "-e", envString)
	}

	args = append(args, a.imageName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	a.containerId = buf.String()[:12]

	ticker := time.NewTicker(1 * time.Second)
	timeOut := time.After(2 * time.Minute)
docker:
	for {
		select {
		case <-ticker.C:
			status, err := exec.Command("docker", "inspect", "-f", "{{.State.Health.Status}}", a.containerId).Output()
			if err != nil {
				return err
			}

			if strings.TrimSpace(string(status)) == "unhealthy" {
				return fmt.Errorf("app failed to start : %s", a.fixtureName)
			}

			if strings.TrimSpace(string(status)) == "healthy" {
				break docker
			}
		case <-timeOut:
			return fmt.Errorf("timed out waiting for app : %s", a.fixtureName)
		}
	}

	cmd = exec.Command("docker", "container", "port", a.containerId)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	a.port = strings.TrimSpace(strings.Split(buf.String(), ":")[1])

	return nil
}

func (a *App) Destroy() error {
	if a.containerId == "" {
		return nil
	}

	cmd := exec.Command("docker", "stop", a.containerId)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("docker", "rm", a.containerId, "-f", "--volumes")
	if err := cmd.Run(); err != nil {
		return err
	}

	a.containerId = ""
	a.port = ""

	if a.imageName == "" {
		return nil
	}

	cmd = exec.Command("docker", "rmi", a.imageName, "-f")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("docker", "image", "prune", "-f")
	if err := cmd.Run(); err != nil {
		return err
	}

	a.imageName = ""
	return nil
}

func (a *App) Files(path string) ([]string, error) {
	cmd := exec.Command("docker", "run", a.imageName, "find", "./..", "-wholename", fmt.Sprintf("*%s*", path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}
	return strings.Split(string(output), "\n"), nil
}

func (a *App) Info() (cID string, imageID string, cacheID []string, e error) {
	volumes, err := getCacheVolumes()
	if err != nil {
		return "", "", []string{}, err
	}

	return a.containerId, a.imageName, volumes, nil
}

func (a *App) Logs() (string, error) {
	cmd := exec.Command("docker", "logs", a.containerId)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return stripColor(string(output)), nil
}

func (a *App) HTTPGet(path string) (string, map[string][]string, error) {
	resp, err := http.Get("http://localhost:" + a.port + path)
	if err != nil {
		return "", nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("received bad response from application")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return string(body), resp.Header, nil
}

func (a *App) HTTPGetBody(path string) (string, error) {
	resp, _, err := a.HTTPGet(path)
	return resp, err
}

func stripColor(input string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(input, "")
}

func getCacheVolumes() ([]string, error) {
	cmd := exec.Command("docker", "volume", "ls", "-q")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	outputArr := strings.Split(string(output), "\n")
	var finalVolumes []string
	for _, line := range outputArr {
		if strings.Contains(line, "pack-cache") {
			finalVolumes = append(finalVolumes, line)
		}
	}
	return outputArr, nil
}

func randomString(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
