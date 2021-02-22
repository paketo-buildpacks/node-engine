package nodeengine

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
}

func Detect(nvmrcParser, buildpackYMLParser, nodeVersionParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement
		version, err := nvmrcParser.ParseVersion(filepath.Join(context.WorkingDir, NvmrcSource))
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

		version, err = buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, BuildpackYMLSource))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Node,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: BuildpackYMLSource,
				},
			})
		}

		version, err = nodeVersionParser.ParseVersion(filepath.Join(context.WorkingDir, NodeVersionSource))
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
