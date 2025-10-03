package nodeengine

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
	Build         bool   `toml:"build"`
	Launch        bool   `toml:"launch"`
}

func Detect(nvmrcParser, nodeVersionParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		projectPath := context.WorkingDir
		customProjPath := os.Getenv("BP_NODE_PROJECT_PATH")

		if customProjPath != "" {
			customProjPath = filepath.Clean(customProjPath)
			projectPath = filepath.Join(projectPath, customProjPath)
			exists, err := fs.Exists(projectPath)
			if err != nil {
				return packit.DetectResult{}, err
			}

			if !exists {
				return packit.DetectResult{},
					packit.Fail.WithMessage("expected value derived from BP_NODE_PROJECT_PATH [%s] to be an existing directory", projectPath)
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

		targetOs := os.Getenv("CNB_TARGET_DISTRO_NAME")
		_, pythonNotFound := exec.LookPath("python")

		installPython := (targetOs != "rhel" && pythonNotFound != nil)
		if installPython {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Cpython,
				Metadata: BuildPlanMetadata{
					Build:  true,
					Launch: false,
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
