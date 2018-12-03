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
		fmt.Fprintf(os.Stderr, "failed to create default detector: %s", err)
		os.Exit(100)
	}

	code, err := detector.Pass(buildplan.BuildPlan{})
	if err != nil {
		detector.Logger.Info(err.Error())
	}

	os.Exit(code)
}
