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
					Name:    "node",
					Version: "package-json-version",
					Metadata: map[string]interface{}{
						"version-source": "package.json",
					},
				},
				{
					Name:    "node",
					Version: "other-version",
				},
				{
					Name:    "node",
					Version: "buildpack-yml-version",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
					},
				},
				{
					Name:    "node",
					Version: "nvmrc-version",
					Metadata: map[string]interface{}{
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "node",
				Version: "buildpack-yml-version",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      package.json  -> \"package-json-version\""))
			Expect(buffer.String()).To(ContainSubstring("      .nvmrc        -> \"nvmrc-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
		})
	})

	context("when a package.json entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "node",
					Version: "package-json-version",
					Metadata: map[string]interface{}{
						"version-source": "package.json",
					},
				},
				{
					Name:    "node",
					Version: "other-version",
				},
				{
					Name:    "node",
					Version: "nvmrc-version",
					Metadata: map[string]interface{}{
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "node",
				Version: "package-json-version",
				Metadata: map[string]interface{}{
					"version-source": "package.json",
				},
			}))
		})
	})

	context("when a .nvmrc entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name:    "node",
					Version: "other-version",
				},
				{
					Name:    "node",
					Version: "nvmrc-version",
					Metadata: map[string]interface{}{
						"version-source": ".nvmrc",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "node",
				Version: "nvmrc-version",
				Metadata: map[string]interface{}{
					"version-source": ".nvmrc",
				},
			}))
		})
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name:    "node",
						Version: "package-json-version",
						Metadata: map[string]interface{}{
							"version-source": "package.json",
						},
					},
					{
						Name:    "node",
						Version: "nvmrc-version",
						Metadata: map[string]interface{}{
							"version-source": ".nvmrc",
							"build":          true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name:    "node",
					Version: "package-json-version",
					Metadata: map[string]interface{}{
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
					Name:    "node",
					Version: "other-version",
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:     "node",
				Version:  "other-version",
				Metadata: map[string]interface{}{},
			}))
		})
	})
}
