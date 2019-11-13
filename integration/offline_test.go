package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOffline(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("when offline", func() {
		it("sets max_old_space_size when nodejs.optimize-memory is set in buildpack.yml", func() {
			app, err := dagger.NewPack(
				filepath.Join("testdata", "optimize_memory"),
				dagger.RandomImage(),
				dagger.SetBuildpacks(offlineNodeBuildpack),
				dagger.SetOffline(),
			).Build()
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			app.Memory = "128m"

			Expect(app.StartWithCommand("node server.js")).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("NodeOptions: --max_old_space_size=96"))
		})
	})
}
