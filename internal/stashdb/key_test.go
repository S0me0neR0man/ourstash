package stashdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKey_Compare(t *testing.T) {
	k0 := NewKey(0, 0, 0)
	k1 := NewKey(0, 0, 1)
	k2 := NewKey(0, 1, 0)
	k3 := NewKey(0, 1, 1)
	k4 := NewKey(1, 0, 0)
	k5 := NewKey(1, 0, 1)
	k6 := NewKey(1, 1, 0)

	require.Equal(t, true, k0.Compare(k1) == KeyLessThan)
	require.Equal(t, true, k1.Compare(k2) == KeyLessThan)
	require.Equal(t, true, k2.Compare(k3) == KeyLessThan)
	require.Equal(t, true, k3.Compare(k4) == KeyLessThan)
	require.Equal(t, true, k4.Compare(k5) == KeyLessThan)
	require.Equal(t, true, k5.Compare(k6) == KeyLessThan)
	require.Equal(t, true, k6.Compare(k1) == KeyMoreThan)
	require.Equal(t, true, k4.Compare(k0) == KeyMoreThan)
	require.Equal(t, true, k4.Compare(k4) == KeyEqual)
}
