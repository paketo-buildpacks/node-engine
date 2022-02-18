package nodeengine_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildEnv  packit.Environment
		launchEnv packit.Environment
		layerPath string
		execdPath string

		buffer      *bytes.Buffer
		environment nodeengine.Environment
	)

	it.Before(func() {
		var err error
		layerPath, err = ioutil.TempDir("", "layer-dir")
		Expect(err).NotTo(HaveOccurred())

		file, err := os.CreateTemp("", "optimize-memory")
		Expect(err).NotTo(HaveOccurred())

		_, err = file.WriteString("optimize-memory-script")
		Expect(err).NotTo(HaveOccurred())
		Expect(file.Close())

		execdPath = file.Name()

		buildEnv = packit.Environment{}
		launchEnv = packit.Environment{}
		buffer = bytes.NewBuffer(nil)
		environment = nodeengine.NewEnvironment(nodeengine.NewLogEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(layerPath)).To(Succeed())
		Expect(os.Remove(execdPath)).To(Succeed())
	})

	context("Configure", func() {
		it("configures the environment variables", func() {
			err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(buildEnv).To(Equal(packit.Environment{
				"NODE_HOME.default":    layerPath,
				"NODE_ENV.default":     "production",
				"NODE_VERBOSE.default": "false",
			}))

			Expect(launchEnv).To(Equal(packit.Environment{
				"NODE_HOME.default":    layerPath,
				"NODE_ENV.default":     "production",
				"NODE_VERBOSE.default": "false",
			}))
		})

		context("when NODE_ENV, NODE_VERBOSE are set", func() {
			it.Before(func() {
				os.Setenv("NODE_ENV", "some-node-env-val")
				os.Setenv("NODE_VERBOSE", "some-node-verbose-val")
			})

			it.After(func() {
				os.Unsetenv("NODE_ENV")
				os.Unsetenv("NODE_VERBOSE")
			})

			it("configures build envs using given value", func() {
				err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, false)
				Expect(err).NotTo(HaveOccurred())

				Expect(buildEnv["NODE_ENV.default"]).To(Equal("some-node-env-val"))
				Expect(buildEnv["NODE_VERBOSE.default"]).To(Equal("some-node-verbose-val"))

				Expect(launchEnv["NODE_ENV.default"]).To(Equal("production"))
				Expect(launchEnv["NODE_VERBOSE.default"]).To(Equal("false"))
			})
		})

		it("writes an exec.d script for available memory and optimization", func() {
			err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, false)
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(layerPath, "exec.d", "0-optimize-memory"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("optimize-memory-script"))

			Expect(buffer.String()).To(ContainSubstring("  Configuring launch environment"))
			Expect(buffer.String()).To(ContainSubstring("    NODE_ENV     -> \"production\""))
			Expect(buffer.String()).To(ContainSubstring(fmt.Sprintf("    NODE_HOME    -> %q", layerPath)))
			Expect(buffer.String()).To(ContainSubstring("    NODE_VERBOSE -> \"false\""))
			Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/0-optimize-memory"))
			Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
			Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
		})

		context("when memory should be optimized", func() {
			it("sets the $OPTIMIZE_MEMORY environment variable", func() {
				err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(launchEnv["OPTIMIZE_MEMORY.default"]).To(Equal("true"))

				Expect(buffer.String()).To(ContainSubstring("  Configuring launch environment"))
				Expect(buffer.String()).To(ContainSubstring("    NODE_ENV        -> \"production\""))
				Expect(buffer.String()).To(ContainSubstring(fmt.Sprintf("    NODE_HOME       -> %q", layerPath)))
				Expect(buffer.String()).To(ContainSubstring("    NODE_VERBOSE    -> \"false\""))
				Expect(buffer.String()).To(ContainSubstring("    OPTIMIZE_MEMORY -> \"true\""))
				Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/0-optimize-memory"))
				Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
				Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
				Expect(buffer.String()).To(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
				Expect(buffer.String()).To(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
			})
		})

		context("failure cases", func() {
			context("when the exec.d directory cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(layerPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(layerPath, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, false)
					Expect(err).To(MatchError(ContainSubstring("exec.d: permission denied")))
				})
			})

			context("when the exec.d script cannot be copied into the exec.d directory", func() {
				it.Before(func() {
					Expect(os.Chmod(execdPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(execdPath, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := environment.Configure(buildEnv, launchEnv, layerPath, execdPath, false)
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("%s: permission denied", execdPath))))
				})
			})
		})
	})
}
