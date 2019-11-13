package fakes

import (
	"sync"

	"github.com/cloudfoundry/node-engine-cnb/node"
)

type EnvironmentConfiguration struct {
	ConfigureCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Env            node.EnvironmentVariables
			Path           string
			OptimizeMemory bool
		}
		Returns struct {
			Error error
		}
		Stub func(node.EnvironmentVariables, string, bool) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 node.EnvironmentVariables, param2 string, param3 bool) error {
	f.ConfigureCall.Lock()
	defer f.ConfigureCall.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.Env = param1
	f.ConfigureCall.Receives.Path = param2
	f.ConfigureCall.Receives.OptimizeMemory = param3
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1, param2, param3)
	}
	return f.ConfigureCall.Returns.Error
}
