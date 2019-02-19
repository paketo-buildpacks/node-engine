package action

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	LINUX  = "linux"
	MACOS  = "macos"
	DARWIN = "darwin"
	PACK   = ".bin/pack"
	DST    = ".bin"
)

func InstallTools(ver string) error {
	flag.Parse()

	args := len(flag.Args())
	if args > 2 {
		flag.Usage()
	}

	version := flag.Arg(1)
	if version == "" {
		if ver == "" {
			version = "latest"
		} else {
			version = ver
		}
	}

	fmt.Println(version)

	if err := installPack(version); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install pack\n")
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}
	return nil
}

type Asset struct {
	DownloadUrl string `json:"browser_download_url"`
}
type GithubRelease struct {
	Assets []Asset `json:"assets"`
}

func installPack(version string) error {
	var OS string

	if OS = runtime.GOOS; OS != DARWIN && OS != LINUX {
		fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", OS)
		os.Exit(1)
	}

	if OS == DARWIN {
		OS = MACOS
	}

	if version != "latest" {
		artifact := fmt.Sprintf("pack-%s-%s.tar.gz", version, OS)
		url := fmt.Sprintf("https://github.com/buildpack/pack/releases/download/v%s/%s", version, artifact)

		return extract(url)
	}

	resp, err := http.Get("https://api.github.com/repos/buildpack/pack/releases/latest")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var release GithubRelease

	json.NewDecoder(resp.Body).Decode(&release)

	if len(release.Assets) < 2 {
		return fmt.Errorf("invalid number of assets from github, please check the following url" +
			"https://api.github.com/repos/buildpack/pack/releases/latest")
	}

	url := release.Assets[0].DownloadUrl
	if OS == MACOS {
		url = release.Assets[1].DownloadUrl
	}
	return extract(url)
}

func extract(rawURL string) error {
	artifact, err := parseArtifact(rawURL)
	if err != nil {
		return err
	}
	version := parseVersion(artifact)
	ok, err := isVersionInstalled(version)
	if err != nil {
		return err
	}
	if ok {
		fmt.Printf("Version %s of pack is already installed\n", version)
		return nil
	}

	fmt.Println("Installing pack", version)

	if err := download(rawURL, artifact); err != nil {
		return err
	}
	defer os.RemoveAll(artifact)

	if err := os.MkdirAll(DST, 0777); err != nil {
		return err
	}

	if err := untar(artifact, DST); err != nil {
		return err
	}

	return nil
}

func download(url, artifact string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(artifact)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func isVersionInstalled(version string) (bool, error) {
	if _, err := os.Stat(PACK); os.IsNotExist(err) {
		return false, nil
	}

	cmd := exec.Command(PACK, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	output := string(out)

	if strings.Contains(output, version) {
		return true, nil
	}
	return false, nil
}

func parseArtifact(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	index := strings.LastIndex(u.Path, "/") + 1
	return u.Path[index:], nil
}

func parseVersion(artifact string) string {
	return strings.Split(artifact, "-")[1]
}

func untar(artifact, dst string) error {
	cmd := exec.Command("tar", "xzvf", artifact, "-C", dst)

	return cmd.Run()
}
