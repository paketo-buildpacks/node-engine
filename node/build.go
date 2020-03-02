package node

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(dependencies []BuildpackMetadataDependency, defaultVersion, stack string, entry packit.BuildpackPlanEntry) (BuildpackMetadataDependency, error)
	Install(dependency BuildpackMetadataDependency, cnbPath, layerPath string) error
}

//go:generate faux --interface EnvironmentConfiguration --output fakes/environment_configuration.go
type EnvironmentConfiguration interface {
	Configure(env EnvironmentVariables, path string, optimizeMemory bool) error
}

//go:generate faux --interface PlanRefinery --output fakes/plan_refinery.go
type PlanRefinery interface {
	BillOfMaterial(dependency BuildpackMetadataDependency) packit.BuildpackPlan
}

//go:generate faux --interface CacheManager --output fakes/cache_manager.go
type CacheManager interface {
	Match(layer packit.Layer, dependency BuildpackMetadataDependency) (bool, error)
}

func Build(entries EntryResolver, dependencies DependencyManager, environment EnvironmentConfiguration, planRefinery PlanRefinery, cacheManager CacheManager, logger scribe.Logger, clock Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		buildpack, err := ParseBuildpack(filepath.Join(context.CNBPath, "buildpack.toml"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Title("%s %s", buildpack.Info.Name, buildpack.Info.Version)
		logger.Process("Resolving Node Engine version")

		entry := entries.Resolve(context.Plan.Entries)

		dependency, err := dependencies.Resolve(buildpack.Metadata.Dependencies, buildpack.Metadata.DefaultVersions.Node, context.Stack, entry)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Selected Node Engine version (using %s): %s", entry.Metadata["version-source"], dependency.Version)
		logger.Break()

		nodeLayer, err := context.Layers.Get(Node, packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		nodeLayer.Build = entry.Metadata["build"] == true
		nodeLayer.Cache = entry.Metadata["cache"] == true

		match, err := cacheManager.Match(nodeLayer, dependency)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if match {
			logger.Process("Reusing cached layer %s", nodeLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   planRefinery.BillOfMaterial(dependency),
				Layers: []packit.Layer{nodeLayer},
			}, nil
		}

		logger.Process("Executing build process")

		err = nodeLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		nodeLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		err = dependencies.Install(dependency, context.CNBPath, nodeLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		config, err := BuildpackYMLParser{}.Parse(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("unable to parse buildpack.yml file: %s", err)
		}

		err = environment.Configure(nodeLayer.SharedEnv, nodeLayer.Path, config.OptimizedMemory)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   planRefinery.BillOfMaterial(dependency),
			Layers: []packit.Layer{nodeLayer},
		}, nil
	}
}
