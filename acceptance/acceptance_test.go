package acceptance

import (
	"path/filepath"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nodejs buildpack", func() {
	var (
		rootDir string
		dagg    *dagger.Dagger
	)

	BeforeEach(func() {
		var err error

		rootDir, err = dagger.FindRoot()
		Expect(err).ToNot(HaveOccurred())

		dagg, err = dagger.NewDagger(rootDir)
		Expect(err).ToNot(HaveOccurred())

		err = dagg.BundleBuildpack()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dagg.Destroy()
	})

	It("should create a working app in an OCI image", func() {
		app, err := dagg.Pack(filepath.Join(rootDir, "fixtures", "simple_app"))
		Expect(err).ToNot(HaveOccurred())

		err = app.Start()
		Expect(err).ToNot(HaveOccurred())
		defer app.Destroy()

		err = app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
	})
})
