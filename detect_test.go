package nodeengine_test

import (
	"errors"
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
		detect             packit.DetectFunc
	)

	it.Before(func() {
		nvmrcParser = &fakes.VersionParser{}
		buildpackYMLParser = &fakes.VersionParser{}

		detect = nodeengine.Detect(nvmrcParser, buildpackYMLParser)
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

	context("when the source code contains a both .nvmrc and buildpack.yml files", func() {
		it.Before(func() {
			nvmrcParser.ParseVersionCall.Returns.Version = "1.2.3"
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
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("failure cases", func() {
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
	})
}
