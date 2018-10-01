package build

import (
	"fmt"
	libbuildpackV3 "github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack"
)

const NodeDependency = "node"

func CreateLaunchMetadata() libbuildpackV3.LaunchMetadata {
	return libbuildpackV3.LaunchMetadata{
		Processes: libbuildpackV3.Processes{
			libbuildpackV3.Process{
				Type:    "web",
				Command: "npm start",
			},
		},
	}
}

type Node struct {
	buildContribution, launchContribution bool
	cacheLayer                            libjavabuildpack.DependencyCacheLayer
	launchLayer                           libjavabuildpack.DependencyLaunchLayer
}

func NewNode(builder libjavabuildpack.Build) (Node, bool, error) {
	bp, ok := builder.BuildPlan[NodeDependency]
	if !ok {
		return Node{}, false, nil
	}

	deps, err := builder.Buildpack.Dependencies()
	if err != nil {
		return Node{}, false, err
	}

	dep, err := deps.Best(NodeDependency, bp.Version, builder.Stack)
	if err != nil {
		return Node{}, false, err
	}

	node := Node{}

	if _, ok := bp.Metadata["build"]; ok {
		node.buildContribution = true
		node.cacheLayer = builder.Cache.DependencyLayer(dep)
	}

	if _, ok := bp.Metadata["launch"]; ok {
		node.launchContribution = true
		node.launchLayer = builder.Launch.DependencyLayer(dep)
	}

	return node, true, nil
}

var environment = map[string]string{
	"NODE_ENV":              "production",
	"NODE_MODULES_CACHE":    "true",
	"NODE_VERBOSE":          "false",
	"NPM_CONFIG_PRODUCTION": "true",
	"NPM_CONFIG_LOGLEVEL":   "error",
	"WEB_MEMORY":            "512",
	"WEB_CONCURRENCY":       "1",
}

func (n Node) Contribute() error {
	if n.buildContribution {
		return n.cacheLayer.Contribute(func(artifact string, layer libjavabuildpack.DependencyCacheLayer) error {
			layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
			if err := libjavabuildpack.ExtractTarGz(artifact, layer.Root, 1); err != nil {
				return err
			}

			layer.Logger.SubsequentLine("Writing NODE_HOME")
			layer.OverrideEnv("NODE_HOME", layer.Root)

			for key, value := range environment {
				layer.Logger.SubsequentLine("Writing " + key)
				layer.OverrideEnv(key, value)
			}

			return nil
		})
	}

	if n.launchContribution {
		return n.launchLayer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
			layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
			if err := libjavabuildpack.ExtractTarGz(artifact, layer.Root, 1); err != nil {
				return err
			}

			layer.Logger.SubsequentLine("Writing profile.d/NODE_HOME")
			layer.WriteProfile("NODE_HOME", fmt.Sprintf("export NODE_HOME=%s", layer.Root))

			for key, value := range environment {
				layer.Logger.SubsequentLine("Writing profile.d/" + key)
				layer.WriteProfile(key, fmt.Sprintf("export %s=%s", key, value))
			}

			return nil
		})
	}
	return nil
}
