package main

import (
	"os"

	nodeengine "github.com/paketo-buildpacks/node-engine/v4"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	nvmrcParser := nodeengine.NewNvmrcParser()
	nodeVersionParser := nodeengine.NewNodeVersionParser()
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	sbomGenerator := Generator{}

	packit.Run(
		nodeengine.Detect(
			nvmrcParser,
			nodeVersionParser,
		),
		nodeengine.Build(
			entryResolver,
			dependencyManager,
			sbomGenerator,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
