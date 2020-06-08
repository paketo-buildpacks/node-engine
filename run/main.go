package main

import (
	"os"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	nvmrcParser := nodeengine.NewNvmrcParser()
	buildpackYMLParser := nodeengine.NewBuildpackYMLParser()
	logEmitter := nodeengine.NewLogEmitter(os.Stdout)
	entryResolver := nodeengine.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := nodeengine.NewEnvironment(logEmitter)
	planRefinery := nodeengine.NewPlanRefinery()

	packit.Run(
		nodeengine.Detect(
			nvmrcParser,
			buildpackYMLParser,
		),
		nodeengine.Build(
			entryResolver,
			dependencyManager,
			environment,
			planRefinery,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
