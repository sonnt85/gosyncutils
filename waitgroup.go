package gosyncutils

import (
	"fmt"
	"sync"
)

// SafeWaitGroup must not be copied after first use.
type SafeWaitGroup struct {
	sync.WaitGroup
	mu sync.RWMutex
	// wait indicate whether Wait is called, if true,
	// then any Add with positive delta will return error.
	cnt  int
	wait bool
}

// Add adds delta, which may be negative, similar to sync.WaitGroup.
// If Add with a positive delta happens after Wait, it will return error,
// which prevent unsafe Add.
func (wg *SafeWaitGroup) Add(delta int) error {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	if wg.wait && delta > 0 {
		return fmt.Errorf("add with positive delta after Wait is forbidden")
	}
	wg.cnt += delta
	wg.WaitGroup.Add(delta)
	return nil
}

// Done decrements the WaitGroup counter.
func (wg *SafeWaitGroup) Done() {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.cnt -= 1
	wg.WaitGroup.Done()
}

// Done decrements the WaitGroup counter.
func (wg *SafeWaitGroup) Count() int {
	wg.mu.RLock()
	defer wg.mu.RUnlock()
	return wg.cnt
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *SafeWaitGroup) Wait() {
	wg.mu.Lock()
	wg.wait = true
	wg.mu.Unlock()
	wg.WaitGroup.Wait()
}
