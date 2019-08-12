package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/node-engine-cnb/nvmrc"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	version := ""
	nvmrcPath := filepath.Join(context.Application.Root, ".nvmrc")
	nvmrcExists, err := helper.FileExists(nvmrcPath)
	if err != nil {
		return context.Fail(), err
	}

	if nvmrcExists {
		version, err = nvmrc.GetVersion(nvmrcPath, context.Logger)
		if err != nil {
			return context.Fail(), err
		}

	}

	buildpackYamlPath := filepath.Join(context.Application.Root, "buildpack.yml")
	buildpackYamlExists, err := helper.FileExists(buildpackYamlPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	if buildpackYamlExists {
		bpYml := &node.BuildpackYAML{}
		err := helper.ReadBuildpackYaml(buildpackYamlPath, bpYml)
		if err != nil {
			return detect.FailStatusCode, err
		}

		if bpYml.Config.Version != "" {
			version = bpYml.Config.Version
		}
	}

	if version != "" {
		return context.Pass(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: node.Dependency}},
			Requires: []buildplan.Required{{
				Name:     node.Dependency,
				Version:  version,
				Metadata: buildplan.Metadata{"build": true},
			}},
		})
	}

	return context.Pass(buildplan.Plan{
		Provides: []buildplan.Provided{{Name: node.Dependency}},
	})
}
