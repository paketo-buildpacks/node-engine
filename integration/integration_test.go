package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		bp    string
		npmBP string
	)

	it.Before(func() {
		RegisterTestingT(t)

		var err error

		err = dagger.BuildCFLinuxFS3()
		Expect(err).ToNot(HaveOccurred())

		bp, err = dagger.PackageBuildpack() // TODO can this return an absolute path?
		Expect(err).ToNot(HaveOccurred())

		npmBP, err = dagger.GetLatestBuildpack("npm-cnb")
		Expect(err).ToNot(HaveOccurred())
	})

	it("sets max_old_space_size when nodejs.optimize-memory is set in buildpack.yml", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "optimize_memory"), bp, npmBP)
		Expect(err).ToNot(HaveOccurred())
		app.Memory = "128m"
		defer app.Destroy()

		Expect(app.Start()).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("NodeOptions: --max_old_space_size=96"))
	})
}
