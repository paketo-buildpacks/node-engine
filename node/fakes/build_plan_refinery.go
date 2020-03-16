package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/postal"
)

type BuildPlanRefinery struct {
	BillOfMaterialCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency postal.Dependency
		}
		Returns struct {
			BuildpackPlan packit.BuildpackPlan
		}
		Stub func(postal.Dependency) packit.BuildpackPlan
	}
}

func (f *BuildPlanRefinery) BillOfMaterial(param1 postal.Dependency) packit.BuildpackPlan {
	f.BillOfMaterialCall.Lock()
	defer f.BillOfMaterialCall.Unlock()
	f.BillOfMaterialCall.CallCount++
	f.BillOfMaterialCall.Receives.Dependency = param1
	if f.BillOfMaterialCall.Stub != nil {
		return f.BillOfMaterialCall.Stub(param1)
	}
	return f.BillOfMaterialCall.Returns.BuildpackPlan
}
