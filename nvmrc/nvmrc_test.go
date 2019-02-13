package nvmrc

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"path/filepath"
	"testing"
)

func TestUnitNvmrc(t *testing.T) {
	spec.Run(t, "Nvmrc", testNvmrc, spec.Report(report.Terminal{}))
}

func testNvmrc(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory
	var nvmrcPath string

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
		nvmrcPath = filepath.Join(factory.Detect.Application.Root, ".nvmrc")
	})

	when("nvmrc is empty", func() {
		it("should fail", func() {
			test.WriteFile(t, nvmrcPath, "")
			_, err := GetVersion(nvmrcPath, factory.Detect.Logger)
			Expect(err).To(HaveOccurred())
		})
	})

	when("there are .nvmrc contents", func() {
		when("the .nvmrc contains only digits", func() {
			it("will trim and transform nvmrc to appropriate semver for Masterminds semver library", func() {
				testCases := [][]string{
					{"10", "10.*.*"},
					{"10.2", "10.2.*"},
					{"v10", "10.*.*"},
					{"10.2.3", "10.2.3"},
					{"v10.2.3", "10.2.3"},
					{"10.1.1", "10.1.1"},
				}

				for _, testCase := range testCases {
					test.WriteFile(t, nvmrcPath, testCase[0])
					Expect(GetVersion(nvmrcPath, factory.Detect.Logger)).To(Equal(testCase[1]), fmt.Sprintf("failed for test case %s : %s", testCase[0], testCase[1]))
				}
			})
		})

		when("the .nvmrc contains lts/something", func() {
			it("will read and trim lts versions to appropriate semver for Masterminds semver library", func() {
				testCases := [][]string{
					{"lts/*", "10.*.*"},
					{"lts/argon", "4.*.*"},
					{"lts/boron", "6.*.*"},
					{"lts/carbon", "8.*.*"},
					{"lts/dubnium", "10.*.*"},
				}

				for _, testCase := range testCases {
					test.WriteFile(t, nvmrcPath, testCase[0])
					Expect(GetVersion(nvmrcPath, factory.Detect.Logger)).To(Equal(testCase[1]), fmt.Sprintf("failed for test case %s : %s", testCase[0], testCase[1]))
				}
			})
		})

		when("the .nvmrc contains 'node'", func() {
			it("should read and trim lts versions", func() {
				test.WriteFile(t, nvmrcPath, "node")
				Expect(GetVersion(nvmrcPath, factory.Detect.Logger)).To(Equal("*"))
			})
		})

		when("given an invalid .nvmrc", func() {
			it("validate should be fail", func() {
				invalidVersions := []string{"11.4.x", "invalid", "~1.1.2", ">11.0", "< 11.4.2", "^1.2.3", "11.*.*", "10.1.X", "lts/invalidname"}
				InvalidVersionError := errors.New("invalid version Invalid Semantic Version specified in .nvmrc")
				for _, version := range invalidVersions {
					test.WriteFile(t, nvmrcPath, version)
					parsedVersion, err := GetVersion(nvmrcPath, factory.Detect.Logger)
					Expect(err).To(Equal(InvalidVersionError))
					Expect(parsedVersion).To(BeEmpty())
				}
			})
		})
	})
}
