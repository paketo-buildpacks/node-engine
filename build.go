package nodeengine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2"
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

func Build(entryResolver EntryResolver, dependencyManager DependencyManager, sbomGenerator SBOMGenerator, logger scribe.Emitter, clock chronos.Clock) packit.BuildFunc {
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
		if entry.Name == "" && len(allEntries) == 0 {
			logger.Process("Node no longer requested by plan")
			return packit.BuildResult{}, nil
		} else {
			logger.Process("Resolving Node Engine version")
			logger.Candidates(allEntries)
		}

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

		nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
		versionSource, _ := entry.Metadata["version-source"].(string)
		if versionSource == "buildpack.yml" {
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
			buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
		}

		cachedSHA, ok := nodeLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 { //nolint:staticcheck
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
			DepKey: dependency.SHA256, //nolint:staticcheck
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

		nodeLayer.SharedEnv.Default("NODE_HOME", nodeLayer.Path)
		nodeLayer.SharedEnv.Default("NODE_ENV", "production")
		nodeLayer.SharedEnv.Default("NODE_VERBOSE", "false")
		if config.OptimizedMemory {
			nodeLayer.LaunchEnv.Default("OPTIMIZE_MEMORY", "true")
		}

		// NOTE: ensures OpenSSL CA store works with Node v18 and higher. Waiting
		// for resolution on https://github.com/nodejs/node/issues/43560 to decide
		// how to properly fix this.
		nodeVersion, err := semver.NewVersion(dependency.Version)
		if err != nil {
			return packit.BuildResult{}, err
		}
		if !nodeVersion.LessThan(semver.MustParse("18.0.0")) {
			nodeLayer.SharedEnv.Append("SSL_CERT_DIR", "/etc/ssl/certs", ":")
		}

		logger.EnvironmentVariables(nodeLayer)

		nodeLayer.ExecD = []string{filepath.Join(context.CNBPath, "bin", "optimize-memory")}

		logger.Subprocess("Writing exec.d/0-optimize-memory")
		logger.Action("Calculates available memory based on container limits at launch time.")
		logger.Action("Made available in the MEMORY_AVAILABLE environment variable.")
		if config.OptimizedMemory {
			logger.Action("Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.")
			logger.Action("Limits the total size of all objects on the heap to 75%% of the MEMORY_AVAILABLE.")
		}
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
