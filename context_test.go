package let_test

import (
	"context"
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestWithinContext(t *testing.T) {
	t.Run("context is injected", func(t *testing.T) {
		type k string
		ctx := context.WithValue(t.Context(), k("foo"), "bar")

		v := ""
		task := let.NewWithinContext(ctx, func(ctx context.Context) error {
			v, _ = ctx.Value(k("foo")).(string)
			return nil
		})
		defer let.Halt(task)

		task.Run(t.Context())
		require.Equal(t, "bar", v)
	})
	t.Run("stop causes cancel of injected context", func(t *testing.T) {
		ctx := t.Context()

		v := ""
		c := make(chan struct{})
		task := let.NewWithinContext(ctx, func(ctx context.Context) error {
			<-c
			<-ctx.Done()
			v = "foo"
			<-c
			return nil
		})
		defer let.Halt(task)

		go task.Run(ctx)

		// Ensure the task run started.
		c <- struct{}{}

		task.Close()

		// Ensure the value is set.
		c <- struct{}{}

		require.Equal(t, "foo", v)
	})
}
