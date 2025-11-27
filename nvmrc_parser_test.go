package nodeengine_test

import (
	"fmt"
	"os"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine/v5"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNvmrcParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser nodeengine.NvmrcParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", ".nvmrc")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()
		Expect(file.Close()).To(Succeed())

		parser = nodeengine.NewNvmrcParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns a version constraint", func() {
		testCases := map[string]string{
			"10":           "10",
			"10.2":         "10.2",
			"v10":          "10",
			"10.2.3":       "10.2.3",
			"v10.2.3":      "10.2.3",
			"10.1.1":       "10.1.1",
			"10.1.*":       "10.1.*",
			"10.*":         "10.*",
			"lts/*":        "20.*",
			"lts/argon":    "4.*",
			"lts/boron":    "6.*",
			"lts/carbon":   "8.*",
			"lts/dubnium":  "10.*",
			"lts/erbium":   "12.*",
			"lts/fermium":  "14.*",
			"lts/gallium":  "16.*",
			"lts/hydrogen": "18.*",
			"lts/iron":     "20.*",
			"node":         "*",
		}

		for input, output := range testCases {
			err := os.WriteFile(path, []byte(input), 0644)
			Expect(err).NotTo(HaveOccurred())

			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(output), fmt.Sprintf("input of %q failed to produce output of %q", input, output))
		}
	})

	context("when the .nvmrc file does not exist", func() {
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
		context("when the .nvmrc contains a malformed semver number", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte("1.2.this is not a number"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := parser.ParseVersion(path)
				Expect(err).To(MatchError("invalid version constraint specified in .nvmrc: \"1.2.this is not a number\""))
			})
		})
	})
}
