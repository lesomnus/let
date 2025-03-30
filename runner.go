package let

import (
	"context"
	"errors"
	"slices"
	"sync"
)

// Runner runs multiple tasks simultaneously.
type Runner interface {
	Task

	// Go runs the given Task in a new goroutine.
	// Returns [ErrClosed] if the Runner stopped.
	Go(t Task) error
}

type runner struct {
	ctx    context.Context
	cancel context.CancelFunc

	mutex   sync.Mutex
	tasks   []Task
	stopped bool
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// The return errors of [Stop], [Close], and [Wait] is the result of
// joining the return errors of all child tasks using [errors.Join].
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
// To fail fast, like [golang.org/x/sync/errgroup.Group], use [NewGroup].
func NewRunnerWithContext(ctx context.Context) Runner {
	r := &runner{tasks: []Task{}}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// The return errors of [Stop], [Close], and [Wait] is the result of
// joining the return errors of all child tasks using [errors.Join].
// To fail fast, like [golang.org/x/sync/errgroup.Group], use [NewGroup].
func NewRunner() Runner {
	return NewRunnerWithContext(context.Background())
}

func (r *runner) Go(task Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return ErrClosed
	}

	r.tasks = append(r.tasks, task)

	go task.Run(r.ctx)
	return nil
}

func (r *runner) Run(ctx context.Context) error {
	<-r.ctx.Done()
	return nil
}

func (r *runner) stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.stopped = true
}

func (r *runner) Stop(ctx context.Context) error {
	r.stop()

	errs := []error{}
	for _, t := range slices.Backward(r.tasks) {
		if err := t.Stop(ctx); err != nil && err != ErrClosed {
			errs = append(errs, err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	r.cancel()
	return errors.Join(errs...)
}

func (r *runner) Close() error {
	r.stop()
	defer r.cancel()

	errs := []error{}
	for _, t := range slices.Backward(r.tasks) {
		if err := t.Close(); err != nil && err != ErrClosed {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *runner) Wait() error {
	<-r.ctx.Done()

	errs := []error{}
	for _, t := range slices.Backward(r.tasks) {
		errs = append(errs, t.Wait())
	}

	return errors.Join(errs...)
}
