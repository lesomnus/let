package let_test

import (
	"context"
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestSeq(t *testing.T) {
	t.Run("runs task sequentially", func(t *testing.T) {
		c := make(chan int)
		task := let.Seq(
			let.New(func(ctx context.Context) error {
				c <- 1
				return nil
			}),
			let.New(func(ctx context.Context) error {
				c <- 2
				return nil
			}),
			let.New(func(ctx context.Context) error {
				c <- 3
				return nil
			}),
		)
		defer let.Halt(task)

		go task.Run(t.Context())

		v := <-c
		require.Equal(t, 1, v)
		v = <-c
		require.Equal(t, 2, v)
		v = <-c
		require.Equal(t, 3, v)
	})
	t.Run("stops in the middle of the sequence", func(t *testing.T) {
		c := make(chan int)
		task := let.Seq(
			let.New(func(ctx context.Context) error {
				c <- 1
				return nil
			}),
			let.New(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}),
			let.New(func(ctx context.Context) error {
				t.Fail()
				return nil
			}),
		)
		defer let.Halt(task)

		go task.Run(t.Context())

		v := <-c
		require.Equal(t, 1, v)

		task.Stop(t.Context())

		go func() {
			select {
			case <-t.Context().Done():
			case c <- 42:
			}
		}()

		v = <-c
		require.Equal(t, 42, v)
	})
}
