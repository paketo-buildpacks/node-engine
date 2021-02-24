package nodeengine_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/node-engine/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		cnbDir            string
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		clock             chronos.Clock
		timeStamp         time.Time
		environment       *fakes.EnvironmentConfiguration
		planRefinery      *fakes.BuildPlanRefinery
		buffer            *bytes.Buffer

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
  id = "org.some-org.some-buildpack"
  name = "Some Buildpack"
  version = "some-version"

[metadata]
  [metadata.default-versions]
    node = "10.x"

  [[metadata.dependencies]]
    deprecation_date = 2021-04-01T00:00:00Z
    id = "some-dep"
    name = "Some Dep"
    sha256 = "some-sha"
    stacks = ["some-stack"]
    uri = "some-uri"
    version = "some-dep-version"
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "node",
			Metadata: map[string]interface{}{
				"version":        "~10",
				"version-source": "BP_NODE_VERSION",
			},
		}
		entryResolver.MergeLayerTypesCall.Returns.Launch = false
		entryResolver.MergeLayerTypesCall.Returns.Build = false

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "Node Engine"}

		environment = &fakes.EnvironmentConfiguration{}
		planRefinery = &fakes.BuildPlanRefinery{}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"name":    "node-dependency-name",
						"sha256":  "node-dependency-sha",
						"stacks":  []string{"some-stack"},
						"uri":     "node-dependency-uri",
						"version": "~10",
					},
				},
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := nodeengine.NewLogEmitter(buffer)

		build = nodeengine.Build(entryResolver, dependencyManager, environment, planRefinery, logEmitter, clock)

		buildContext = packit.BuildContext{
			CNBPath: cnbDir,
			Stack:   "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "1.2.3",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "~10",
							"version-source": "BP_NODE_VERSION",
						},
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		}

	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs node", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"name":    "node-dependency-name",
							"sha256":  "node-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "node-dependency-uri",
							"version": "~10",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name:      "node",
					Path:      filepath.Join(layersDir, "node"),
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    false,
					Cache:     false,
					Metadata: map[string]interface{}{
						nodeengine.DepKey: "",
						"built_at":        timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(filepath.Join(layersDir, "node")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "~10",
					"version-source": "BP_NODE_VERSION",
				},
			},
		}))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Name).To(Equal("node"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Entries).To(Equal(
			[]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "~10",
						"version-source": "BP_NODE_VERSION",
					},
				},
			}))
		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("node"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("~10"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(planRefinery.BillOfMaterialCall.CallCount).To(Equal(1))
		Expect(planRefinery.BillOfMaterialCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine"}))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine"}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "node")))

		Expect(environment.ConfigureCall.Receives.BuildEnv).To(Equal(packit.Environment{}))
		Expect(environment.ConfigureCall.Receives.LaunchEnv).To(Equal(packit.Environment{}))
		Expect(environment.ConfigureCall.Receives.Path).To(Equal(filepath.Join(layersDir, "node")))
		Expect(environment.ConfigureCall.Receives.OptimizeMemory).To(BeFalse())

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
		Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using BP_NODE_VERSION): "))
		Expect(buffer.String()).ToNot(ContainSubstring("WARNING: Setting the Node version through buildpack.yml will be deprecated soon in Node Engine Buildpack v2.0.0."))
		Expect(buffer.String()).ToNot(ContainSubstring("Please specify the version through the $BP_NODE_VERSION environment variable instead. See README.md for more information."))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
	})

	context("when the buildpack.yml contains a directive to optimize memory", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, "buildpack.yml"), []byte(`---
nodejs:
  optimize-memory: true`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("tells the environment to optimize memory", func() {
			buildContext.WorkingDir = workingDir
			_, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(environment.ConfigureCall.Receives.BuildEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.LaunchEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.Path).To(Equal(filepath.Join(layersDir, "node")))
			Expect(environment.ConfigureCall.Receives.OptimizeMemory).To(BeTrue())
		})
	})

	context("when the os environment contains a directive to optimize memory", func() {
		it.Before(func() {
			Expect(os.Setenv("OPTIMIZE_MEMORY", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("OPTIMIZE_MEMORY")).To(Succeed())
		})

		it("tells the environment to optimize memory", func() {
			_, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(environment.ConfigureCall.Receives.BuildEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.LaunchEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.Path).To(Equal(filepath.Join(layersDir, "node")))
			Expect(environment.ConfigureCall.Receives.OptimizeMemory).To(BeTrue())
		})
	})

	context("when the build plan entry includes the build, launch flags", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "~10",
					"version-source": "BP_NODE_VERSION",
					"launch":         true,
					"build":          true,
				},
			}

			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true

			planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"name":    "node-dependency-name",
							"sha256":  "node-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "node-dependency-uri",
							"version": "~10",
						},
					},
				},
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the node layer as build, cached and launch", func() {
			buildContext.Plan.Entries[0].Metadata["launch"] = true
			buildContext.Plan.Entries[0].Metadata["build"] = true
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"name":    "node-dependency-name",
								"sha256":  "node-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "node-dependency-uri",
								"version": "~10",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "node",
						Path:      filepath.Join(layersDir, "node"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     true,
						Launch:    true,
						Cache:     true,
						Metadata: map[string]interface{}{
							nodeengine.DepKey: "",
							"built_at":        timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
		})
	})

	context("when we refine the buildpack plan", func() {
		it.Before(func() {
			planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "new-dep",
						Metadata: map[string]interface{}{
							"version":          "some-version",
							"some-extra-field": "an-extra-value",
						},
					},
				},
			}
		})
		it("refines the BuildpackPlan", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "new-dep",
							Metadata: map[string]interface{}{
								"version":          "some-version",
								"some-extra-field": "an-extra-value",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "node",
						Path:      filepath.Join(layersDir, "node"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     false,
						Launch:    false,
						Cache:     false,
						Metadata: map[string]interface{}{
							nodeengine.DepKey: "",
							"built_at":        timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))

		})
	})

	context("when there is a dependency cache match", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(layersDir, "node.toml"), []byte("[metadata]\ndependency-sha = \"some-sha\"\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				Name:   "Node Engine",
				SHA256: "some-sha",
			}
		})

		it("exits build process early", func() {
			_, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(planRefinery.BillOfMaterialCall.CallCount).To(Equal(1))
			Expect(planRefinery.BillOfMaterialCall.Receives.Dependency).To(Equal(postal.Dependency{
				Name:   "Node Engine",
				SHA256: "some-sha",
			}))

			Expect(dependencyManager.InstallCall.CallCount).To(Equal(0))
			Expect(environment.ConfigureCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using BP_NODE_VERSION): "))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
		})
	})

	context("when the entry version source is buildpack.yml", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "~10",
					"version-source": "buildpack.yml",
				},
			}
		})

		it("returns result that installs version in buildpack.yml and provides deprecation warning", func() {
			buildContext.Plan.Entries[0].Metadata["version-source"] = "buildpack.yml"
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "node",
							Metadata: map[string]interface{}{
								"name":    "node-dependency-name",
								"sha256":  "node-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "node-dependency-uri",
								"version": "~10",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "node",
						Path:      filepath.Join(layersDir, "node"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     false,
						Launch:    false,
						Cache:     false,
						Metadata: map[string]interface{}{
							nodeengine.DepKey: "",
							"built_at":        timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))

			Expect(filepath.Join(layersDir, "node")).To(BeADirectory())

			Expect(entryResolver.ResolveCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "~10",
						"version-source": "buildpack.yml",
					},
				},
			}))

			Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
			Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("node"))
			Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("~10"))
			Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(planRefinery.BillOfMaterialCall.CallCount).To(Equal(1))
			Expect(planRefinery.BillOfMaterialCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine"}))

			Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine"}))
			Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
			Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "node")))

			Expect(environment.ConfigureCall.Receives.BuildEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.LaunchEnv).To(Equal(packit.Environment{}))
			Expect(environment.ConfigureCall.Receives.Path).To(Equal(filepath.Join(layersDir, "node")))
			Expect(environment.ConfigureCall.Receives.OptimizeMemory).To(BeFalse())

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using buildpack.yml): "))
			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the Node version through buildpack.yml will be deprecated soon in Node Engine Buildpack v2.0.0."))
			Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_NODE_VERSION environment variable instead. See README.md for more information."))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		})
	})

	context("failure cases", func() {
		context("when a dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when a dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})

		context("when the layers directory cannot be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the layer directory cannot be removed", func() {
			var layerDir string
			it.Before(func() {
				layerDir = filepath.Join(layersDir, nodeengine.Node)
				Expect(os.MkdirAll(filepath.Join(layerDir, "baller"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(layerDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())
				Expect(os.RemoveAll(layerDir)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the environment cannot be configured", func() {
			it.Before(func() {
				environment.ConfigureCall.Returns.Error = errors.New("failed to configure environment")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to configure environment"))
			})
		})
	})
}
