package node_test

import (
	"testing"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanRefiner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect      = NewWithT(t).Expect
		planRefiner node.PlanRefiner
		dependency  node.BuildpackMetadataDependency
	)

	it.Before(func() {
		planRefiner = node.NewPlanRefiner()

		dependency = node.BuildpackMetadataDependency{
			ID:      "some-id",
			Name:    "some-name",
			Stacks:  node.BuildpackMetadataDependencyStacks{"some-stack"},
			URI:     "some-uri",
			SHA256:  "some-sha",
			Version: "some-version",
		}
	})

	context("BillOfMaterial", func() {
		it("", func() {
			refinedBuildPlan := planRefiner.BillOfMaterial(dependency)
			Expect(refinedBuildPlan.Entries).To(HaveLen(1))
			Expect(refinedBuildPlan.Entries[0].Name).To(Equal("some-id"))
			Expect(refinedBuildPlan.Entries[0].Version).To(Equal("some-version"))
			Expect(refinedBuildPlan.Entries[0].Metadata).To(Equal(map[string]interface{}{
				"licenses": []string{},
				"name":     "some-name",
				"sha256":   "some-sha",
				"stacks":   node.BuildpackMetadataDependencyStacks{"some-stack"},
				"uri":      "some-uri",
			},
			))
		})
	})

}
