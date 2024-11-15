package internal_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/paketo-buildpacks/node-engine/v5/cmd/inspector/internal"
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
	})

	it.After(func() {
		Expect(os.RemoveAll(root)).To(Succeed())
	})

	context("when $BPL_DEBUG_ENABLE is set", func() {
		it.Before(func() {
			environment["BPL_DEBUG_ENABLED"] = "true"
		})

		it("--inspect is added to NODE_OPTIONS", func() {
			environment["NODE_OPTIONS"] = "--existing"
			buffer := bytes.NewBuffer(nil)

			err := internal.Run(environment, buffer, root)
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(MatchTOML(`
				NODE_OPTIONS = "--existing --inspect"
			`))
		})

		context("when $BPL_DEBUG_PORT is set", func() {
			it.Before(func() {
				environment["BPL_DEBUG_PORT"] = "1111"
			})

			it("sets the inspector port", func() {
				buffer := bytes.NewBuffer(nil)

				err := internal.Run(environment, buffer, root)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(MatchTOML(`
					NODE_OPTIONS = "--inspect=127.0.0.1:1111"
				`))
			})
		})

		context("when $NODE_OPTIONS contains --inspect flag", func() {
			it.Before(func() {
				environment["NODE_OPTIONS"] = "--inspect=0.0.0.0:8888"
			})

			it("does not change it", func() {
				buffer := bytes.NewBuffer(nil)

				err := internal.Run(environment, buffer, root)
				Expect(err).NotTo(HaveOccurred())

				Expect(buffer.String()).To(BeEmpty())
			})
		})
	})

	context("when $BPL_DEBUG_ENABLED is not set", func() {
		it("NODE_OPTIONS are not changed", func() {
			environment["NODE_OPTIONS"] = "--existing"

			buffer := bytes.NewBuffer(nil)
			err := internal.Run(environment, buffer, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(BeEmpty())
		})
	})

}
