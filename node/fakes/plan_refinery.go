package fakes

import (
	"sync"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
)

type PlanRefinery struct {
	BillOfMaterialCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency node.BuildpackMetadataDependency
		}
		Returns struct {
			BuildpackPlan packit.BuildpackPlan
		}
		Stub func(node.BuildpackMetadataDependency) packit.BuildpackPlan
	}
}

func (f *PlanRefinery) BillOfMaterial(param1 node.BuildpackMetadataDependency) packit.BuildpackPlan {
	f.BillOfMaterialCall.Lock()
	defer f.BillOfMaterialCall.Unlock()
	f.BillOfMaterialCall.CallCount++
	f.BillOfMaterialCall.Receives.Dependency = param1
	if f.BillOfMaterialCall.Stub != nil {
		return f.BillOfMaterialCall.Stub(param1)
	}
	return f.BillOfMaterialCall.Returns.BuildpackPlan
}
