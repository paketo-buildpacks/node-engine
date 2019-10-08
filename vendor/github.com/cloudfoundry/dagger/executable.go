package dagger

import (
	"bytes"
	"github.com/buildpack/libbuildpack/logger"
	executable "github.com/cloudfoundry/libbuildpack/cutlass/docker"
	"os/exec"
)

const PackExecutableName = "pack"

type PackExecutable struct {
	name   string
	logger logger.Logger
}

func NewPackExecutable(logger logger.Logger) PackExecutable {
	return PackExecutable{
		name:   PackExecutableName,
		logger: logger,
	}
}

func (p PackExecutable) Execute(options executable.ExecuteOptions, args ...string) (string, string, error) {
	p.logger.Info("Running Pack")

	packCmd := exec.Command(p.name, args...)
	if options.Dir != "" {
		packCmd.Dir = options.Dir
	}
	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})

	packCmd.Stdout = stdout
	packCmd.Stderr = stderr

	p.logger.Info("")
	if err := packCmd.Run(); err != nil {
		return "", "", err
	}
	return stdout.String(), stderr.String(), nil
}


