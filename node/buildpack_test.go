package node_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		path = file.Name()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("ParseBuildpack", func() {
		it.Before(func() {
			err := ioutil.WriteFile(path, []byte(`
api = "0.2"

[buildpack]
  id = "org.some-org.some-buildpack"
  name = "Some Buildpack"
  version = "some-version"

[metadata]
  [metadata.default-versions]
    node = "10.x"

  [[metadata.dependencies]]
    id = "some-dep"
    name = "Some Dep"
    sha256 = "some-sha"
    stacks = ["some-stack"]
    uri = "some-uri"
    version = "some-dep-version"
`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("parses a buildpack.yml file", func() {
			buildpack, err := node.ParseBuildpack(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(buildpack).To(Equal(node.Buildpack{
				Info: node.BuildpackInfo{
					ID:      "org.some-org.some-buildpack",
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Metadata: node.BuildpackMetadata{
					DefaultVersions: node.BuildpackMetadataDefaultVersions{
						Node: "10.x",
					},
					Dependencies: []node.BuildpackMetadataDependency{
						{
							ID:      "some-dep",
							Name:    "Some Dep",
							SHA256:  "some-sha",
							Stacks:  node.BuildpackMetadataDependencyStacks{"some-stack"},
							URI:     "some-uri",
							Version: "some-dep-version",
						},
					},
				},
			}))
		})

		context("failure cases", func() {
			context("when the file does not exist", func() {
				it("fails", func() {
					_, err := node.ParseBuildpack("not a real file")
					Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.toml:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
}
