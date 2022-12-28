package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_iterator_next(t *testing.T) {
	tree := newRedBlackTree()
	require.NotNil(t, tree)
	require.EqualValues(t, tree.sizeof(), 0, "not empty")

	tree.put(NewKey(1, 0, 0))
	it := tree.iterator()
	require.EqualValues(t, it.pos, begin)
	require.EqualValues(t, true, it.node == nil)

	flag := it.next()
	require.EqualValues(t, it.pos, onmyway)
	require.EqualValues(t, true, flag)
	require.EqualValues(t, true, it.node != nil)

	flag = it.next()
	require.EqualValues(t, it.pos, end)
	require.EqualValues(t, false, flag)
	require.EqualValues(t, true, it.node == nil)

	it.begin()
	require.EqualValues(t, it.pos, begin)
	require.EqualValues(t, true, it.node == nil)

	flag = it.next()
	require.EqualValues(t, it.pos, onmyway)
	require.EqualValues(t, true, flag)
	require.EqualValues(t, true, it.node != nil)

	flag = it.prev()
	require.EqualValues(t, it.pos, begin)
	require.EqualValues(t, false, flag)
	require.EqualValues(t, true, it.node == nil)

}
