package fakes

import (
	"sync"
	"time"
)

type Clock struct {
	NowCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Time time.Time
		}
		Stub func() time.Time
	}
}

func (f *Clock) Now() time.Time {
	f.NowCall.Lock()
	defer f.NowCall.Unlock()
	f.NowCall.CallCount++
	if f.NowCall.Stub != nil {
		return f.NowCall.Stub()
	}
	return f.NowCall.Returns.Time
}
