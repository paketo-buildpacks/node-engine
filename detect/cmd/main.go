package main

import (
	"fmt"
	"os"

	libbuildpackV3 "github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/nodejs-cnb-buildpack/detect"
)

func main() {
	detector, err := libbuildpackV3.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create default detector: %s", err)
		os.Exit(100)
	}

	if err := detect.UpdateBuildPlan(&detector); err != nil {
		detector.Logger.Debug("failed nodejs detection: %s", err)
		detector.Fail()
	}

	detector.Pass(detector.BuildPlan)
}
