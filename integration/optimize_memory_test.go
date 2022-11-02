package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		source, err = occam.Source(filepath.Join("testdata", "optimize_memory"))
		Expect(err).NotTo(HaveOccurred())

		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("never").
			WithBuildpacks(
				settings.Buildpacks.NodeEngine.Online,
				settings.Buildpacks.Processes.Online,
			).
			WithEnv(map[string]string{"BP_NODE_OPTIMIZE_MEMORY": "true"}).
			Execute(name, source)

		Expect(err).NotTo(HaveOccurred())

		container, err = docker.Container.Run.
			WithMemory("128m").
			WithPublish("8080").
			WithEnv(map[string]string{"NODE_OPTIONS": "--no-warnings"}).
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("NodeOptions: --no-warnings --max_old_space_size=96")).OnPort(8080))

		Expect(logs).To(ContainLines(
			"  Configuring launch environment",
			`    NODE_ENV        -> "production"`,
			fmt.Sprintf(`    NODE_HOME       -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
			`    NODE_VERBOSE    -> "false"`,
			`    OPTIMIZE_MEMORY -> "true"`,
			`    SSL_CERT_DIR    -> "$SSL_CERT_DIR:/etc/ssl/certs"`,
			"",
			"    Writing exec.d/0-optimize-memory",
			"      Calculates available memory based on container limits at launch time.",
			"      Made available in the MEMORY_AVAILABLE environment variable.",
			"      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.",
			"      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE.",
		))
	})
}
