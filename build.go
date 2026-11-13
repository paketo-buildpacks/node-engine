package nodeengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
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

// IsNpmUpgradeReusable reports whether a cached layer was built with the same
// pinned npm. The Node.js checksum alone does not change when the pinned npm
// does, so it is checked separately.
func IsNpmUpgradeReusable(nodeLayer packit.Layer, npmChecksum string, logger scribe.Emitter) bool {
	cachedChecksum, _ := nodeLayer.Metadata[NpmDepKey].(string)
	reusable := npmChecksum == cachedChecksum

	logger.Debug.Subprocess("Checksum of pinned npm: %s", npmChecksum)
	logger.Debug.Subprocess("Checksum of npm in layer: %s", cachedChecksum)
	logger.Debug.Subprocess("Npm upgrade matches: %v", reusable)

	return reusable
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

			nodeLayer.Launch, nodeLayer.Build, nodeLayer.Cache = true, true, false
			nodeLayer.Metadata = map[string]interface{}{
				BuildKey:  true,
				LaunchKey: true,
			}

			nodeLayer.SharedEnv.Default("NODE_HOME", "")
		} else {
			logger.Candidates(allEntries)

			version, _ := entry.Metadata["version"].(string)
			dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.SelectedDependency(entry, dependency, clock.Now())

			npmDependency, err := resolvePinnedNpm(dependencyManager, dependency.Version, context.CNBPath, context.Stack)
			if err != nil {
				return packit.BuildResult{}, err
			}

			var npmChecksum string
			if npmDependency != nil {
				npmChecksum = npmDependency.Checksum
			}

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

			// Adds the pinned npm to the BOM when it replaced the bundled npm,
			// in this build or in the reused layer.
			addNpmToBOM := func() {
				if sbomDisabled {
					return
				}
				legacySBOM = append(legacySBOM, dependencyManager.GenerateBillOfMaterials(*npmDependency)...)
				if build {
					buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
				}
				if launch {
					launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
				}
			}

			if IsLayerReusable(nodeLayer, dependency.Checksum, build, launch, logger) &&
				IsNpmUpgradeReusable(nodeLayer, npmChecksum, logger) {
				logger.Process("Reusing cached layer %s", nodeLayer.Path)
				logger.Break()

				if upgraded, _ := nodeLayer.Metadata[NpmUpgradedKey].(bool); upgraded && npmDependency != nil {
					addNpmToBOM()
				}

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

			// Only recorded when the upgrade applies to this Node.js version.
			if npmChecksum != "" {
				nodeLayer.Metadata[NpmDepKey] = npmChecksum
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

			// Runs after Deliver, which unpacks the bundled npm. The layer SBOM
			// below does not list npm, so the pinned npm is added to the BOM.
			if npmDependency != nil {
				upgraded, err := upgradeBundledNpm(dependencyManager, logger, clock, *npmDependency, dependency.Version, context.CNBPath, nodeLayer.Path, context.Platform.Path)
				if err != nil {
					return packit.BuildResult{}, err
				}

				if upgraded {
					nodeLayer.Metadata[NpmUpgradedKey] = true
					addNpmToBOM()
				}
			}

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
			nodeLayer.SharedEnv.Default("NODE_HOME", nodeLayer.Path)
		}

		var optimizedMemory bool
		if os.Getenv("BP_NODE_OPTIMIZE_MEMORY") == "true" {
			optimizedMemory = true
		}

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

// For the Node.js release line below, the npm bundled in the Node.js
// distribution is upgraded to the "npm" dependency pinned in buildpack.toml,
// so that vulnerabilities fixed in newer npm releases can be picked up before
// a new Node.js release bundles them.
const (
	NpmUpgradeNodeVersions = "24.x"

	NpmDepKey      = "npm-dependency-sha"
	NpmUpgradedKey = "npm-upgraded"
)

// resolvePinnedNpm returns the pinned npm dependency, or nil if the upgrade
// does not apply to the given Node.js version.
func resolvePinnedNpm(dependencyManager DependencyManager, nodeVersion, cnbPath, stack string) (*postal.Dependency, error) {
	// A version that cannot be parsed is left as is.
	applies, err := versionSatisfies(nodeVersion, NpmUpgradeNodeVersions)
	if err != nil || !applies {
		return nil, nil
	}

	dependency, err := dependencyManager.Resolve(filepath.Join(cnbPath, "buildpack.toml"), Npm, "", stack)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve pinned npm for Node.js %s: %w", nodeVersion, err)
	}

	return &dependency, nil
}

// upgradeBundledNpm replaces the npm bundled in the Node.js layer with the
// pinned npm, if it is newer, has the same major version and supports the
// Node.js version. It reports whether npm was replaced. The new npm is
// unpacked into a staging directory first, so a failure keeps the bundled npm.
func upgradeBundledNpm(dependencyManager DependencyManager, logger scribe.Emitter, clock chronos.Clock, npmDependency postal.Dependency, nodeVersion, cnbPath, layerPath, platformPath string) (bool, error) {
	modulesDir := filepath.Join(layerPath, "lib", "node_modules")
	pkgDir := filepath.Join(modulesDir, "npm")

	logger.Process("Checking npm bundled in the Node.js distribution")

	bundled, err := readNpmPackage(pkgDir)
	if errors.Is(err, os.ErrNotExist) {
		logger.Subprocess("No bundled npm found, skipping upgrade")
		logger.Break()
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read bundled npm version: %w", err)
	}

	bundledVersion, err := semver.NewVersion(bundled.Version)
	if err != nil {
		return false, fmt.Errorf("failed to parse bundled npm version %q: %w", bundled.Version, err)
	}

	pinnedVersion, err := semver.NewVersion(npmDependency.Version)
	if err != nil {
		return false, fmt.Errorf("failed to parse pinned npm version %q: %w", npmDependency.Version, err)
	}

	switch {
	case bundledVersion.Major() != pinnedVersion.Major():
		logger.Subprocess("Keeping bundled npm %s (pinned npm %s is a different major version)", bundledVersion, pinnedVersion)
		logger.Break()
		return false, nil
	case !bundledVersion.LessThan(pinnedVersion):
		logger.Subprocess("Keeping bundled npm %s (not older than pinned npm %s)", bundledVersion, pinnedVersion)
		logger.Break()
		return false, nil
	}

	stagingDir, err := os.MkdirTemp(modulesDir, ".npm-upgrade-")
	if err != nil {
		return false, fmt.Errorf("failed to create npm staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	// MkdirTemp uses mode 0700; the app's run user must be able to read npm.
	if err := os.Chmod(stagingDir, 0o755); err != nil {
		return false, fmt.Errorf("failed to set permissions on npm staging directory: %w", err)
	}

	logger.Subprocess("Installing npm %s", pinnedVersion)
	duration, err := clock.Measure(func() error {
		return dependencyManager.Deliver(npmDependency, cnbPath, stagingDir, platformPath)
	})
	if err != nil {
		return false, err
	}
	logger.Action("Completed in %s", duration.Round(time.Millisecond))

	pinned, err := readNpmPackage(stagingDir)
	if err != nil {
		return false, fmt.Errorf("failed to read pinned npm package: %w", err)
	}

	if pinned.Engines.Node != "" {
		supported, err := versionSatisfies(nodeVersion, pinned.Engines.Node)
		if err != nil {
			return false, fmt.Errorf("failed to check npm %s engines: %w", pinnedVersion, err)
		}
		if !supported {
			logger.Action("WARNING: npm %s requires Node.js %q, which does not include %s", pinnedVersion, pinned.Engines.Node, nodeVersion)
			logger.Action("Keeping bundled npm %s", bundledVersion)
			logger.Break()
			return false, nil
		}
	}

	if err := os.RemoveAll(pkgDir); err != nil {
		return false, fmt.Errorf("failed to remove bundled npm: %w", err)
	}
	if err := os.Rename(stagingDir, pkgDir); err != nil {
		return false, fmt.Errorf("failed to install pinned npm: %w", err)
	}

	logger.Action("Upgraded bundled npm %s -> %s", bundledVersion, pinnedVersion)
	logger.Break()

	return true, nil
}

type npmPackage struct {
	Version string `json:"version"`
	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`
}

func readNpmPackage(pkgDir string) (npmPackage, error) {
	var pkg npmPackage

	content, err := os.ReadFile(filepath.Join(pkgDir, "package.json"))
	if err != nil {
		return pkg, err
	}

	err = json.Unmarshal(content, &pkg)
	return pkg, err
}

func versionSatisfies(version, constraint string) (bool, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}
