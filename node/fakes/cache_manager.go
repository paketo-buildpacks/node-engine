package fakes

import (
	"sync"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
)

type CacheManager struct {
	MatchCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Layer      packit.Layer
			Dependency node.BuildpackMetadataDependency
		}
		Returns struct {
			Bool  bool
			Error error
		}
		Stub func(packit.Layer, node.BuildpackMetadataDependency) (bool, error)
	}
}

func (f *CacheManager) Match(param1 packit.Layer, param2 node.BuildpackMetadataDependency) (bool, error) {
	f.MatchCall.Lock()
	defer f.MatchCall.Unlock()
	f.MatchCall.CallCount++
	f.MatchCall.Receives.Layer = param1
	f.MatchCall.Receives.Dependency = param2
	if f.MatchCall.Stub != nil {
		return f.MatchCall.Stub(param1, param2)
	}
	return f.MatchCall.Returns.Bool, f.MatchCall.Returns.Error
}
