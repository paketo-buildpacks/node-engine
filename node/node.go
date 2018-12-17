package node

import (
	"encoding/json"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"os"
)

const Dependency = "node"

type packageJSON struct {
	Engines engines `json:"engines"`
}

type engines struct {
	Node string `json:"node"`
}

func GetVersion(packageFile string) (version string, err error) {
	file, err := os.Open(packageFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	pkgJSON := packageJSON{}
	if err := json.NewDecoder(file).Decode(&pkgJSON); err != nil {
		return "", err
	}

	return pkgJSON.Engines.Node, nil
}

type Contributor struct {
	buildContribution  bool
	launchContribution bool
	layer              layers.DependencyLayer
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{layer: context.Layers.DependencyLayer(dep)}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

func (n Contributor) Contribute() error {
	return n.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
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

		return nil
	}, n.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
