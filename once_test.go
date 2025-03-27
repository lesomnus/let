package let_test

import (
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestOnce(t *testing.T) {
	t.Run("run only once", func(t *testing.T) {
		task := let.Once(let.Nop())

		err := task.Run(t.Context())
		require.NoError(t, err)

		err = task.Run(t.Context())
		require.ErrorIs(t, err, let.ErrClosed)
	})
}
