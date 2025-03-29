package let

import (
	"context"
	"errors"
	"slices"
	"sync"
)

type group struct {
	ctx    context.Context
	cancel context.CancelFunc

	mutex   sync.Mutex
	tasks   []Task
	stopped bool

	wg  sync.WaitGroup
	err error
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// If a task fails, all its child tasks are stopped.
// The [Stop], [Close], and [Wait] methods return the error from the first failed task.
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
func NewGroupWithContext(ctx context.Context) Runner {
	r := &group{tasks: []Task{}}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// If a task fails, all its child tasks are stopped.
// The [Stop], [Close], and [Wait] methods return the error from the first failed task.
func NewGroup() Runner {
	return NewGroupWithContext(context.Background())
}

func (r *group) Go(task Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return ErrClosed
	}

	r.tasks = append(r.tasks, task)

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		err := task.Run(r.ctx)

		r.mutex.Lock()
		defer r.mutex.Unlock()
		if r.stopped {
			// This is not the first return of the Run
			// or the Run is stopped by context cancel.
			return
		}

		r.stopped = true
		r.err = err
		r.stopTasks(context.Background())
	}()
	return nil
}

func (r *group) Run(ctx context.Context) error {
	return nil
}

func (r *group) stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.stopped = true
}

func (r *group) stopTasks(ctx context.Context) error {
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

func (r *group) Stop(ctx context.Context) error {
	r.stop()

	return r.stopTasks(ctx)
}

func (r *group) Close() error {
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

func (r *group) Wait() error {
	<-r.ctx.Done()

	for _, t := range r.tasks {
		t.Wait()
	}

	return r.err
}
