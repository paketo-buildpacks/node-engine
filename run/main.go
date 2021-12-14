package main

import (
	"os"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	nvmrcParser := nodeengine.NewNvmrcParser()
	nodeVersionParser := nodeengine.NewNodeVersionParser()
	logEmitter := nodeengine.NewLogEmitter(os.Stdout)
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := nodeengine.NewEnvironment(logEmitter)
	sbomGenerator := Generator{}

	packit.Run(
		nodeengine.Detect(
			nvmrcParser,
			nodeVersionParser,
		),
		nodeengine.Build(
			entryResolver,
			dependencyManager,
			environment,
			sbomGenerator,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
