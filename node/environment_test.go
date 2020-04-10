package node_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/paketo-buildpacks/node-engine/node"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

var (
	MemoryAvailableScript = strings.TrimSpace(`
if [[ -z "$MEMORY_AVAILABLE" ]]; then
		memory_in_bytes="$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)"
		MEMORY_AVAILABLE="$(( $memory_in_bytes / ( 1024 * 1024 ) ))"
fi
export MEMORY_AVAILABLE
`)

	OptimizeMemoryScript = `export NODE_OPTIONS="--max_old_space_size=$(( $MEMORY_AVAILABLE * 75 / 100 ))"`
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		env  packit.Environment
		path string

		buffer      *bytes.Buffer
		environment node.Environment
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "layer-dir")
		Expect(err).NotTo(HaveOccurred())

		env = packit.Environment{}
		buffer = bytes.NewBuffer(nil)
		environment = node.NewEnvironment(node.NewLogEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Configure", func() {
		it("configures the environment variables", func() {
			err := environment.Configure(env, path, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(env).To(Equal(packit.Environment{
				"NODE_HOME.override":    path,
				"NODE_ENV.override":     "production",
				"NODE_VERBOSE.override": "false",
			}))
		})

		it("writes a profile.d script for available memory", func() {
			err := environment.Configure(env, path, false)
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(path, "profile.d", "0_memory_available.sh"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(MemoryAvailableScript))

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    NODE_ENV     -> \"production\""))
			Expect(buffer.String()).To(ContainSubstring(fmt.Sprintf("    NODE_HOME    -> %q", path)))
			Expect(buffer.String()).To(ContainSubstring("    NODE_VERBOSE -> \"false\""))
			Expect(buffer.String()).To(ContainSubstring("    Writing profile.d/0_memory_available.sh"))
			Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
			Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
		})

		it("does not write a profile.d script for optimizing memory", func() {
			err := environment.Configure(env, path, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(path, "profile.d", "1_optimize_memory.sh")).NotTo(BeARegularFile())
			Expect(buffer.String()).NotTo(ContainSubstring("Writing profile.d/1_optimize_memory.sh"))
		})

		context("when memory should be optimized", func() {
			it("writes a profile.d script for optimizing memory", func() {
				err := environment.Configure(env, path, true)
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(filepath.Join(path, "profile.d", "1_optimize_memory.sh"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(OptimizeMemoryScript))

				Expect(buffer.String()).To(ContainSubstring("    Writing profile.d/1_optimize_memory.sh"))
				Expect(buffer.String()).To(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
				Expect(buffer.String()).To(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
			})
		})

		context("failure cases", func() {
			context("when the profile.d directory cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := environment.Configure(env, path, false)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the 0_memory_available.sh script cannot be created", func() {
				it.Before(func() {
					scriptPath := filepath.Join(path, "profile.d", "0_memory_available.sh")

					Expect(os.MkdirAll(filepath.Dir(scriptPath), os.ModePerm)).To(Succeed())

					_, err := os.Create(scriptPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Chmod(scriptPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := environment.Configure(env, path, false)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the 1_optimize_memory.sh script cannot be created", func() {
				it.Before(func() {
					scriptPath := filepath.Join(path, "profile.d", "1_optimize_memory.sh")

					Expect(os.MkdirAll(filepath.Dir(scriptPath), os.ModePerm)).To(Succeed())

					_, err := os.Create(scriptPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Chmod(scriptPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := environment.Configure(env, path, true)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
