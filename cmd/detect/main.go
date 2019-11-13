package main

import (
	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
)

func main() {
	nvmrcParser := node.NewNvmrcParser()
	buildpackYMLParser := node.NewBuildpackYMLParser()

	packit.Detect(node.Detect(nvmrcParser, buildpackYMLParser))
}
