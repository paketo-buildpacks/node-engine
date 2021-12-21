package main

import (
	"os"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	nvmrcParser := nodeengine.NewNvmrcParser()
	buildpackYMLParser := nodeengine.NewBuildpackYMLParser()
	nodeVersionParser := nodeengine.NewNodeVersionParser()
	logEmitter := nodeengine.NewLogEmitter(os.Stdout)
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := nodeengine.NewEnvironment(logEmitter)

	packit.Run(
		nodeengine.Detect(
			nvmrcParser,
			buildpackYMLParser,
			nodeVersionParser,
		),
		nodeengine.Build(
			entryResolver,
			dependencyManager,
			environment,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
