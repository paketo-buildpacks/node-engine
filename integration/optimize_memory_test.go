package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/occam/matchers"
	. "github.com/onsi/gomega"
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
	})

	it("sets max_old_space_size when nodejs.optimize-memory is set in buildpack.yml", func() {
		var err error
		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithNoPull().
			WithBuildpacks(nodeBuildpack).
			Execute(name, filepath.Join("testdata", "optimize_memory"))

		Expect(err).NotTo(HaveOccurred())

		container, err = docker.Container.Run.WithMemory("128m").WithCommand("node server.js").Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())

		response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort()))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		content, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(ContainSubstring("NodeOptions: --max_old_space_size=96"))

		Expect(GetBuildLogs(logs.String())).To(ContainSequence([]string{
			"    Writing profile.d/1_optimize_memory.sh",
			"      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.",
			"      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE.",
		}), logs.String())
	})
}
