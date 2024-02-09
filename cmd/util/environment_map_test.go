package util_test

import (
	"testing"

	"github.com/paketo-buildpacks/node-engine/cmd/util"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironmentMap(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("loads the environment into a map", func() {
		env := util.LoadEnvironmentMap([]string{
			"SOME_KEY=some-value",
			"OTHER_KEY=other-value=extra-value",
		})

		Expect(env).To(Equal(map[string]string{
			"SOME_KEY":  "some-value",
			"OTHER_KEY": "other-value=extra-value",
		}))
	})
}
