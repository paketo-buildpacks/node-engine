package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	bp string
)

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())
	bp, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())
	defer func() {
		dagger.DeleteBuildpack(bp)
	}()

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))

}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("sets max_old_space_size when nodejs.optimize-memory is set in buildpack.yml", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "optimize_memory"), bp)
		Expect(err).ToNot(HaveOccurred())
		app.Memory = "128m"
		defer app.Destroy()

		Expect(app.StartWithCommand("node server.js")).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("NodeOptions: --max_old_space_size=96"))
	})
}
