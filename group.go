package let

import (
	"context"
	"sync"
)

type group struct {
	*runner

	wg  sync.WaitGroup
	err error
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// If a task fails, all its child tasks are stopped.
// The [Stop], [Close], and [Wait] methods return the error from the first failed task.
// Unlike [NewWithContext], cancel of the given context does not
// result Stop of all the child Tasks.
func NewGroupWithContext(ctx context.Context) Runner {
	r := &group{runner: newRunner(ctx)}
	r.invoke = r.run

	return r
}

// NewRunner creates a Runner that runs multiple tasks simultaneously.
// If a task fails, all its child tasks are stopped.
// The [Stop], [Close], and [Wait] methods return the error from the first failed task.
func NewGroup() Runner {
	return NewGroupWithContext(context.Background())
}

func (r *group) run(t Task, ctx context.Context) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		err := t.Run(ctx)
		if r.stop() {
			// This is not the first return of the Run
			// or the Run is stopped by context cancel.
			return
		}

		r.err = err
		r.stopTasks()
	}()
}

func (r *group) Wait() error {
	defer r.wg.Wait()

	<-r.ctx.Done()

	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, t := range r.tasks {
		t.Wait()
	}

	return r.err
}
