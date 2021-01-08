package nodeengine_test

import (
	"bytes"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter nodeengine.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = nodeengine.NewLogEmitter(buffer)
	})

	context("Environment", func() {
		it("prints details about the environment", func() {
			emitter.Environment(packit.Environment{
				"NODE_HOME.override":    "/some/path",
				"NODE_ENV.override":     "production",
				"NODE_VERBOSE.override": "false",
			}, true)

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    NODE_ENV     -> \"production\""))
			Expect(buffer.String()).To(ContainSubstring("    NODE_HOME    -> \"/some/path\""))
			Expect(buffer.String()).To(ContainSubstring("    NODE_VERBOSE -> \"false\""))
			Expect(buffer.String()).To(ContainSubstring("    Writing profile.d/0_memory_available.sh"))
			Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
			Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
			Expect(buffer.String()).To(ContainSubstring("    Writing profile.d/1_optimize_memory.sh"))
			Expect(buffer.String()).To(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
			Expect(buffer.String()).To(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
		})

		context("when not optimizing memory", func() {
			it("omits those details", func() {
				emitter.Environment(packit.Environment{
					"NODE_HOME.override":    "/some/path",
					"NODE_ENV.override":     "production",
					"NODE_VERBOSE.override": "false",
				}, false)

				Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
				Expect(buffer.String()).To(ContainSubstring("    Writing profile.d/0_memory_available.sh"))
				Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
				Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
				Expect(buffer.String()).NotTo(ContainSubstring("    Writing profile.d/1_optimize_memory.sh"))
				Expect(buffer.String()).NotTo(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
				Expect(buffer.String()).NotTo(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
			})
		})
	})

	context("Candidates", func() {
		it("prints a formatted map of version source inputs", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
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
						"version":        "nvmrc-version",
						"version-source": ".nvmrc",
					},
				},
				{
					Name: "node",
				},
			})

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      package.json  -> \"package-json-version\""))
			Expect(buffer.String()).To(ContainSubstring("      .nvmrc        -> \"nvmrc-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"*\""))
		})
	})
}
