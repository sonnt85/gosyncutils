package gosyncutils

import (
	"time"
)

// SemWait must not be copied after first use.
type SemWait struct {
	eov *EventOpject[int]
}

func NewSemWait() *SemWait {
	obj := NewEventOpject[int]()
	return &SemWait{
		eov: obj,
	}
}

// Add adds delta, which may be negative, similar to sync.WaitGroup.
// If Add with a positive delta happens after Wait, it will return error,
// which prevent unsafe Add.
func (sw *SemWait) Add(delta int) {
	sw.eov.EditThenSendBroadcast(func(obj int) int {
		return obj + delta
	})
}

// Done decrements the WaitGroup counter.
func (sw *SemWait) Done() {
	sw.eov.EditThenSendBroadcast(func(obj int) int {
		return obj - 1
	})
}

// Done decrements the WaitGroup counter.
func (sw *SemWait) ResetCount() {
	sw.eov.EditThenSendBroadcast(func(obj int) int {
		return 0
	})
}

// Done decrements the WaitGroup counter.
func (sw *SemWait) Count() int {
	return sw.eov.Get()
}

// Wait blocks until the WaitGroup counter is zero.
func (sw *SemWait) Wait() {
	sw.eov.WaitUntil(func(obj int) bool {
		return obj == 0
	})
}

func (sw *SemWait) WaitWithTime(t time.Duration) bool {
	c := sw.eov.WaitUntil(func(obj int) bool {
		return obj == 0
	}, true)

	select {
	case <-c:
		return true
	case <-time.After(t):
		return false
	}
}
func (sw *SemWait) GetChanelWait() chan struct{} {
	return sw.eov.WaitUntil(func(obj int) bool {
		return obj == 0
	}, true)
}
