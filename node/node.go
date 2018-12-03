package node

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const NodeDependency = "node"

type NodeContributor struct {
	buildContribution  bool
	launchContribution bool
	layer              layers.DependencyLayer
}

func NewNodeContributor(builder build.Build) (NodeContributor, bool, error) {
	plan, wantDependency := builder.BuildPlan[NodeDependency]
	if !wantDependency {
		return NodeContributor{}, false, nil
	}

	deps, err := builder.Buildpack.Dependencies()
	if err != nil {
		return NodeContributor{}, false, err
	}

	dep, err := deps.Best(NodeDependency, plan.Version, builder.Stack)
	if err != nil {
		return NodeContributor{}, false, err
	}

	contributor := NodeContributor{layer: builder.Layers.DependencyLayer(dep)}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

func (n NodeContributor) Contribute() error {
	return n.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := layers.ExtractTarGz(artifact, layer.Root, 1); err != nil {
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

func (n NodeContributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
