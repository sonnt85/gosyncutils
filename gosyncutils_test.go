package gosyncutils

import (
	"testing"
	"time"
)

func TestWaitgroup(t *testing.T) {
	wg := NewSafeWaitGroup()
	cnt := 5
	for i := 0; i < cnt; i++ {
		j := i
		wg.Add(1)
		go func() {
			time.Sleep(time.Second * 1)
			if j != 5 {
				wg.Done()
			}
		}()
	}
	time.Sleep(time.Millisecond * 10)
	t.Log("waiting..")
	wg.Wait()
}
