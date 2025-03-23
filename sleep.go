package let

import (
	"context"
	"time"
)

func Sleep(d time.Duration) Task {
	return New(func(ctx context.Context) error {
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
