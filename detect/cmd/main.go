package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack"
)

func main() {
	detector, err := libbuildpack.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create default detector: %s", err)
		os.Exit(100)
	}

	detector.Pass(libbuildpack.BuildPlan{})
}
