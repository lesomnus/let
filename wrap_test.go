package let_test

import (
	"context"
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	t.Run("run is intercepted", func(t *testing.T) {
		v := []string{}
		c := make(chan struct{})
		task := let.Wrap(
			let.New(func(ctx context.Context) error {
				v = append(v, "bar")
				return nil
			}),
			func(ctx context.Context, next func(ctx context.Context) error) error {
				v = append(v, "foo")
				next(ctx)
				v = append(v, "baz")
				<-c
				return nil
			},
		)
		defer let.Halt(task)
		go task.Run(t.Context())

		// Ensure the value is set.
		c <- struct{}{}

		require.Equal(t, []string{"foo", "bar", "baz"}, v)
	})
	t.Run("stop prevents the next run", func(t *testing.T) {
		v := 0
		task := let.Wrap(
			let.Nop(),
			func(ctx context.Context, next func(ctx context.Context) error) error {
				v++
				return nil
			},
		)
		defer let.Halt(task)

		task.Run(t.Context())
		task.Stop(t.Context())
		task.Run(t.Context())
		require.Equal(t, 1, v)
	})
	t.Run("stop is propagated", func(t *testing.T) {
		v := ""
		c := make(chan struct{})
		task := let.Wrap(
			let.New(func(ctx context.Context) error {
				<-c
				<-ctx.Done()
				v = "foo"
				<-c
				return nil
			}),
			func(ctx context.Context, next func(ctx context.Context) error) error {
				return next(ctx)
			},
		)
		defer let.Halt(task)
		go task.Run(t.Context())

		// Ensure the run is started.
		c <- struct{}{}

		task.Close()

		// Ensure the value is set.
		c <- struct{}{}

		require.Equal(t, "foo", v)
	})
}
