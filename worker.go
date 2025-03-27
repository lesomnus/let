package let

import (
	"context"
	"sync"
)

type worker struct {
	ctx    context.Context
	cancel context.CancelFunc

	mutex   sync.Mutex
	tasks   map[Task]struct{}
	stopped bool
}

// NewWorker creates a Runner that runs multiple tasks simultaneously.
// Worker eats up all errors from the child tasks so [Stop], [Close] and [Wait]
// always returns nil regardless of any errors encountered by the child tasks.
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
// To collect all errors, use [NewRunner].
func NewWorkerWithContext(ctx context.Context) Runner {
	r := &worker{tasks: map[Task]struct{}{}}
	r.ctx, r.cancel = context.WithCancel(ctx)

	return r
}

// NewWorker creates a Runner that runs multiple tasks simultaneously.
// Worker eats up all errors from the child tasks so [Stop], [Close] and [Wait]
// always returns nil regardless of any errors encountered by the child tasks.
// To collect all errors, use [NewRunner].
func NewWorker() Runner {
	return NewWorkerWithContext(context.Background())
}

func (r *worker) Go(task Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return ErrClosed
	}

	r.tasks[task] = struct{}{}

	go func() {
		task.Run(r.ctx)

		r.mutex.Lock()
		defer r.mutex.Unlock()
		if r.stopped {
			return
		}

		task.Stop(context.Background())
		task.Wait()
		delete(r.tasks, task)
	}()
	return nil
}

func (r *worker) Run(ctx context.Context) error {
	return nil
}

func (r *worker) stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.stopped = true
}

func (r *worker) Stop(ctx context.Context) error {
	r.stop()

	for t := range r.tasks {
		t.Stop(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	r.cancel()
	return nil
}

func (r *worker) Close() error {
	r.stop()
	defer r.cancel()

	for t := range r.tasks {
		t.Close()
	}

	return nil
}

func (r *worker) Wait() error {
	<-r.ctx.Done()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for t := range r.tasks {
		t.Wait()
	}

	// Worker does not collect errors from the child tasks so free the memory.
	r.tasks = nil

	return nil
}
