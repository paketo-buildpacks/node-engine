package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/cloudfoundry/nodejs-cnb/internal/build"
)

func main() {
	builder, err := libjavabuildpack.DefaultBuild()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create default builder: %s", err)
		os.Exit(100)
	}

	node, ok, err := build.NewNode(builder)
	if err != nil {
		builder.Logger.Info(err.Error())
		builder.Failure(102)
		return
	}

	if ok {
		if err := node.Contribute(); err != nil {
			builder.Logger.Info(err.Error())
			builder.Failure(103)
			return
		}
	}

	builder.Success()
}
