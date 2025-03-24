package let

import (
	"context"
	"errors"
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

func NewGroupWithContext(ctx context.Context) Runner {
	r := &group{tasks: []Task{}}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

func NewGroup() Runner {
	return NewGroupWithContext(context.Background())
}

func (r *group) Go(task Task) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return
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

func (r *group) Stop(ctx context.Context) error {
	r.stop()

	return r.stopTasks(ctx)
}

func (r *group) Close() error {
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

func (r *group) Wait() error {
	<-r.ctx.Done()

	for _, t := range r.tasks {
		t.Wait()
	}

	return r.err
}
