package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/Masterminds/semver"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Dependency = "node"
const VersionSource = "version-source"

type Config struct {
	OptimizeMemory bool   `yaml:"optimize-memory"`
	Version        string `yaml:"version"`
}

type BuildpackYAML struct {
	Config Config `yaml:"nodejs"`
}

type Contributor struct {
	BuildpackYAML      BuildpackYAML
	buildContribution  bool
	launchContribution bool
	layer              layers.DependencyLayer
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	buildpackYAML, err := LoadBuildpackYAML(context.Application.Root)
	if err != nil {
		return Contributor{}, false, err
	}

	plan, wantDependency, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil {
		return Contributor{}, false, err
	}

	if !wantDependency {
		return Contributor{}, false, nil
	}

	dep, err := context.Buildpack.RuntimeDependency(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{layer: context.Layers.DependencyLayer(dep), BuildpackYAML: buildpackYAML}
	contributor.buildContribution, _ = plan.Metadata["build"].(bool)
	contributor.launchContribution, _ = plan.Metadata["launch"].(bool)

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NODE_HOME", layer.Root); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NODE_ENV", "production"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NODE_MODULES_CACHE", "true"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NODE_VERBOSE", "false"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NPM_CONFIG_PRODUCTION", "true"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("NPM_CONFIG_LOGLEVEL", "error"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("WEB_MEMORY", "512"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("WEB_CONCURRENCY", "1"); err != nil {
			return err
		}

		if err := layer.WriteProfile("0_memory_available.sh", memoryAvailable()); err != nil {
			return err
		}

		if c.BuildpackYAML.Config.OptimizeMemory || os.Getenv("OPTIMIZE_MEMORY") == "true" {
			if err := layer.WriteProfile("1_optimize_memory.sh", `export NODE_OPTIONS="--max_old_space_size=$(( $MEMORY_AVAILABLE * 75 / 100 ))"`); err != nil {
				return err
			}
		}

		return nil
	}, c.flags()...)
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}

func (c Contributor) GetLayer() layers.DependencyLayer {
	return c.layer
}

func LoadBuildpackYAML(appRoot string) (BuildpackYAML, error) {
	var err error
	buildpackYAML := BuildpackYAML{}
	bpYamlPath := filepath.Join(appRoot, "buildpack.yml")

	if exists, err := helper.FileExists(bpYamlPath); err != nil {
		return BuildpackYAML{}, err
	} else if exists {
		err = helper.ReadBuildpackYaml(bpYamlPath, &buildpackYAML)
	}
	return buildpackYAML, err
}

func PriorityPlanMerge(a, b buildpackplan.Plan) (buildpackplan.Plan, error) {
	aVersion := a.Version
	bVersion := b.Version
	aSource := a.Metadata[VersionSource]
	bSource := b.Metadata[VersionSource]

	if aVersion == "" && bVersion == "" {
		return mergePlans(a, b, "", nil)
	} else if aVersion == "" {
		return mergePlans(a, b, bVersion, bSource)
	} else if bVersion == "" {
		return mergePlans(a, b, aVersion, aSource)
	}

	aPriority := getPriority(aSource)
	bPriority := getPriority(bSource)
	if aPriority > bPriority {
		return mergePlans(a, b, aVersion, aSource)
	} else if aPriority == bPriority {
		version, err := getHighestVersion(aVersion, bVersion)
		if err != nil {
			return buildpackplan.Plan{}, fmt.Errorf("failed to get the highest version between %s and %s: %v", aVersion, bVersion, err)
		}
		return mergePlans(a, b, version, aSource)
	} else {
		return mergePlans(a, b, bVersion, bSource)
	}
}

func getHighestVersion(aVersion, bVersion string) (string, error) {
	aSemver, err := semver.NewVersion(aVersion)
	if err != nil {
		return "", fmt.Errorf("failed to convert version %s to semver", aVersion)
	}
	bSemver, err := semver.NewVersion(bVersion)
	if err != nil {
		return "", fmt.Errorf("failed to convert version %s to semver", bVersion)
	}
	version := aVersion
	if aSemver.LessThan(bSemver) {
		version = bVersion
	}

	return version, nil
}

func getPriority(versionSource interface{}) int {
	priorities := map[interface{}]int{
		"buildpack.yml": 3,
		"package.json":  2,
		".nvmrc":        1,
		"":              -1,
	}
	val, ok := priorities[versionSource]

	// Any source is higher than empty string
	if !ok {
		val = 0
	}
	return val
}

func mergePlans(a, b buildpackplan.Plan, version string, versionSource interface{}) (buildpackplan.Plan, error) {
	aBuildVal, err := getBooleanVal(a.Metadata["build"])
	if err != nil {
		return buildpackplan.Plan{}, errors.Wrapf(err, "could not determine 'build' metadata of %s", a.Name)
	}

	bBuildVal, err := getBooleanVal(b.Metadata["build"])
	if err != nil {
		return buildpackplan.Plan{}, errors.Wrapf(err, "could not determine 'build' metadata of %s", b.Name)
	}

	aLaunchVal, err := getBooleanVal(a.Metadata["launch"])
	if err != nil {
		return buildpackplan.Plan{}, errors.Wrapf(err, "could not determine 'launch' metadata of %s", a.Name)
	}

	bLaunchVal, err := getBooleanVal(b.Metadata["launch"])
	if err != nil {
		return buildpackplan.Plan{}, errors.Wrapf(err, "could not determine 'launch' metadata of %s", b.Name)
	}

	metadata := a.Metadata // NOTE: Mutating metadata also mutates a.Metadata
	for key, val := range b.Metadata {
		ignoreKeys := []string{VersionSource, "build", "launch"}
		if !contains(ignoreKeys, key) && val != "" {
			if aVal, ok := metadata[key]; ok && aVal != "" && aVal != val {
				val = aVal.(string) + "," + val.(string)
			}
			metadata[key] = val
		}
	}

	if versionSource != nil && versionSource != "" {
		metadata[VersionSource] = versionSource
	}

	if aBuildVal || bBuildVal {
		metadata["build"] = true
	}

	if aLaunchVal || bLaunchVal {
		metadata["launch"] = true
	}

	return buildpackplan.Plan{
		Name:     a.Name,
		Version:  version,
		Metadata: metadata,
	}, nil
}

func contains(slice []string, val string) bool {
	for _, x := range slice {
		if x == val {
			return true
		}
	}

	return false
}

func getBooleanVal(val interface{}) (bool, error) {
	if val == nil || val == "" {
		return false, nil
	}

	if b, isString := val.(string); isString {
		return b == "true", nil
	} else if b, isBool := val.(bool); isBool {
		return b, nil
	}

	return false, fmt.Errorf("could not get boolean value of %v", val)
}

func memoryAvailable() string {
	return `which jq
if [[ $? -eq 0 ]]; then
	MEMORY_AVAILABLE="$(echo $VCAP_APPLICATION | jq .limits.mem)"
fi

if [[ -z "$MEMORY_AVAILABLE" ]]; then
	memory_in_bytes="$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)"
	MEMORY_AVAILABLE="$(( $memory_in_bytes / ( 1024 * 1024 ) ))"
fi
export MEMORY_AVAILABLE
`
}
