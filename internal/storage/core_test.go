package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	store, err := NewStorage()
	require.NoError(t, err)
	require.NotNil(t, store)

	chainTo, err := NewPutChain()
	require.NoError(t, err)
	require.NotNil(t, chainTo)
	err = store.New(chainTo)
	require.NoError(t, err)

	//t.Run("new storage", func(t *testing.T) {
	//	}
	//})
}
