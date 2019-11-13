package node_test

import (
	"testing"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testCacheHandler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheHandler node.CacheHandler
	)

	it.Before(func() {
		cacheHandler = node.NewCacheHandler()
	})

	context("Match", func() {
		var layer packit.Layer
		it.Before(func() {
			layer = packit.Layer{
				Metadata: map[string]interface{}{
					node.DepKey: "some-sha",
				},
			}
		})

		context("when the layer metadata and choosen dependency shas match", func() {
			it("it returns true and no error", func() {
				match, err := cacheHandler.Match(layer, node.BuildpackMetadataDependency{SHA256: "some-sha"})
				Expect(err).ToNot(HaveOccurred())

				Expect(match).To(BeTrue())
			})
		})

		context("when the layer metadata and choosen dependency shas do not match", func() {
			it("it returns false and no error", func() {
				match, err := cacheHandler.Match(layer, node.BuildpackMetadataDependency{SHA256: "some-other-sha"})
				Expect(err).ToNot(HaveOccurred())

				Expect(match).To(BeFalse())
			})
		})

		context("when the layer metadata does not contain the dependency-sha", func() {
			it("it returns false and no error", func() {
				match, err := cacheHandler.Match(packit.Layer{}, node.BuildpackMetadataDependency{SHA256: "some-sha"})
				Expect(err).ToNot(HaveOccurred())

				Expect(match).To(BeFalse())
			})
		})

		context("failure cases", func() {
			it.Before(func() {
				layer.Metadata[node.DepKey] = 10
			})

			it.After(func() {
				layer.Metadata[node.DepKey] = "some-sha"
			})

			it("returns an error when the type in the layer metadata is not a string", func() {
				_, err := cacheHandler.Match(layer, node.BuildpackMetadataDependency{SHA256: "some-sha"})
				Expect(err).To(MatchError("layer metadata dependency-sha was not a string"))
			})
		})
	})

}
