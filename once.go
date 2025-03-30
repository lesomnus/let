package let

import (
	"context"
	"sync/atomic"
)

type once struct {
	Task

	s atomic.Bool
}

func Once(t Task) Task {
	return &once{Task: t}
}

func (t *once) Run(ctx context.Context) error {
	if t.s.Swap(true) {
		return ErrClosed
	}

	defer Halt(t.Task)
	return t.Task.Run(ctx)
}
