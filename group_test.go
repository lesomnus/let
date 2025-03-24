package let_test

import (
	"context"
	"io"
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) {
	t.Run("stop on first error return", func(t *testing.T) {
		c := make(chan struct{})

		r := let.NewGroup()
		r.Go(let.New(func(ctx context.Context) error {
			<-c
			return io.EOF
		}))
		r.Go(let.New(func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}))

		close(c)

		err := r.Wait()
		require.Equal(t, io.EOF, err)
	})
}
