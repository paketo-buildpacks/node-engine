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
			Path           string
			OptimizeMemory bool
		}
		Returns struct {
			Error error
		}
		Stub func(packit.Environment, packit.Environment, string, bool) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 packit.Environment, param2 packit.Environment, param3 string, param4 bool) error {
	f.ConfigureCall.mutex.Lock()
	defer f.ConfigureCall.mutex.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.BuildEnv = param1
	f.ConfigureCall.Receives.LaunchEnv = param2
	f.ConfigureCall.Receives.Path = param3
	f.ConfigureCall.Receives.OptimizeMemory = param4
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1, param2, param3, param4)
	}
	return f.ConfigureCall.Returns.Error
}
