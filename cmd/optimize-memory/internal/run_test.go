package internal_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/node-engine/v3/cmd/optimize-memory/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
)

func testRun(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		environment map[string]string
		root        string
	)

	it.Before(func() {
		environment = map[string]string{}

		var err error
		root, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(root, "sys", "fs", "cgroup", "memory"), os.ModePerm)).To(Succeed())

		err = os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory.max"), []byte("1073741824"), 0600)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory", "memory.limit_in_bytes"), []byte("2147483648"), 0600)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(root)).To(Succeed())
	})

	context("when $MEMORY_AVAILABLE is not already set", func() {
		context("when /sys/fs/cgroup/cgroup.controllers exists", func() {
			it.Before(func() {
				file, err := os.Create(filepath.Join(root, "sys", "fs", "cgroup", "cgroup.controllers"))
				Expect(err).NotTo(HaveOccurred())
				Expect(file.Close()).To(Succeed())
			})

			it("assigns it to the value of /sys/fs/cgroup/memory.max", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchTOML(`
					MEMORY_AVAILABLE = "1024"
				`))
			})

			context("when the memory value is empty", func() {
				it.Before(func() {
					err := os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory.max"), nil, 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("does not assign it", func() {
					buffer := bytes.NewBuffer(nil)
					err := internal.Run(environment, buffer, root)
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(BeEmpty())
				})
			})

			context("when the memory value is max", func() {
				it.Before(func() {
					err := os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory.max"), []byte("max"), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("does not assign it", func() {
					buffer := bytes.NewBuffer(nil)
					err := internal.Run(environment, buffer, root)
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(BeEmpty())
				})
			})
		})

		context("when /sys/fs/cgroup/cgroup.controllers does not exist", func() {
			it("assigns it to the value of /sys/fs/cgroup/memory/memory.limit_in_bytes", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchTOML(`
					MEMORY_AVAILABLE = "2048"
				`))
			})

			context("when the memory value is empty", func() {
				it.Before(func() {
					err := os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory", "memory.limit_in_bytes"), nil, 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("does not assign it", func() {
					buffer := bytes.NewBuffer(nil)
					err := internal.Run(environment, buffer, root)
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(BeEmpty())
				})
			})

			context("when the memory value is max", func() {
				it.Before(func() {
					err := os.WriteFile(filepath.Join(root, "sys", "fs", "cgroup", "memory", "memory.limit_in_bytes"), []byte("max"), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("does not assign it", func() {
					buffer := bytes.NewBuffer(nil)
					err := internal.Run(environment, buffer, root)
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(BeEmpty())
				})
			})
		})
	})

	context("when $MEMORY_AVAILABLE is already set", func() {
		it.Before(func() {
			environment["MEMORY_AVAILABLE"] = "4294967296"
		})

		it("uses the value already set", func() {
			buffer := bytes.NewBuffer(nil)
			err := internal.Run(environment, buffer, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchTOML(`
				MEMORY_AVAILABLE = "4096"
			`))
		})
	})

	context("when $OPTIMIZE_MEMORY is set", func() {
		it.Before(func() {
			environment["MEMORY_AVAILABLE"] = "4294967296"
			environment["OPTIMIZE_MEMORY"] = "true"
		})

		it("configures NODE_OPTIONS", func() {
			buffer := bytes.NewBuffer(nil)
			err := internal.Run(environment, buffer, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchTOML(`
				MEMORY_AVAILABLE = "4096"
				NODE_OPTIONS = "--max_old_space_size=3072"
			`))
		})

		context("when $NODE_OPTIONS is already set", func() {
			it.Before(func() {
				environment["NODE_OPTIONS"] = "--no-warnings"
			})

			it("merges this option onto the end", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(MatchTOML(`
				MEMORY_AVAILABLE = "4096"
				NODE_OPTIONS = "--no-warnings --max_old_space_size=3072"
			`))
			})
		})
	})

	context("failure cases", func() {
		context("when the cgroup.controllers file cannot be stat'd", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(root, "sys", "fs", "cgroup"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(root, "sys", "fs", "cgroup"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).To(MatchError(ContainSubstring("cgroup.controllers: permission denied")))
			})
		})

		context("when the memory files don't exist", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(root, "sys", "fs", "cgroup", "memory.max"))).To(Succeed())
				Expect(os.Remove(filepath.Join(root, "sys", "fs", "cgroup", "memory", "memory.limit_in_bytes"))).To(Succeed())
			})

			it("returns an error", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).To(MatchError(ContainSubstring("memory.limit_in_bytes: no such file or directory")))
			})
		})

		context("when the $MEMORY_AVAILABLE value is not an integer", func() {
			it.Before(func() {
				environment["MEMORY_AVAILABLE"] = "not-an-integer"
			})

			it("returns an error", func() {
				buffer := bytes.NewBuffer(nil)
				err := internal.Run(environment, buffer, root)
				Expect(err).To(MatchError(ContainSubstring(`parsing "not-an-integer": invalid syntax`)))
			})
		})

		context("when the output cannot be written to", func() {
			it("returns an error", func() {
				buffer, err := os.Create(filepath.Join(root, "output"))
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.Close()).To(Succeed())

				err = internal.Run(environment, buffer, root)
				Expect(err).To(MatchError(ContainSubstring("output: file already closed")))
			})
		})
	})
}
