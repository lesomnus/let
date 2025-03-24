package let_test

import (
	"context"
	"sync"
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("runs are exclusive", func(t *testing.T) {
		const N = 10000
		var wg sync.WaitGroup
		wg.Add(N)

		i := 0
		task := let.New(func(ctx context.Context) error {
			i++
			wg.Done()
			return nil
		})
		defer let.Halt(task)

		for range N {
			go task.Run(t.Context())
		}
		wg.Wait()

		require.Equal(t, N, i)
	})
	t.Run("run can be canceled before its start", func(t *testing.T) {
		c := make(chan struct{})
		task := let.New(func(ctx context.Context) error {
			for {
				select {
				case c <- struct{}{}:
				case <-ctx.Done():
					return nil
				}
			}
		})
		defer let.Halt(task)

		go task.Run(t.Context())

		<-c // Ensure the select statement hit.

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := task.Run(ctx)
		require.ErrorIs(t, err, ctx.Err())
	})
	t.Run("context is canceled on the task stop", func(t *testing.T) {
		c := make(chan struct{})
		task := let.New(func(ctx context.Context) error {
			for {
				select {
				case c <- struct{}{}:
				case <-ctx.Done():
					return nil
				}
			}
		})
		defer let.Halt(task)

		go func() {
			<-c // Ensure the select statement hit.
			task.Stop(t.Context())
		}()

		err := task.Run(t.Context())
		require.NoError(t, err)
	})
	t.Run("context is canceled on the task close", func(t *testing.T) {
		c := make(chan struct{})
		task := let.New(func(ctx context.Context) error {
			for {
				select {
				case c <- struct{}{}:
				case <-ctx.Done():
					return nil
				}
			}
		})
		defer let.Halt(task)

		go func() {
			<-c // Ensure the select statement hit.
			task.Close()
		}()

		err := task.Run(t.Context())
		require.NoError(t, err)
	})
	t.Run("stopped task does not run", func(t *testing.T) {
		i := 0
		task := let.New(func(ctx context.Context) error {
			i++
			return nil
		})
		defer let.Halt(task)

		task.Stop(t.Context())
		err := task.Run(t.Context())
		require.Equal(t, let.ErrClosed, err)
		require.Equal(t, 0, i)
	})
	t.Run("closed task does not run", func(t *testing.T) {
		i := 0
		task := let.New(func(ctx context.Context) error {
			i++
			return nil
		})
		defer let.Halt(task)

		task.Close()
		err := task.Run(t.Context())
		require.Equal(t, let.ErrClosed, err)
		require.Equal(t, 0, i)
	})
}
