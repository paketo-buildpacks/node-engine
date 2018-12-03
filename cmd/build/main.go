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
		fmt.Fprintf(os.Stderr, "failed to create default builder: %s", err)
		os.Exit(101)
	}

	builder.Logger.FirstLine(build.Logger.PrettyIdentity(builder.Buildpack))

	nodeContributor, willContribute, err := node.NewNodeContributor(builder)
	if err != nil {
		builder.Logger.Info(err.Error())
		os.Exit(102)
	}

	if willContribute {
		if err := nodeContributor.Contribute(); err != nil {
			builder.Logger.Info(err.Error())
			os.Exit(103)
		}
	}

	code, err := builder.Success(buildplan.BuildPlan{})
	if err != nil {
		builder.Logger.Info(err.Error())
	}

	os.Exit(code)
}
