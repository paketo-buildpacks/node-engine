package main

import (
	"github.com/cloudfoundry/packit"
	"github.com/paketo-buildpacks/node-engine/node"
)

func main() {
	nvmrcParser := node.NewNvmrcParser()
	buildpackYMLParser := node.NewBuildpackYMLParser()

	packit.Detect(node.Detect(nvmrcParser, buildpackYMLParser))
}
