package nodeengine_test

import (
	"bytes"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer   *bytes.Buffer
		resolver nodeengine.PlanEntryResolver
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		resolver = nodeengine.NewPlanEntryResolver(nodeengine.NewLogEmitter(buffer))
	})

	context("when a buildpack.yml entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "package-json-version",
						"version-source": "package.json",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "buildpack-yml-version",
						"version-source": "buildpack.yml",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "node-version-version",
						"version-source": ".node-version",
					},
				},
				{
					Name: "npm",
					Metadata: map[string]interface{}{
						"version":        "npm-version",
						"version-source": ".npmrc",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "nvmrc-version",
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "buildpack-yml-version",
					"version-source": "buildpack.yml",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      package.json  -> \"package-json-version\""))
			Expect(buffer.String()).To(ContainSubstring("      .nvmrc        -> \"nvmrc-version\""))
			Expect(buffer.String()).To(ContainSubstring("      .node-version -> \"node-version-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
			Expect(buffer.String()).NotTo(ContainSubstring("      .npmrc        -> \"npm-version\""))
		})
	})

	context("when a package.json entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "package-json-version",
						"version-source": "package.json",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "nvmrc-version",
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "package-json-version",
					"version-source": "package.json",
				},
			}))
		})
	})

	context("when a .nvmrc entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "nvmrc-version",
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "nvmrc-version",
					"version-source": ".nvmrc",
				},
			}))
		})
	})

	context("when a .node-version entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "node-version-version",
						"version-source": ".node-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "node-version-version",
					"version-source": ".node-version",
				},
			}))
		})
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "package-json-version",
							"version-source": "package.json",
						},
					},
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "nvmrc-version",
							"version-source": ".nvmrc",
							"build":          true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "package-json-version",
						"version-source": "package.json",
						"build":          true,
					},
				}))
			})
		})
	})

	context("when an unknown source entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version": "other-version",
				},
			}))
		})
	})
}
