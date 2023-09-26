package gosyncutils

import (
	"fmt"
	"testing"
	"time"
)

func TestWaitgroup(t *testing.T) {
	wg := new(SafeWaitGroup)
	for i := 0; i < 5; i++ {
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
	fmt.Print("waiting..\n")
	wg.Wait()
	time.Sleep(time.Second * 2)
	wg.Wait()
	fmt.Print("done..\n")
	wg.Wait()
	fmt.Print("done..\n")

}
