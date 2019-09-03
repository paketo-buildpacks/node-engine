package dagger

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/pkg/errors"
)

const (
	Tiny              = "org.cloudfoundry.stacks.tiny"
	CFLinuxFS3        = "org.cloudfoundry.stacks.cflinuxfs3"
	Bionic            = "io.buildpacks.stacks.bionic"
	DefaultBuildImage = "cloudfoundry/build:full-cnb"
	DefaultRunImage   = "cloudfoundry/run:full-cnb"
	TestBuilderImage  = "cfbuildpacks/spec-change-builder"
	logBufferSize     = 1024
)

var (
	logQueue                chan chan []byte
	stdoutMutex             sync.Mutex
	queueIsInitialized      bool
	queueIsInitializedMutex sync.Mutex
)

// This returns the build logs as part of the error case
func PackBuild(appDir string, buildpacks ...string) (*App, error) {
	return PackBuildNamedImage(RandStringRunes(16), appDir, buildpacks...)
}

func PackBuildWithEnv(appDir string, env map[string]string, buildpacks ...string) (*App, error) {
	return PackBuildNamedImageWithEnv(RandStringRunes(16), appDir, env, buildpacks...)
}

// This pack builds an app from appDir into appImageName, to allow specifying an image name in a test
func PackBuildNamedImage(appImage, appDir string, bpPaths ...string) (*App, error) {
	return PackBuildNamedImageWithEnv(appImage, appDir, nil, bpPaths...)
}

func PackBuildNamedImageWithEnv(appImage, appDir string, env map[string]string, bpPaths ...string) (*App, error) {
	buildLogs := &bytes.Buffer{}

	cmd := exec.Command("pack", "build", appImage, "--builder", TestBuilderImage)
	for _, bp := range bpPaths {
		cmd.Args = append(cmd.Args, "--buildpack", bp)
	}

	for key, val := range env {
		cmd.Args = append(cmd.Args, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	var w io.Writer
	queueIsInitializedMutex.Lock()
	if queueIsInitialized {
		log := make(chan []byte, logBufferSize)
		logQueue <- log
		cw := newChanWriter(log)
		w = cw
		defer cw.Close()
	} else {
		w = os.Stdout
	}
	queueIsInitializedMutex.Unlock()

	cmd.Dir = appDir
	cmd.Stdout = io.MultiWriter(w, buildLogs)
	cmd.Stderr = io.MultiWriter(w, buildLogs)
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

type chanWriter struct {
	channel chan []byte
}

func newChanWriter(c chan []byte) *chanWriter {
	return &chanWriter{c}
}

func (c *chanWriter) Write(p []byte) (n int, err error) {
	c.channel <- append([]byte{}, p...) // Create a copy to avoid mutation of backing slice
	return len(p), nil
}

func (c *chanWriter) Close() {
	close(c.channel)
}

func SyncParallelOutput(f func()) {
	startOutputStream()
	defer stopOutputStream()
	f()
}

func startOutputStream() {
	fmt.Println("Starting to stream output...")
	logQueue = make(chan chan []byte, 1024) // Arbitrary buffer size to reduce blocking
	queueIsInitializedMutex.Lock()
	queueIsInitialized = true
	queueIsInitializedMutex.Unlock()
	go printLoop()
}

func stopOutputStream() {
	close(logQueue)
	fmt.Println("Stopped streaming output.")
}

func printLoop() {
	for log := range logQueue {
		printLog(log)
	}
}

func printLog(log chan []byte) {
	for line := range log {
		stdoutMutex.Lock()
		fmt.Print(string(line))
		stdoutMutex.Unlock()
	}
}
