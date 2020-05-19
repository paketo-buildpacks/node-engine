package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
)

type EntryResolver struct {
	ResolveCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			BuildpackPlanEntrySlice []packit.BuildpackPlanEntry
		}
		Returns struct {
			BuildpackPlanEntry packit.BuildpackPlanEntry
		}
		Stub func([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
	}
}

func (f *EntryResolver) Resolve(param1 []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	f.ResolveCall.Lock()
	defer f.ResolveCall.Unlock()
	f.ResolveCall.CallCount++
	f.ResolveCall.Receives.BuildpackPlanEntrySlice = param1
	if f.ResolveCall.Stub != nil {
		return f.ResolveCall.Stub(param1)
	}
	return f.ResolveCall.Returns.BuildpackPlanEntry
}
