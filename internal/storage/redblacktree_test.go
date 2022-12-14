package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_redBlackTree_Put(t1 *testing.T) {
	tree := NewRedBlackTree()
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 2, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 3, 0, 0))
	tree.Put(NewKey(0, 4, 0, 0))
	tree.Put(NewKey(0, 5, 0, 0))
	tree.Put(NewKey(0, 6, 0, 0))

	fmt.Println(tree)
	//
	//  RedBlackTree
	//  │           ┌── 6
	//  │       ┌── 5
	//  │   ┌── 4
	//  │   │   └── 3
	//  └── 2
	//       └── 1

	require.EqualValues(t1, 6, tree.Size(), "wrong size`")
	require.EqualValues(t1, 4, tree.GetNode(NewKey(0, 4, 0, 0)).Size(), "wrong size`")
	require.EqualValues(t1, 6, tree.GetNode(NewKey(0, 2, 0, 0)).Size(), "wrong size`")
	require.EqualValues(t1, 0, tree.GetNode(NewKey(0, 8, 0, 0)).Size(), "wrong size`")
}
