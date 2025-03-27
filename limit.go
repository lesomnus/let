package let

import (
	"context"
	"sync"
)

type limited struct {
	Task

	m sync.Mutex
	n int
	l int
}

// Limit creates a new Task that limits the number of times it can be Run.
// Once the limit is reached, the Task will return `ErrClosed` on subsequent Run.
func Limit(n int, t Task) Task {
	if n < 1 {
		panic("n must be larger than 0")
	}
	return &limited{Task: t, l: n}
}

func (t *limited) touch() int {
	t.m.Lock()
	defer t.m.Unlock()
	n := t.n
	if t.n >= t.l {
		return t.l
	}

	t.n++
	return n
}

func (t *limited) Run(ctx context.Context) error {
	n := t.touch()
	if n >= t.l {
		return ErrClosed
	}
	if n+1 == t.l {
		defer t.Stop(ctx)
	}
	return t.Task.Run(ctx)
}
