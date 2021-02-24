package nodeengine_test

import (
	"io/ioutil"
	"os"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser nodeengine.BuildpackYMLParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`---
nodejs:
  version: 1.2.3
`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = nodeengine.NewBuildpackYMLParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		context("buildpack.yml sets optimize-memory", func() {
			it.Before(func() {
				err := ioutil.WriteFile(path, []byte(`---
nodejs:
  optimize-memory: true
  version: "1.2.3"
`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("parses a buildpack.yml file", func() {
				configData, err := parser.Parse(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(configData.OptimizedMemory).To(BeTrue())
				Expect(configData.Version).To(Equal("1.2.3"))
			})
		})
		context("buildpack.yml does not set optimize-memory", func() {
			it("parses a buildpack.yml file", func() {
				configData, err := parser.Parse(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(configData.OptimizedMemory).To(BeFalse())
				Expect(configData.Version).To(Equal("1.2.3"))
			})
		})
	})

	context("ParseVersion", func() {
		it("parses the node version from a buildpack.yml file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		context("when the buildpack.yml file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(BeEmpty())
			})
		})

		context("failure cases", func() {
			context("when the buildpack.yml file cannot be read", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(path, 0644)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the contents of the buildpack.yml file are malformed", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
				})
			})
		})
	})
}
