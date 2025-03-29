package let

import "context"

// OverrideContext replaces the context with the given `ctx` when `Task.Run` is called.
// Canceling the context passed to `Task.Run` will also cancel the given `ctx`.
// This is useful when you want to use your context while passing the Task to a Runner.
// Keep in mind that you loose access to the caller's context.
func OverrideContext(ctx context.Context, t Task) Task {
	return Wrap(t, func(ctx_caller context.Context, next func(ctx context.Context) error) error {
		ctx_next, cancel := context.WithCancel(ctx)
		defer cancel()

		stop := context.AfterFunc(ctx_caller, cancel)
		defer stop()

		return next(ctx_next)
	})
}
