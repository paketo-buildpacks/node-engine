package build_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestBuild(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Build Suite")
}
