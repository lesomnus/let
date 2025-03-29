package let

import "context"

type withContext struct {
	ctx context.Context
	Task
}

// OverrideContext replaces the context with the given `ctx` when `Task.Run` is called.
// Canceling the context passed to `Task.Run` will also cancel the given `ctx`.
// This is useful when you want to use your context while passing the Task to a Runner.
// Keep in mind that you loose access to the caller's context.
func OverrideContext(ctx context.Context, t Task) Task {
	return withContext{ctx, t}
}

func (t withContext) Run(ctx context.Context) error {
	ctx_base, cancel := context.WithCancel(t.ctx)
	defer cancel()

	stop := context.AfterFunc(ctx, cancel)
	defer stop()

	return t.Task.Run(ctx_base)
}
