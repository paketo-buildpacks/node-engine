package nodeengine_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/node-engine/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		nvmrcParser        *fakes.VersionParser
		buildpackYMLParser *fakes.VersionParser
		nodeVersionParser  *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		nvmrcParser = &fakes.VersionParser{}
		buildpackYMLParser = &fakes.VersionParser{}
		nodeVersionParser = &fakes.VersionParser{}

		detect = nodeengine.Detect(nvmrcParser, buildpackYMLParser, nodeVersionParser)
	})

	it("returns a plan that provides node", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: nodeengine.Node},
			},
			Or: []packit.BuildPlan{
				{
					Provides: []packit.BuildPlanProvision{
						{Name: nodeengine.Node},
						{Name: nodeengine.Npm},
					},
				},
			},
		}))
	})

	context("when the source code contains an .nvmrc file", func() {
		it.Before(func() {
			nvmrcParser.ParseVersionCall.Returns.Version = "1.2.3"
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "1.2.3",
							VersionSource: ".nvmrc",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "1.2.3",
									VersionSource: ".nvmrc",
								},
							},
						},
					},
				},
			}))

			Expect(nvmrcParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/.nvmrc"))
		})
	})

	context("when the source code contains a buildpack.yml file", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.6"
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "4.5.6",
							VersionSource: "buildpack.yml",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "4.5.6",
									VersionSource: "buildpack.yml",
								},
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("when $BP_NODE_VERSION is set", func() {
		it.Before(func() {
			os.Setenv("BP_NODE_VERSION", "4.5.6")
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_VERSION")
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "4.5.6",
							VersionSource: "BP_NODE_VERSION",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "4.5.6",
									VersionSource: "BP_NODE_VERSION",
								},
							},
						},
					},
				},
			}))
		})
	})

	context("when $BP_NODE_VERSION is set and buildpack.yml exists", func() {
		it.Before(func() {
			os.Setenv("BP_NODE_VERSION", "4.5.6")
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.7"
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_VERSION")
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "4.5.6",
							VersionSource: "BP_NODE_VERSION",
						},
					},
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "4.5.7",
							VersionSource: "buildpack.yml",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "4.5.6",
									VersionSource: "BP_NODE_VERSION",
								},
							},
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "4.5.7",
									VersionSource: "buildpack.yml",
								},
							},
						},
					},
				},
			}))
			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("when the source code contains a .node-version file", func() {
		it.Before(func() {
			nodeVersionParser.ParseVersionCall.Returns.Version = "7.8.9"
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "7.8.9",
							VersionSource: ".node-version",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "7.8.9",
									VersionSource: ".node-version",
								},
							},
						},
					},
				},
			}))

			Expect(nodeVersionParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/.node-version"))
		})
	})

	context("when the source code contains .nvmrc, buildpack.yml and .node-version files", func() {
		it.Before(func() {
			nvmrcParser.ParseVersionCall.Returns.Version = "1.2.3"
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.6"
			nodeVersionParser.ParseVersionCall.Returns.Version = "7.8.9"
		})

		it("returns a plan that provides and requires that version of node", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: nodeengine.Node},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "1.2.3",
							VersionSource: ".nvmrc",
						},
					},
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "4.5.6",
							VersionSource: "buildpack.yml",
						},
					},
					{
						Name: nodeengine.Node,
						Metadata: nodeengine.BuildPlanMetadata{
							Version:       "7.8.9",
							VersionSource: ".node-version",
						},
					},
				},
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: nodeengine.Node},
							{Name: nodeengine.Npm},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "1.2.3",
									VersionSource: ".nvmrc",
								},
							},
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "4.5.6",
									VersionSource: "buildpack.yml",
								},
							},
							{
								Name: nodeengine.Node,
								Metadata: nodeengine.BuildPlanMetadata{
									Version:       "7.8.9",
									VersionSource: ".node-version",
								},
							},
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("when $BP_NODE_PROJECT_PATH is set", func() {
		var workingDir string
		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(filepath.Join(workingDir, "custom", "path"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("BP_NODE_PROJECT_PATH", "custom/path")
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_PROJECT_PATH")
			err := os.RemoveAll(workingDir)
			Expect(err).NotTo(HaveOccurred())
		})

		it("it uses the custom path for config file parsers", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(nvmrcParser.ParseVersionCall.Receives.Path).To(Equal(fmt.Sprintf("%s/.nvmrc", filepath.Join(workingDir, "custom", "path"))))
			Expect(nodeVersionParser.ParseVersionCall.Receives.Path).To(Equal(fmt.Sprintf("%s/.node-version", filepath.Join(workingDir, "custom", "path"))))
		})
	})

	context("failure cases", func() {
		context("when the dir specified by BP_NODE_PROJECT_PATH does not exist", func() {
			var workingDir string

			it.Before(func() {
				var err error
				workingDir, err = ioutil.TempDir("", "working-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Setenv("BP_NODE_PROJECT_PATH", "src/does/not/exist")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_NODE_PROJECT_PATH")).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("expected value derived from BP_NODE_PROJECT_PATH [%s] to be an existing directory", filepath.Join(workingDir, "src", "does", "not", "exist")),
						),
					),
				)
			})
		})

		context("when the nvmrc parser fails", func() {
			it.Before(func() {
				nvmrcParser.ParseVersionCall.Returns.Err = errors.New("failed to parse .nvmrc")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse .nvmrc"))
			})
		})

		context("when the nvmrc parser fails", func() {
			it.Before(func() {
				nvmrcParser.ParseVersionCall.Returns.Err = errors.New("failed to parse .nvmrc")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse .nvmrc"))
			})
		})

		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml"))
			})
		})

		context("when the .node-version parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse .node-version")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse .node-version"))
			})
		})
	})
}
