package fakes

import (
	"sync"

	packit "github.com/paketo-buildpacks/packit/v2"
)

type EnvironmentConfiguration struct {
	ConfigureCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			BuildEnv       packit.Environment
			LaunchEnv      packit.Environment
			LayerPath      string
			ExecdPath      string
			OptimizeMemory bool
		}
		Returns struct {
			Error error
		}
		Stub func(packit.Environment, packit.Environment, string, string, bool) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 packit.Environment, param2 packit.Environment, param3 string, param4 string, param5 bool) error {
	f.ConfigureCall.mutex.Lock()
	defer f.ConfigureCall.mutex.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.BuildEnv = param1
	f.ConfigureCall.Receives.LaunchEnv = param2
	f.ConfigureCall.Receives.LayerPath = param3
	f.ConfigureCall.Receives.ExecdPath = param4
	f.ConfigureCall.Receives.OptimizeMemory = param5
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1, param2, param3, param4, param5)
	}
	return f.ConfigureCall.Returns.Error
}
