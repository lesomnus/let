package let

import (
	"context"
)

type Task interface {
	// Run runs the task body and returns the error from the task body.
	// After [Task.Stop] or [Task.Close], the task body is not run and [ErrClosed] is returned.
	Run(ctx context.Context) error
	// Stop gracefully stops a task.
	// Once Stop has been called on a task, it may not be reused;
	// future calls to Run will return ErrClosed.
	Stop(ctx context.Context) error
	// Close forcefully stops the task body if possible and
	// returns any error from closing the underlying resource.
	// Use [Stop] for graceful stop.
	Close() error
	// Wait blocks until the task body returns and all underlying resource cleaned up.
	// Wait returns the same error what Run returns.
	Wait() error
}

func Halt(t Task) error {
	err := t.Close()
	t.Wait()
	return err
}

type task struct {
	f func(ctx context.Context) error

	ctx    context.Context
	cancel context.CancelFunc

	token chan struct{}
	err   error

	closer
}

// NewWithContext creates a Task that allows only one Run at a time.
// The Run starts only after the previous Run has finished and blocked until the run finished.
// Cancel of the given context results Stop of the Task.
func NewWithContext(ctx context.Context, f func(ctx context.Context) error) Task {
	t := &task{
		f:     f,
		token: make(chan struct{}, 1),
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.token <- struct{}{}

	initCloser(&t.closer)
	return t
}

// New creates a Task that allows only one Run at a time.
// The Run starts only after the previous Run has finished and blocked until the run finished.
func New(f func(ctx context.Context) error) Task {
	return NewWithContext(context.Background(), f)
}

func Nop() Task {
	return New(func(ctx context.Context) error {
		return nil
	})
}

func (t *task) Run(ctx context.Context) error {
	// Check if the task is already closed.
	if t.closed.Load() {
		return ErrClosed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.done:
		return ErrClosed
	case <-t.token:
		// Previous task is done.
	}

	defer func() {
		t.token <- struct{}{}

		if t.ctx.Err() != nil {
			// Task was stopped so notifies that it was the last run.
			t.close()
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := context.AfterFunc(t.ctx, cancel)
	defer stop()

	t.err = t.f(ctx)
	return t.err
}

func (t *task) Stop(ctx context.Context) error {
	t.cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.token:
	case <-t.done:
	}

	t.close()
	return nil
}

func (t *task) Done() <-chan struct{} {
	return t.done
}

func (t *task) Close() error {
	t.cancel()
	select {
	case <-t.token:
	case <-t.done:
	}

	t.close()
	return nil
}

func (t *task) Wait() error {
	<-t.done
	return t.err
}
