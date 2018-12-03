package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/nodejs-cnb/node"
)

func main() {
	builder, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default builder: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(builder)
	if err != nil {
		builder.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(builder build.Build) (int, error) {
	builder.Logger.FirstLine(builder.Logger.PrettyIdentity(builder.Buildpack))

	nodeContributor, willContribute, err := node.NewContributor(builder)
	if err != nil {
		return builder.Failure(102), err
	}

	if willContribute {
		if err := nodeContributor.Contribute(); err != nil {
			return builder.Failure(103), err
		}
	}

	return builder.Success(buildplan.BuildPlan{})
}
