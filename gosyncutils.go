package gosyncutils

import (
	"reflect"
	"sync"
	"time"
)

type EventObject[T any] struct {
	Obj T
	*sync.Mutex
	*sync.Cond
	broacast *sync.Cond
}

func NewEventObject[T any]() *EventObject[T] {
	var obj T
	mw := EventObject[T]{
		Obj: obj,
	}
	mw.Mutex = new(sync.Mutex)
	mw.Cond = sync.NewCond(mw.Mutex)
	mw.broacast = sync.NewCond(new(sync.Mutex))
	return &mw
}

func (mw *EventObject[T]) Get() T {
	mw.Lock()
	defer mw.Unlock()
	return mw.Obj
}

func (mw *EventObject[T]) Set(value T) {
	mw.Lock()
	defer mw.Unlock()
	mw.Obj = value
}

func (mw *EventObject[T]) SetThenSendSignal(value T) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendSignal()
}

func (mw *EventObject[T]) SendSignal() {
	mw.Signal()
	mw.broacast.Broadcast()
}

// SendBroadcast sends a broadcast signal to all goroutines waiting on the Cond or broacast.Cond.
// If excludeCommonBroadcast is not empty and its first element is true,
// only the broacast.Cond is notified. Otherwise, both the Cond and broacast.Cond are notified.
func (mw *EventObject[T]) SendBroadcast(excludeCommonBroadcast ...bool) {
	mw.Broadcast()
	if len(excludeCommonBroadcast) == 0 || !excludeCommonBroadcast[0] {
		mw.broacast.Broadcast()
	}
}

// Deprecated: Use SendBroadcast instead.
func (mw *EventObject[T]) SendBroacast(excludeCommonBroadcast ...bool) {
	mw.SendBroadcast(excludeCommonBroadcast...)
}

// SendBroadcastOnly sends a broadcast signal only to the broacast.Cond.
func (mw *EventObject[T]) SendBroadcastOnly() {
	mw.broacast.Broadcast()
}

// Deprecated: Use SendBroadcastOnly instead.
func (mw *EventObject[T]) SendBroacastOnly() {
	mw.SendBroadcastOnly()
}

// WaitBroadcast blocks until a broadcast signal is received.
func (mw *EventObject[T]) WaitBroadcast(value ...T) {
	mw.broacast.L.Lock()
	mw.broacast.Wait()
	mw.broacast.L.Unlock()
}

// Deprecated: Use WaitBroadcast instead.
func (mw *EventObject[T]) WaitBroacast(value ...T) {
	mw.WaitBroadcast(value...)
}

func (mw *EventObject[T]) SetThenSendBroadcast(value T, excludeCommonBroadcast ...bool) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendBroadcast(excludeCommonBroadcast...)
}

func (mw *EventObject[T]) EditThenSendSignal(editFunc func(obj T) T) {
	mw.Lock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	mw.Unlock()
	mw.SendSignal()
	return
}

func (mw *EventObject[T]) EditThenSendBroadcast(editFunc func(obj T) T, excludeCommonBroadcast ...bool) {
	mw.Lock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	mw.Unlock()
	mw.SendBroadcast(excludeCommonBroadcast...)
}

func (mw *EventObject[T]) Edit(editFunc func(obj T) T) {
	mw.Lock()
	defer mw.Unlock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
}

func (mw *EventObject[T]) WaitUntil(condFunc func(obj T) bool, retviachan ...interface{}) chan struct{} {
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

func (mw *EventObject[T]) WaitSignal(retviachan ...interface{}) chan struct{} {
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

func (mw *EventObject[T]) WaitSignalThenEdit(editFunc func(obj T) T, retviachan ...interface{}) chan struct{} {
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
func (mw *EventObject[T]) TestThenWaitSignalIfNotMatch(compareValue T, retviachan ...interface{}) chan struct{} {
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

func (mw *EventObject[T]) TestThenWaitSignalIfMatch(compareValue T, retviachan ...interface{}) chan struct{} {
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

func (mw *EventObject[T]) TestThenEditIfMatch(testFunc func(obj T) bool, editFunc func(obj T) T) (b bool) {
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

func (mw *EventObject[T]) WaitWhile(condFunc func(obj T) bool, retviachan ...interface{}) chan struct{} {
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

// Deprecated: Use NewEventObject instead.
func NewEventOpject[T any]() *EventObject[T] {
	return NewEventObject[T]()
}
