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

func testInspector(t *testing.T, context spec.G, it spec.S) {
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

	it("sets --inspect if set with env variable BPL_DEBUG_ENABLED", func() {
		var err error
		source, err = occam.Source(filepath.Join("testdata", "simple_app"))
		Expect(err).NotTo(HaveOccurred())

		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("never").
			WithBuildpacks(
				settings.Buildpacks.NodeEngine.Online,
				settings.Buildpacks.Processes.Online,
			).
			WithEnv(map[string]string{
				"BP_NODE_EXCLUDE_BUILD_PYTHON": "",
			}).
			Execute(name, source)

		Expect(err).NotTo(HaveOccurred())
		Expect(logs).To(ContainLines(
			"    Writing exec.d/1-inspector",
		))

		container, err = docker.Container.Run.
			WithMemory("128m").
			WithPublish("8080").
			WithEnv(map[string]string{"BPL_DEBUG_ENABLED": "true", "BPL_DEBUG_PORT": "9000", "NODE_OPTIONS": "--no-warnings"}).
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("NodeOptions: --no-warnings --inspect=127.0.0.1:9000")).OnPort(8080).WithEndpoint("/node-options"))

	})
}
