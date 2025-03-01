package gosyncutils

import (
	"context"
	"time"
)

func WaitFor(ctx context.Context, interval time.Duration, timeout time.Duration, condition func() bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return timeoutCtx.Err()
		case <-ticker.C:
			if condition() {
				return nil
			}
		}
	}
}

func WaitUntil(ctx context.Context, interval time.Duration, timeout time.Duration, condition func() bool) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return timeoutCtx.Err()
		case <-ticker.C:
			if condition() {
				return nil
			}
		}
	}
}

func WaitWhile(ctx context.Context, interval time.Duration, timeout time.Duration, condition func() bool) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return timeoutCtx.Err()
		case <-ticker.C:
			if !condition() {
				return nil
			}
		}
	}
}
