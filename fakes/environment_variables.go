package fakes

import "sync"

type EnvironmentVariables struct {
	OverrideCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Key   string
			Value string
		}
		Stub func(string, string)
	}
}

func (f *EnvironmentVariables) Override(param1 string, param2 string) {
	f.OverrideCall.Lock()
	defer f.OverrideCall.Unlock()
	f.OverrideCall.CallCount++
	f.OverrideCall.Receives.Key = param1
	f.OverrideCall.Receives.Value = param2
	if f.OverrideCall.Stub != nil {
		f.OverrideCall.Stub(param1, param2)
	}
}
