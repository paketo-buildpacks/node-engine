package nodeengine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(name string, entries []packit.BuildpackPlanEntry, priorities []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch bool, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface EnvironmentConfiguration --output fakes/environment_configuration.go
type EnvironmentConfiguration interface {
	Configure(buildEnv, launchEnv packit.Environment, path string, optimizeMemory bool) error
}

func Build(entryResolver EntryResolver, dependencyManager DependencyManager, environment EnvironmentConfiguration, logger LogEmitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Node Engine version")

		priorities := []interface{}{
			"BP_NODE_VERSION",
			"buildpack.yml",
			"package.json",
			".nvmrc",
			".node-version",
		}

		entry, allEntries := entryResolver.Resolve("node", context.Plan.Entries, priorities)
		logger.Candidates(allEntries)

		version, _ := entry.Metadata["version"].(string)
		var dependency postal.Dependency
		var err error
		dependency, err = dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())
		bom := dependencyManager.GenerateBillOfMaterials(dependency)

		versionSource, _ := entry.Metadata["version-source"].(string)
		if versionSource == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the Node version through buildpack.yml will be deprecated soon in Node Engine Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_NODE_VERSION environment variable instead. See README.md for more information.")
			logger.Break()
		}

		nodeLayer, err := context.Layers.Get(Node)
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := entryResolver.MergeLayerTypes("node", context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: bom}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: bom}
		}

		cachedSHA, ok := nodeLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", nodeLayer.Path)
			logger.Break()

			nodeLayer.Launch, nodeLayer.Build, nodeLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{nodeLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		nodeLayer, err = nodeLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		nodeLayer.Launch, nodeLayer.Build, nodeLayer.Cache = launch, build, build

		nodeLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		logger.Subprocess("Installing Node Engine %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencyManager.Deliver(dependency, context.CNBPath, nodeLayer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		// Check if buildpack.yml specifies optimize_memory
		config, err := BuildpackYMLParser{}.Parse(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("unable to parse buildpack.yml file: %s", err)
		}
		if config.OptimizedMemory {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Enabling memory optimization through buildpack.yml will be deprecated soon in Node Engine Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please enable through the $BP_NODE_OPTIMIZE_MEMORY environment variable instead. See README.md for more information.")
			logger.Break()
		}

		if os.Getenv("BP_NODE_OPTIMIZE_MEMORY") == "true" {
			config.OptimizedMemory = true
		}

		err = environment.Configure(nodeLayer.BuildEnv, nodeLayer.SharedEnv, nodeLayer.Path, config.OptimizedMemory)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{nodeLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
