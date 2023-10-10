package nodeengine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/paketo-buildpacks/libnodejs"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
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

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func IsLayerReusable(nodeLayer packit.Layer, depChecksum string, build bool, launch bool, logger scribe.Emitter) bool {
	logger.Debug.Process("Checking if layer %s can be reused", nodeLayer.Path)

	metadata := nodeLayer.Metadata
	cachedChecksum, _ := metadata[DepKey].(string)
	logger.Debug.Subprocess("Checksum of dependency: %s", depChecksum)
	logger.Debug.Subprocess("Checksum of layer: %s", cachedChecksum)

	cachedBuild, found := metadata[BuildKey].(bool)
	buildOK := found && (build == cachedBuild)
	logger.Debug.Subprocess("Build requirements match: %v", buildOK)

	cachedLaunch, found := metadata[LaunchKey].(bool)
	launchOK := found && (launch == cachedLaunch)
	logger.Debug.Subprocess("Launch requirements match: %v", launchOK)

	logger.Debug.Break()

	return cargo.Checksum(depChecksum).MatchString(cachedChecksum) && buildOK && launchOK
}

func Build(entryResolver EntryResolver, dependencyManager DependencyManager, sbomGenerator SBOMGenerator, logger scribe.Emitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		nodeLayer, err := context.Layers.Get(Node)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Resolving Node Engine version")

		entry, allEntries := libnodejs.ResolveNodeVersion(entryResolver.Resolve, context.Plan)
		if entry.Name == "" && len(allEntries) == 0 {
			logger.Process("Node no longer requested by plan, satisfied by extension")

			logger.Process("Setting up launch layer for environment variables")
			nodeLayer, err = nodeLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}

			nodeLayer.Launch, nodeLayer.Build, nodeLayer.Cache = true, false, false
			nodeLayer.Metadata = map[string]interface{}{
				BuildKey:  false,
				LaunchKey: true,
			}
		} else {
			logger.Candidates(allEntries)

			version, _ := entry.Metadata["version"].(string)
			dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.SelectedDependency(entry, dependency, clock.Now())

			sbomDisabled, err := checkSbomDisabled()
			if err != nil {
				return packit.BuildResult{}, err
			}

			var legacySBOM []packit.BOMEntry
			if !sbomDisabled {
				legacySBOM = dependencyManager.GenerateBillOfMaterials(dependency)
			}

			launch, build := entryResolver.MergeLayerTypes("node", context.Plan.Entries)

			if build {
				buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
			}

			if launch {
				launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
			}

			if IsLayerReusable(nodeLayer, dependency.Checksum, build, launch, logger) {
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
				DepKey:    dependency.Checksum,
				BuildKey:  build,
				LaunchKey: launch,
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

			if sbomDisabled {
				logger.Subprocess("Skipping SBOM generation for Node Engine")
				logger.Break()
			} else {
				logger.GeneratingSBOM(nodeLayer.Path)
				var sbomContent sbom.SBOM
				duration, err = clock.Measure(func() error {
					sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, nodeLayer.Path)
					return err
				})
				if err != nil {
					return packit.BuildResult{}, err
				}

				logger.Action("Completed in %s", duration.Round(time.Millisecond))
				logger.Break()

				logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
				nodeLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
				if err != nil {
					return packit.BuildResult{}, err
				}
			}
		}

		var optimizedMemory bool
		if os.Getenv("BP_NODE_OPTIMIZE_MEMORY") == "true" {
			optimizedMemory = true
		}

		nodeLayer.SharedEnv.Default("NODE_HOME", nodeLayer.Path)
		nodeLayer.SharedEnv.Default("NODE_ENV", "production")
		nodeLayer.SharedEnv.Default("NODE_VERBOSE", "false")
		nodeLayer.SharedEnv.Default("NODE_OPTIONS", "--use-openssl-ca")
		if optimizedMemory {
			nodeLayer.LaunchEnv.Default("OPTIMIZE_MEMORY", "true")
		}

		logger.EnvironmentVariables(nodeLayer)
		nodeLayer.ExecD = []string{
			filepath.Join(context.CNBPath, "bin", "optimize-memory"),
			filepath.Join(context.CNBPath, "bin", "inspector"),
		}

		logger.Subprocess("Writing exec.d/0-optimize-memory")
		logger.Action("Calculates available memory based on container limits at launch time.")
		logger.Action("Made available in the MEMORY_AVAILABLE environment variable.")
		if optimizedMemory {
			logger.Action("Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.")
			logger.Action("Limits the total size of all objects on the heap to 75%% of the MEMORY_AVAILABLE.")
		}
		logger.Subprocess("Writing exec.d/1-inspector")
		logger.Break()

		return packit.BuildResult{
			Layers: []packit.Layer{nodeLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}

func checkSbomDisabled() (bool, error) {
	if disableStr, ok := os.LookupEnv("BP_DISABLE_SBOM"); ok {
		disable, err := strconv.ParseBool(disableStr)
		if err != nil {
			return false, fmt.Errorf("failed to parse BP_DISABLE_SBOM value %s: %w", disableStr, err)
		}
		return disable, nil
	}
	return false, nil
}
