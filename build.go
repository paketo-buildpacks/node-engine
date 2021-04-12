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
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface EnvironmentConfiguration --output fakes/environment_configuration.go
type EnvironmentConfiguration interface {
	Configure(buildEnv, launchEnv packit.Environment, path string, optimizeMemory bool) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan
}

func Build(entries EntryResolver, dependencies DependencyManager, environment EnvironmentConfiguration, planRefinery BuildPlanRefinery, logger LogEmitter, clock chronos.Clock) packit.BuildFunc {
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

		entry, allEntries := entries.Resolve("node", context.Plan.Entries, priorities)
		logger.Candidates(allEntries)

		version, _ := entry.Metadata["version"].(string)
		var dependency postal.Dependency
		var err error
		dependency, err = dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

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

		bom := planRefinery.BillOfMaterial(postal.Dependency{
			ID:      dependency.ID,
			Name:    dependency.Name,
			SHA256:  dependency.SHA256,
			Stacks:  dependency.Stacks,
			URI:     dependency.URI,
			Version: dependency.Version,
		})

		cachedSHA, ok := nodeLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", nodeLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{nodeLayer},
			}, nil
		}

		logger.Process("Executing build process")

		nodeLayer, err = nodeLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		nodeLayer.Launch, nodeLayer.Build = entries.MergeLayerTypes("node", context.Plan.Entries)
		nodeLayer.Cache = nodeLayer.Build

		nodeLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		logger.Subprocess("Installing Node Engine %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, nodeLayer.Path)
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
			Plan:   bom,
			Layers: []packit.Layer{nodeLayer},
		}, nil
	}
}
