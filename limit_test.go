package let_test

import (
	"testing"

	"github.com/lesomnus/let"
	"github.com/stretchr/testify/require"
)

func TestLimit(t *testing.T) {
	t.Run("run only N times", func(t *testing.T) {
		task := let.Limit(3, let.Nop())

		err := task.Run(t.Context())
		require.NoError(t, err)

		err = task.Run(t.Context())
		require.NoError(t, err)

		err = task.Run(t.Context())
		require.NoError(t, err)

		err = task.Run(t.Context())
		require.ErrorIs(t, err, let.ErrClosed)
	})
	t.Run("stops before the limit", func(t *testing.T) {
		task := let.Limit(3, let.Nop())

		err := task.Run(t.Context())
		require.NoError(t, err)

		err = task.Stop(t.Context())
		require.NoError(t, err)

		err = task.Run(t.Context())
		require.ErrorIs(t, err, let.ErrClosed)
	})
}
