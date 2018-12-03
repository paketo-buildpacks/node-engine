package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	detector, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default detector: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(detector)
	if err != nil {
		detector.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(detector detect.Detect) (int, error) {
	return detector.Pass(buildplan.BuildPlan{})
}
