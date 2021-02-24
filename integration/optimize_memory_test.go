package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOptimizeMemory(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		image     occam.Image
		container occam.Container
		name      string
		source    string
	)

	it.Before(func() {
		docker = occam.NewDocker()
		pack = occam.NewPack()

		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	it("sets max_old_space_size when nodejs.optimize-memory is set with env variable BP_NODE_OPTIMIZE_MEMORY", func() {
		var err error
		source, err = occam.Source(filepath.Join("testdata", "optimize_mem_app"))
		Expect(err).NotTo(HaveOccurred())

		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("never").
			WithBuildpacks(
				nodeBuildpack,
				buildPlanBuildpack,
			).
			WithEnv(map[string]string{"BP_NODE_OPTIMIZE_MEMORY": "true"}).
			Execute(name, source)

		Expect(err).NotTo(HaveOccurred())

		container, err = docker.Container.Run.
			WithMemory("128m").
			WithCommand("node server.js").
			WithPublish("8080").
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("NodeOptions: --max_old_space_size=96")).OnPort(8080))

		Expect(logs).To(ContainLines(
			"    Writing profile.d/1_optimize_memory.sh",
			"      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.",
			"      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE.",
		))
	})
}
