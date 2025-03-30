package let

import (
	"context"
	"sync"
)

type held struct {
	mutex sync.Mutex

	started bool
	stopped bool

	base Runner
	ts   []Task
}

// Hold defers the run of the [Task] given by [Runner.Go] until returned [Runner.Run].
// If [Stop] or [Close] is called before [Run], deferred Tasks are remain untouched.
func Hold(r Runner) Runner {
	return &held{
		base: r,
		ts:   []Task{},
	}
}

func (r *held) Go(t Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.stopped {
		return ErrClosed
	}
	if r.started {
		return r.Go(t)
	}

	r.ts = append(r.ts, t)
	return nil
}

func (r *held) start() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.started = true

	for _, t := range r.ts {
		r.base.Go(t)
	}

	r.ts = nil
}

func (r *held) Run(ctx context.Context) error {
	r.start()
	return r.base.Run(ctx)
}

func (r *held) stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.stopped = true
	r.ts = nil
}

func (r *held) Stop(ctx context.Context) error {
	r.stop()
	return r.base.Stop(ctx)
}

func (r *held) Close() error {
	r.stop()
	return r.base.Close()
}

func (r *held) Wait() error {
	return r.base.Wait()
}
