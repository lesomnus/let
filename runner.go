package let

import (
	"context"
	"errors"
	"sync"
)

type Runner interface {
	Task

	// Go runs the given Task in a new goroutine.
	Go(t Task)
}

type runner struct {
	ctx    context.Context
	cancel context.CancelFunc

	mutex   sync.Mutex
	tasks   []Task
	stopped bool
}

func NewRunnerWithContext(ctx context.Context) Runner {
	r := &runner{tasks: []Task{}}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

func NewRunner() Runner {
	return NewRunnerWithContext(context.Background())
}

func (r *runner) Go(task Task) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return
	}

	r.tasks = append(r.tasks, task)

	go task.Run(r.ctx)
}

func (r *runner) Run(ctx context.Context) error {
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
	for _, t := range r.tasks {
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
	for _, t := range r.tasks {
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
