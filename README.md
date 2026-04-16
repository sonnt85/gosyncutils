# gosyncutils

[![Go Reference](https://pkg.go.dev/badge/github.com/sonnt85/gosyncutils.svg)](https://pkg.go.dev/github.com/sonnt85/gosyncutils)

Synchronization utilities for Go — generic event objects, safe wait groups, semaphores, and polling helpers built on top of `sync.Cond`.

## Installation

```bash
go get github.com/sonnt85/gosyncutils
```

## Features

- `EventObject[T]` — generic condition variable wrapper that holds a typed value; supports `WaitUntil`, `WaitWhile`, `WaitSignal`, `WaitBroadcast`, and `Edit`/`Set` with automatic broadcast
- `SafeWaitGroup` — `sync.WaitGroup` alternative that never panics on negative counter; supports `Count()` and `ResetCount()`
- `SemWait` — semaphore-style wait with `Add`/`Done`/`Wait`, `WaitWithTime`, and channel-based wait
- `WaitFor` / `WaitUntil` / `WaitWhile` — polling helpers with ticker interval and timeout using `context.Context`

## Usage

```go
// EventObject: wait until a condition is met
obj := gosyncutils.NewEventObject[int]()
go func() {
    time.Sleep(100 * time.Millisecond)
    obj.SetThenSendBroadcast(42)
}()
obj.WaitUntil(func(v int) bool { return v == 42 })
fmt.Println(obj.Get()) // 42

// SafeWaitGroup
wg := gosyncutils.NewSafeWaitGroup()
wg.Add(3)
go func() { defer wg.Done(); /* work */ }()
wg.Wait()

// SemWait
sem := gosyncutils.NewSemWait()
sem.Add(1)
go func() { defer sem.Done(); /* work */ }()
sem.Wait()

// Polling helper
err := gosyncutils.WaitFor(ctx, 100*time.Millisecond, 5*time.Second, func() bool {
    return resourceReady()
})
```

## API

### EventObject[T]

- `NewEventObject[T]() *EventObject[T]` — create (alias: `NewEventOpject` deprecated)
- `Get() T` / `Set(value T)` — get/set the stored value
- `SetThenSendSignal(value T)` / `SetThenSendBroadcast(value T)` — set and notify
- `Edit(func(T) T)` / `EditThenSendSignal(...)` / `EditThenSendBroadcast(...)` — mutate and notify
- `SendSignal()` / `SendBroadcast()` / `SendBroadcastOnly()` — manual notify
- `WaitUntil(condFunc func(T) bool, ...) chan struct{}` — block until condition is true
- `WaitWhile(condFunc func(T) bool, ...) chan struct{}` — block while condition is true
- `WaitSignal(...) chan struct{}` — block until next signal
- `WaitBroadcast(...)` — block until next broadcast
- `WaitSignalThenEdit(editFunc, ...) chan struct{}` — wait for signal then edit
- `TestThenWaitSignalIfMatch(compareValue T, ...) chan struct{}` — conditional wait
- `TestThenWaitSignalIfNotMatch(compareValue T, ...) chan struct{}` — conditional wait
- `TestThenEditIfMatch(testFunc, editFunc func(T) T) bool` — atomic test-and-edit

### SafeWaitGroup

- `NewSafeWaitGroup() *SafeWaitGroup`
- `Add(delta int) error` / `Done()` / `Wait()` / `Count() int` / `ResetCount()`

### SemWait

- `NewSemWait() *SemWait`
- `Add(delta int)` / `Done()` / `Wait()` / `Count() int` / `ResetCount()`
- `WaitWithTime(t time.Duration) bool` — wait with timeout, returns true if completed
- `GetChanelWait() chan struct{}` — return channel that closes when count reaches zero

### Polling

- `WaitFor(ctx, interval, timeout, condition func() bool) error`
- `WaitUntil(ctx, interval, timeout, condition func() bool) error`
- `WaitWhile(ctx, interval, timeout, condition func() bool) error`

## Author

**sonnt85** — [thanhson.rf@gmail.com](mailto:thanhson.rf@gmail.com)

## License

MIT License - see [LICENSE](LICENSE) for details.
