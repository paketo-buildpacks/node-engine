package dagger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type App struct {
	ImageName   string
	CacheImage  string
	ContainerID string
	Memory      string
	Env         map[string]string
	buildLogs   *bytes.Buffer
	logProc     *exec.Cmd
	port        string
	fixtureName string
	healthCheck HealthCheck
}

type HealthCheck struct {
	command  string
	interval string
	timeout  string
}

func (a *App) Start() error {
	buf := &bytes.Buffer{}

	if a.Env["PORT"] == "" {
		a.Env["PORT"] = "8080"
	}

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

	args = append(args, a.ImageName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	a.ContainerID = buf.String()[:12]

	ticker := time.NewTicker(1 * time.Second)
	timeOut := time.After(2 * time.Minute)
docker:
	for {
		select {
		case <-ticker.C:
			status, err := exec.Command("docker", "inspect", "-f", "{{.State.Health.Status}}", a.ContainerID).Output()
			if err != nil {
				return err
			}

			if strings.TrimSpace(string(status)) == "unhealthy" {
				logs, _ := a.Logs()
				return errors.Errorf("app failed to start: %s\n%s\n", a.fixtureName, logs)
			}

			if strings.TrimSpace(string(status)) == "healthy" {
				break docker
			}
		case <-timeOut:
			return fmt.Errorf("timed out waiting for app : %s", a.fixtureName)
		}
	}

	cmd = exec.Command("docker", "container", "port", a.ContainerID)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "docker error: failed to get port")
	}
	a.port = strings.TrimSpace(strings.Split(buf.String(), ":")[1])

	return nil
}

func (a *App) Destroy() error {
	cntrExists, err := DockerArtifactExists(a.ContainerID)
	if err != nil {
		return err
	}
	if cntrExists {
		cmd := exec.Command("docker", "stop", a.ContainerID)
		if err := cmd.Run(); err != nil {
			return err
		}

		cmd = exec.Command("docker", "rm", a.ContainerID, "-f", "--volumes")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	imgExists, err := DockerArtifactExists(a.ImageName)
	if err != nil {
		return err
	}

	if imgExists {
		cmd := exec.Command("docker", "rmi", a.ImageName, "-f")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	cacheExists, err := DockerArtifactExists(a.CacheImage)
	if err != nil {
		return err
	}

	if cacheExists {
		cmd := exec.Command("docker", "rmi", a.CacheImage, "-f")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	cmd := exec.Command("docker", "image", "prune", "-f")
	if err := cmd.Run(); err != nil {
		return err
	}

	*a = App{}
	return nil
}

func (a *App) Logs() (string, error) {
	cmd := exec.Command("docker", "logs", a.ContainerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return stripColor(string(output)), nil
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

func (a *App) Files(path string) ([]string, error) {
	// Ensures that the error and results from "Permission denied" don't get sent to the output
	line := fmt.Sprintf("docker run %s find ./.. -wholename *%s* 2>&1 | grep -v \"Permission denied\"", a.ImageName, path)
	cmd := exec.Command("bash", "-c", line)
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

	return a.ContainerID, a.ImageName, volumes, nil
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

func DockerArtifactExists(name string) (bool, error) {
	cmd := exec.Command("docker", "inspect", name)
	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "No such object") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
