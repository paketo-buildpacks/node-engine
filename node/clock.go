package node

import "time"

type Clock struct {
	now func() time.Time
}

func NewClock(now func() time.Time) Clock {
	return Clock{now: now}
}

func (c Clock) Now() time.Time {
	return c.now()
}
