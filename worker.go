package let

import (
	"context"
	"sync"
)

type worker struct {
	ctx    context.Context
	cancel context.CancelFunc

	mutex sync.Mutex

	queue map[Task]struct{}
	tasks map[Task]struct{}

	run_ctx context.Context
	started bool
	stopped bool

	stop_done chan struct{}

	wg sync.WaitGroup
}

// NewWorker creates a Runner that runs multiple tasks simultaneously.
// Worker eats up all errors from the child tasks so [Stop], [Close] and [Wait]
// always returns nil regardless of any errors encountered by the child tasks.
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
// To collect all errors, use [NewRunner].
func NewWorkerWithContext(ctx context.Context) Runner {
	r := &worker{
		queue: map[Task]struct{}{},
		tasks: map[Task]struct{}{},

		stop_done: make(chan struct{}),
	}
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
	if !r.started {
		r.queue[task] = struct{}{}
		return nil
	}

	r.tasks[task] = struct{}{}
	r.run(task, r.run_ctx)

	return nil
}

func (r *worker) start(ctx context.Context) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.run_ctx = ctx
	r.started = true

	for t := range r.queue {
		r.run(t, ctx)
	}

	r.tasks = r.queue
	r.queue = nil
}

func (r *worker) run(t Task, ctx context.Context) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		t.Run(ctx)

		r.mutex.Lock()
		defer r.mutex.Unlock()
		if r.stopped {
			return
		}

		Halt(t)
		delete(r.tasks, t)
	}()
}

func (r *worker) Run(ctx context.Context) error {
	r.start(ctx)
	<-r.ctx.Done()
	return nil
}

func (r *worker) stop() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	last := r.stopped

	r.stopped = true
	r.queue = nil

	return last
}

func (r *worker) stopTasks() {
	defer r.cancel()
	defer close(r.stop_done)

	ctx := context.Background()

	for t := range r.tasks {
		t.Stop(ctx)
	}
}

func (r *worker) Stop(ctx context.Context) error {
	if !r.stop() {
		go r.stopTasks()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.stop_done:
		return nil
	}
}

func (r *worker) Close() error {
	r.stop()
	defer r.cancel()
	defer close(r.stop_done)

	for t := range r.tasks {
		t.Close()
	}

	return nil
}

func (r *worker) Wait() error {
	defer r.wg.Wait()

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
