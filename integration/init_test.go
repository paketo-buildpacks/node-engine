package integration

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/dagger"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

var (
	nodeBuildpack        string
	offlineNodeBuildpack string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())

	nodeBuildpack, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	offlineNodeBuildpack, _, err = dagger.PackageCachedBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	// HACK: we need to fix dagger and the package.sh scripts so that this isn't required
	nodeBuildpack = fmt.Sprintf("%s.tgz", nodeBuildpack)
	offlineNodeBuildpack = fmt.Sprintf("%s.tgz", offlineNodeBuildpack)

	defer func() {
		dagger.DeleteBuildpack(nodeBuildpack)
		dagger.DeleteBuildpack(offlineNodeBuildpack)
	}()

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Logging", testLogging)
	suite("Offline", testOffline)
	suite("OptimizeMemory", testOptimizeMemory)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)
	suite.Run(t)
}

func GetBuildLogs(raw string) []string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "[builder]") {
			lines = append(lines, strings.TrimPrefix(line, "[builder] "))
		}
	}

	return lines
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git", lager.NewLogger("git logger"))
	execOut, _, err := gitExec.Execute(pexec.Execution{
		Args: []string{"describe", "--abbrev=0", "--tags"},
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(execOut, "v")), nil
}

func ContainSequence(expected interface{}) types.GomegaMatcher {
	return &containSequenceMatcher{
		expected: expected,
	}
}

type containSequenceMatcher struct {
	expected interface{}
}

func (matcher *containSequenceMatcher) Match(actual interface{}) (success bool, err error) {
	if reflect.TypeOf(actual).Kind() != reflect.Slice {
		return false, errors.New("not a slice")
	}

	expectedLength := reflect.ValueOf(matcher.expected).Len()
	actualLength := reflect.ValueOf(actual).Len()
	for i := 0; i < (actualLength - expectedLength + 1); i++ {
		aSlice := reflect.ValueOf(actual).Slice(i, i+expectedLength)
		eSlice := reflect.ValueOf(matcher.expected).Slice(0, expectedLength)

		match := true
		for j := 0; j < eSlice.Len(); j++ {
			aValue := aSlice.Index(j)
			eValue := eSlice.Index(j)

			if eMatcher, ok := eValue.Interface().(types.GomegaMatcher); ok {
				m, err := eMatcher.Match(aValue.Interface())
				if err != nil {
					return false, err
				}

				if !m {
					match = false
				}
			} else if !reflect.DeepEqual(aValue.Interface(), eValue.Interface()) {
				match = false
			}
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

func (matcher *containSequenceMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to contain sequence", matcher.expected)
}

func (matcher *containSequenceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to contain sequence", matcher.expected)
}
