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

	invoke func(t Task, ctx context.Context)

	mutex sync.Mutex

	queue []Task
	tasks []Task

	run_ctx context.Context
	started bool
	stopped bool

	stop_err  error
	stop_done chan struct{}
}

func newRunner(ctx context.Context) *runner {
	r := &runner{
		queue: []Task{},
		tasks: []Task{},

		stop_done: make(chan struct{}),
	}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// The return errors of [Stop], [Close], and [Wait] is the result of
// joining the return errors of all child tasks using [errors.Join].
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
// To fail fast, like [golang.org/x/sync/errgroup.Group], use [NewGroup].
func NewRunnerWithContext(ctx context.Context) Runner {
	r := newRunner(ctx)
	r.invoke = r.run
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
	if !r.started {
		r.queue = append(r.queue, task)
		return nil
	}

	r.tasks = append(r.tasks, task)
	r.invoke(task, r.run_ctx)

	return nil
}

func (r *runner) start(ctx context.Context) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.run_ctx = ctx
	r.started = true

	for _, t := range r.queue {
		r.invoke(t, ctx)
	}

	r.tasks = r.queue
	r.queue = nil
}

func (*runner) run(t Task, ctx context.Context) {
	go t.Run(ctx)
}

func (r *runner) Run(ctx context.Context) error {
	r.start(ctx)
	<-r.ctx.Done()
	return nil
}

func (r *runner) stop() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	last := r.stopped

	r.stopped = true
	r.queue = nil

	return last
}

func (r *runner) stopTasks() {
	defer r.cancel()
	defer close(r.stop_done)

	ctx := context.Background()

	errs := []error{}
	for _, t := range slices.Backward(r.tasks) {
		if err := t.Stop(ctx); err != nil && err != ErrClosed {
			errs = append(errs, err)
		}
	}

	r.stop_err = errors.Join(errs...)
}

func (r *runner) Stop(ctx context.Context) error {
	if !r.stop() {
		go r.stopTasks()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.stop_done:
		return r.stop_err
	}
}

func (r *runner) Close() error {
	defer r.cancel()
	if !r.stop() {
		defer close(r.stop_done)
	}

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
	for _, t := range r.tasks {
		errs = append(errs, t.Wait())
	}

	return errors.Join(errs...)
}
