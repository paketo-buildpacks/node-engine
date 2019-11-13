package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOptimizeMemory(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("sets max_old_space_size when nodejs.optimize-memory is set in buildpack.yml", func() {
		app, err := dagger.NewPack(
			filepath.Join("testdata", "optimize_memory"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(nodeBuildpack),
		).Build()
		Expect(err).ToNot(HaveOccurred())
		app.Memory = "128m"
		defer app.Destroy()

		Expect(app.StartWithCommand("node server.js")).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("NodeOptions: --max_old_space_size=96"))

		logs := GetBuildLogs(app.BuildLogs())
		Expect(logs).To(ContainSequence([]string{
			"      Writing profile.d/1_optimize_memory.sh",
			"        Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.",
			"        Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE.",
		}))
	})
}
