package nodeengine

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
}

func Detect(nvmrcParser, nodeVersionParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		projectPath := context.WorkingDir
		customProjPath := os.Getenv("BP_NODE_PROJECT_PATH")

		if customProjPath != "" {
			customProjPath = filepath.Clean(customProjPath)
			projectPath = filepath.Join(projectPath, customProjPath)
			_, err := os.Stat(projectPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return packit.DetectResult{},
						packit.Fail.WithMessage("expected value derived from BP_NODE_PROJECT_PATH [%s] to be an existing directory", projectPath)
				} else {
					return packit.DetectResult{}, err
				}
			}
		}

		version, err := nvmrcParser.ParseVersion(filepath.Join(projectPath, NvmrcSource))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Node,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: NvmrcSource,
				},
			})
		}

		version = os.Getenv("BP_NODE_VERSION")
		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Node,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "BP_NODE_VERSION",
				},
			})
		}

		version, err = nodeVersionParser.ParseVersion(filepath.Join(projectPath, NodeVersionSource))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Node,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: NodeVersionSource,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Node},
				},
				Requires: requirements,
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: Node},
							{Name: Npm},
						},
						Requires: requirements,
					},
				},
			},
		}, nil
	}
}
