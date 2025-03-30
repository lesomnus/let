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
		go r.Run(t.Context())

		r.Go(let.New(func(ctx context.Context) error {
			<-c
			return io.EOF
		}))
		r.Go(let.New(func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}))

		c <- struct{}{}

		err := r.Wait()
		require.Equal(t, io.EOF, err)
	})
}
