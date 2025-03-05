package gosyncutils

// SafeWaitGroup must not be copied after first use.
type SafeWaitGroup struct {
	*EventOpject[int]
}

func NewSafeWaitGroup() *SafeWaitGroup {
	return &SafeWaitGroup{
		EventOpject: NewEventOpject[int](),
	}
}

// Add adds delta, which may be negative, similar to sync.WaitGroup.
// If Add with a positive delta happens after Wait, it will return error,
// which prevent unsafe Add.
func (wg *SafeWaitGroup) Add(delta int) error {
	wg.EditThenSendBroadcast(func(obj int) int {
		if delta < 0 {
			return obj
		}
		return obj + delta
	})
	return nil
}

// Done decrements the WaitGroup counter.
func (wg *SafeWaitGroup) Done() {
	wg.EditThenSendBroadcast(func(obj int) int {
		if obj <= 1 {
			return 0
		}
		return obj - 1
	})
}

// Done decrements the WaitGroup counter.
func (wg *SafeWaitGroup) ResetCount() {
	wg.EditThenSendBroadcast(func(obj int) int {
		return 0
	})
}

// Done decrements the WaitGroup counter.
func (wg *SafeWaitGroup) Count() int {
	return wg.Get()
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *SafeWaitGroup) Wait() {
	wg.WaitWhile(func(obj int) bool {
		return obj != 0
	})
}
