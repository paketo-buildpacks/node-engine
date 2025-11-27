package nodeengine_test

import (
	"fmt"
	"os"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine/v5"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNodeVersionParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser nodeengine.NodeVersionParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", ".node-version")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()
		Expect(file.Close()).To(Succeed())

		parser = nodeengine.NewNodeVersionParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns a version constraint", func() {
		testCases := map[string]string{
			"10.2":    "10.2",
			"10.2.3":  "10.2.3",
			"v10.2.3": "10.2.3",
		}

		for input, output := range testCases {
			err := os.WriteFile(path, []byte(input), 0644)
			Expect(err).NotTo(HaveOccurred())

			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(output), fmt.Sprintf("input of %q failed to produce output of %q", input, output))
		}
	})

	context("when the .node-version file does not exist", func() {
		it.Before(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("returns an empty version", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeEmpty())
		})
	})

	context("failure cases", func() {
		context("when the .node-version contains a malformed semver number", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte("1.2.this is not a number"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseVersion(path)
				Expect(err).To(MatchError("invalid version constraint specified in .node-version: \"1.2.this is not a number\""))
			})
		})

		context("when the .node-version contains a value prefixed with `lts/`", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte("lts/*"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseVersion(path)
				Expect(err).To(MatchError("invalid version constraint specified in .node-version: \"lts/*\""))
			})
		})

		context("when the .node-version contains the value `node`", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte("node"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseVersion(path)
				Expect(err).To(MatchError("invalid version constraint specified in .node-version: \"node\""))
			})
		})
	})
}
