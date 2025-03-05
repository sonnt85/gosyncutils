package gosyncutils

import (
	"reflect"
	"sync"
	"time"
)

type EventOpject[T any] struct {
	Obj T
	*sync.Mutex
	*sync.Cond
	broacast *sync.Cond
}

func NewEventOpject[T any]() *EventOpject[T] {
	var obj T
	mw := EventOpject[T]{
		Obj: obj,
	}
	mw.Mutex = new(sync.Mutex)
	mw.Cond = sync.NewCond(mw.Mutex)
	mw.broacast = sync.NewCond(new(sync.Mutex))
	return &mw
}

func (mw *EventOpject[T]) Get() T {
	mw.Lock()
	defer mw.Unlock()
	return mw.Obj
}

func (mw *EventOpject[T]) Set(value T) {
	mw.Lock()
	defer mw.Unlock()
	mw.Obj = value
}

func (mw *EventOpject[T]) SetThenSendSignal(value T) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendSignal()
}

func (mw *EventOpject[T]) SendSignal() {
	mw.Signal()
	mw.broacast.Broadcast()
}

// SendBroacast sends a broadcast signal to all goroutines waiting on the Cond or broacast.Cond.
// If not_included_common_brocast is not empty and its first element is true,
// only the broacast.Cond is notified. Otherwise, both the Cond and broacast.Cond are notified.
func (mw *EventOpject[T]) SendBroacast(not_included_common_brocast ...bool) {
	mw.Broadcast() // Notify all goroutines waiting on the Cond.
	if len(not_included_common_brocast) == 0 || !not_included_common_brocast[0] {
		mw.broacast.Broadcast()
	}
}

func (mw *EventOpject[T]) SendBroacastOnly() {
	mw.broacast.Broadcast()
}

func (mw *EventOpject[T]) WaitBroacast(value ...T) {
	mw.broacast.L.Lock()
	mw.broacast.Wait()
	mw.broacast.L.Unlock()
}

func (mw *EventOpject[T]) SetThenSendBroadcast(value T, not_included_common_brocast ...bool) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendBroacast(not_included_common_brocast...)
}

func (mw *EventOpject[T]) EditThenSendSignal(editFunc func(obj T) T) {
	mw.Lock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	mw.Unlock()
	mw.SendSignal()
	return
}

func (mw *EventOpject[T]) EditThenSendBroadcast(editFunc func(obj T) T, not_included_common_brocast ...bool) {
	mw.Lock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	mw.Unlock()
	mw.SendBroacast(not_included_common_brocast...)
}

func (mw *EventOpject[T]) Edit(editFunc func(obj T) T) {
	mw.Lock()
	defer mw.Unlock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
}

func (mw *EventOpject[T]) WaitUntil(condFunc func(obj T) bool, retviachan ...interface{}) chan struct{} {
	f := func() {
		mw.Lock()
		defer mw.Unlock()
		for !condFunc(mw.Obj) {
			mw.Wait()
		}
	}
	c := make(chan struct{}, 1)
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) WaitSignal(retviachan ...interface{}) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		mw.Wait()
		mw.Unlock()
	}
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) WaitSignalThenEdit(editFunc func(obj T) T, retviachan ...interface{}) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		mw.Wait()
		if editFunc != nil {
			newObj := editFunc(mw.Obj)
			mw.Obj = newObj
		}
		mw.Unlock()
	}
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}

// TestThenWaitSignalIfNotMatch waits for the event object to be unlocked
// and checks if the object's value is not equal to the compareValue. If the
// condition is met, the function waits for the event object to be signaled.
// If the retviachan argument is provided and true, the function returns a
// channel that receives a value after the event object is unlocked and the
// condition is met. The function also accepts a time.Duration value to set a
// timeout for the wait operation.
func (mw *EventOpject[T]) TestThenWaitSignalIfNotMatch(compareValue T, retviachan ...interface{}) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		if !reflect.DeepEqual(mw.Obj, compareValue) {
			mw.Wait()
		}
		mw.Unlock()
	}
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) TestThenWaitSignalIfMatch(compareValue T, retviachan ...interface{}) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		if reflect.DeepEqual(mw.Obj, compareValue) {
			mw.Wait()
		}
		mw.Unlock()
	}
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) TestThenEditIfMatch(testFunc func(obj T) bool, editFunc func(obj T) T) (b bool) {
	mw.Lock()
	if testFunc != nil {
		if testFunc(mw.Obj) {
			if editFunc != nil {
				newobj := editFunc(mw.Obj)
				mw.Obj = newobj
				b = true
			}
		}
	}
	mw.Unlock()
	return
}

func (mw *EventOpject[T]) WaitWhile(condFunc func(obj T) bool, retviachan ...interface{}) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		defer mw.Unlock()
		for condFunc(mw.Obj) {
			mw.Wait()
		}
	}
	timeout := time.Second
	viachanel := false
	if len(retviachan) != 0 {
		switch rv := retviachan[0].(type) {
		case time.Duration:
			timeout = rv
			viachanel = true
		case bool:
			viachanel = rv
		}
	}
	if viachanel {
		go func() {
			f()
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				return
			case c <- struct{}{}:
				return
			}
		}()
	} else {
		f()
	}
	return c
}
