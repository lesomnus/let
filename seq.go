package let

import (
	"context"
	"errors"
	"slices"
)

type seq struct {
	root  Task
	tasks []Task
}

// Seq creates a Task that runs the given Tasks sequentially.
// If any step returns an error and the error is returned
// immediately without running rest of the steps.
// If ErrClosed is returned, it returns nil instead.
func Seq(ts ...Task) Task {
	t := New(func(ctx context.Context) error {
		for _, t := range ts {
			err := t.Run(ctx)
			if err == nil {
				continue
			}
			if err == ErrClosed {
				// The step was closed, so stop the sequence.
				return nil
			}

			return err
		}
		return nil
	})
	return &seq{t, ts}
}

func (t *seq) Run(ctx context.Context) error {
	return t.root.Run(ctx)
}

func (t *seq) Stop(ctx context.Context) error {
	errs := make([]error, 0, len(t.tasks))

	// Children must be stopped gracefully so stop the root last
	// to prevent cancel of the root context.
	// To prevent the run of the next step after a step finishes, stop it backward.
	for _, v := range slices.Backward(t.tasks) {
		errs = append(errs, v.Stop(ctx))
	}

	t.root.Stop(ctx)
	return errors.Join(errs...)
}

func (t *seq) Close() error {
	errs := make([]error, 0, len(t.tasks))

	// Children must be closed before the close of the root
	// to get their error not the error caused by root context.
	// To prevent the run of the next step after a step finishes, stop it backward.
	for _, v := range slices.Backward(t.tasks) {
		errs = append(errs, v.Close())
	}

	t.root.Close()
	return errors.Join(errs...)
}

func (t *seq) Wait() error {
	errs := make([]error, 0, len(t.tasks))
	for _, v := range slices.Backward(t.tasks) {
		errs = append(errs, v.Wait())
	}

	t.root.Wait()
	return errors.Join(errs...)
}
