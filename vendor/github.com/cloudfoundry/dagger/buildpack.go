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
	"sync"
	"time"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/pkg/errors"
)

var downloadCache sync.Map

func init() {
	rand.Seed(time.Now().UnixNano())
	downloadCache = sync.Map{}
}

func FindBPRoot() (string, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for {
		if dir == "/" {
			return "", fmt.Errorf("could not find buildpack.toml in the directory hierarchy")
		}
		if exist, err := helper.FileExists(filepath.Join(dir, "buildpack.toml")); err != nil {
			return "", err
		} else if exist {
			return dir, nil
		}
		dir, err = filepath.Abs(filepath.Join(dir, ".."))
		if err != nil {
			return "", err
		}
	}
}

func PackageBuildpack(root string) (string, error) {
	path, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	bpName := fmt.Sprintf("%s_%s", filepath.Base(path), RandStringRunes(8))
	bpPath := filepath.Join(path, bpName)

	cmd := exec.Command("scripts/package.sh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return bpPath, nil
}

func PackageCachedBuildpack(root string) (string, string, error) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return "", "", err
	}

	tarFile := filepath.Join(tmp, filepath.Base(root))
	cmd := exec.Command("./.bin/packager", tarFile)
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return tarFile, string(out), err
}

func GetLatestBuildpack(name string) (string, error) {
	uri := fmt.Sprintf("https://api.github.com/repos/cloudfoundry/%s/releases/latest", name)
	ctx := context.Background()
	client := NewGitClient(ctx)

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
