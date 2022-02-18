package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
)

type SBOMGenerator struct {
	GenerateFromDependencyCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Dependency postal.Dependency
			Dir        string
		}
		Returns struct {
			SBOM  sbom.SBOM
			Error error
		}
		Stub func(postal.Dependency, string) (sbom.SBOM, error)
	}
}

func (f *SBOMGenerator) GenerateFromDependency(param1 postal.Dependency, param2 string) (sbom.SBOM, error) {
	f.GenerateFromDependencyCall.mutex.Lock()
	defer f.GenerateFromDependencyCall.mutex.Unlock()
	f.GenerateFromDependencyCall.CallCount++
	f.GenerateFromDependencyCall.Receives.Dependency = param1
	f.GenerateFromDependencyCall.Receives.Dir = param2
	if f.GenerateFromDependencyCall.Stub != nil {
		return f.GenerateFromDependencyCall.Stub(param1, param2)
	}
	return f.GenerateFromDependencyCall.Returns.SBOM, f.GenerateFromDependencyCall.Returns.Error
}
