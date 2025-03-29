package let

import (
	"context"
	"sync/atomic"
)

type wrapped struct {
	base Task
	f    func(ctx context.Context, next func(ctx context.Context) error) error

	closed atomic.Bool
}

// Wrap intercepts the Run of the given Task.
// Stop, Close, and Wait on the wrapped Task will Stop, Close, and Wait on the given Task.
func Wrap(t Task, f func(ctx context.Context, next func(ctx context.Context) error) error) Task {
	return &wrapped{base: t, f: f}
}

func (t *wrapped) Run(ctx context.Context) error {
	if t.closed.Load() {
		return ErrClosed
	}
	return t.f(ctx, t.base.Run)
}

func (t *wrapped) Stop(ctx context.Context) error {
	t.closed.Store(true)
	return t.base.Stop(ctx)
}

func (t *wrapped) Close() error {
	t.closed.Store(true)
	return t.base.Close()
}

func (t *wrapped) Wait() error {
	t.closed.Store(true)
	return t.base.Wait()
}
