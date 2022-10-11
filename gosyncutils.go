package gosyncutils

import (
	"reflect"
	"sync"
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
	return
}

func (mw *EventOpject[T]) SetThenSendSignal(value T) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendSignal()
	return
}

func (mw *EventOpject[T]) SendSignal() {
	mw.Signal()
	mw.broacast.Broadcast()
}

func (mw *EventOpject[T]) SendBroacast() {
	mw.Broadcast()
	mw.broacast.Broadcast()
}

func (mw *EventOpject[T]) SendBroacastOnly() {
	mw.broacast.Broadcast()
}

func (mw *EventOpject[T]) WaitBroacast(value T) {
	mw.broacast.L.Lock()
	mw.broacast.Wait()
	mw.broacast.L.Unlock()
}

func (mw *EventOpject[T]) SetThenSendBroadcast(value T) {
	mw.Lock()
	mw.Obj = value
	mw.Unlock()
	mw.SendBroacast()
	return
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

func (mw *EventOpject[T]) EditThenSendBroadcast(editFunc func(obj T) T) {
	mw.Lock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	mw.Unlock()
	mw.SendBroacast()
	return
}

func (mw *EventOpject[T]) Edit(editFunc func(obj T) T) {
	mw.Lock()
	defer mw.Unlock()
	if editFunc != nil {
		newObj := editFunc(mw.Obj)
		mw.Obj = newObj
	}
	return
}

func (mw *EventOpject[T]) WaitUntil(condFunc func(obj T) bool, retviachan ...bool) chan struct{} {
	f := func() {
		mw.Lock()
		defer mw.Unlock()
		for !condFunc(mw.Obj) {
			mw.Wait()
		}
	}
	c := make(chan struct{}, 1)

	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f() //may be leak goroutine beause
			c <- struct{}{}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) WaitSignal(retviachan ...bool) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		mw.Wait()
		mw.Unlock()
	}
	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f()
			c <- struct{}{}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) WaitSignalThenEdit(editFunc func(obj T) T, retviachan ...bool) chan struct{} {
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
	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f()
			c <- struct{}{}
		}()
	} else {
		f()
	}
	return c
}
func (mw *EventOpject[T]) TestThenWaitSignalIfNotMatch(compareValue T, retviachan ...bool) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		if !reflect.DeepEqual(mw.Obj, compareValue) {
			mw.Wait()
		}
		mw.Unlock()
	}
	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f()
			c <- struct{}{}
		}()
	} else {
		f()
	}
	return c
}

func (mw *EventOpject[T]) TestThenWaitSignalIfMatch(compareValue T, retviachan ...bool) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		if reflect.DeepEqual(mw.Obj, compareValue) {
			mw.Wait()
		}
		mw.Unlock()
	}
	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f()
			c <- struct{}{}
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

func (mw *EventOpject[T]) WaitWhile(condFunc func(obj T) bool, retviachan ...bool) chan struct{} {
	c := make(chan struct{}, 1)
	f := func() {
		mw.Lock()
		defer mw.Unlock()
		for condFunc(mw.Obj) {
			mw.Wait()
		}
	}
	if len(retviachan) != 0 && retviachan[0] {
		go func() {
			f()
			c <- struct{}{}
		}()
	} else {
		f()
	}
	return c
}
