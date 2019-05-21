package dagger

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	CFLinuxFS3        = "org.cloudfoundry.stacks.cflinuxfs3"
	Bionic            = "io.buildpacks.stacks.bionic"
	DefaultBuildImage = "cloudfoundry/cnb-build:cflinuxfs3"
	DefaultRunImage   = "cloudfoundry/cnb-run:cflinuxfs3"
	TestBuilderImage  = "cfbuildpacks/cflinuxfs3-cnb-test-builder"
)

// This returns the build logs as part of the error case
func PackBuild(appDir string, buildpacks ...string) (*App, error) {
	return PackBuildNamedImage(RandStringRunes(16), appDir, buildpacks...)
}

// This pack builds an app from appDir into appImageName, to allow specifying an image name in a test
func PackBuildNamedImage(appImage, appDir string, bpPaths ...string) (*App, error) {
	buildLogs := &bytes.Buffer{}

	cmd := exec.Command("pack", "build", appImage, "--builder", TestBuilderImage)
	for _, bp := range bpPaths {
		cmd.Args = append(cmd.Args, "--buildpack", bp)
	}

	cmd.Dir = appDir
	cmd.Stdout = io.MultiWriter(os.Stdout, buildLogs)
	cmd.Stderr = io.MultiWriter(os.Stderr, buildLogs)
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, buildLogs.String())
	}

	sum := sha256.Sum256([]byte(fmt.Sprintf("index.docker.io/library/%s:latest", appImage))) //This is how pack makes cache image names
	cacheImage := fmt.Sprintf("pack-cache-%x", sum[:6])

	app := &App{
		ImageName:   appImage,
		CacheImage:  cacheImage,
		buildLogs:   buildLogs,
		Env:         make(map[string]string),
		fixtureName: appDir,
	}
	return app, nil
}
