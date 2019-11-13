package fakes

import (
	"sync"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
)

type DependencyManager struct {
	InstallCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency node.BuildpackMetadataDependency
			CnbPath    string
			LayerPath  string
		}
		Returns struct {
			Error error
		}
		Stub func(node.BuildpackMetadataDependency, string, string) error
	}
	ResolveCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependencies   []node.BuildpackMetadataDependency
			DefaultVersion string
			Stack          string
			Entry          packit.BuildpackPlanEntry
		}
		Returns struct {
			BuildpackMetadataDependency node.BuildpackMetadataDependency
			Error                       error
		}
		Stub func([]node.BuildpackMetadataDependency, string, string, packit.BuildpackPlanEntry) (node.BuildpackMetadataDependency, error)
	}
}

func (f *DependencyManager) Install(param1 node.BuildpackMetadataDependency, param2 string, param3 string) error {
	f.InstallCall.Lock()
	defer f.InstallCall.Unlock()
	f.InstallCall.CallCount++
	f.InstallCall.Receives.Dependency = param1
	f.InstallCall.Receives.CnbPath = param2
	f.InstallCall.Receives.LayerPath = param3
	if f.InstallCall.Stub != nil {
		return f.InstallCall.Stub(param1, param2, param3)
	}
	return f.InstallCall.Returns.Error
}
func (f *DependencyManager) Resolve(param1 []node.BuildpackMetadataDependency, param2 string, param3 string, param4 packit.BuildpackPlanEntry) (node.BuildpackMetadataDependency, error) {
	f.ResolveCall.Lock()
	defer f.ResolveCall.Unlock()
	f.ResolveCall.CallCount++
	f.ResolveCall.Receives.Dependencies = param1
	f.ResolveCall.Receives.DefaultVersion = param2
	f.ResolveCall.Receives.Stack = param3
	f.ResolveCall.Receives.Entry = param4
	if f.ResolveCall.Stub != nil {
		return f.ResolveCall.Stub(param1, param2, param3, param4)
	}
	return f.ResolveCall.Returns.BuildpackMetadataDependency, f.ResolveCall.Returns.Error
}
