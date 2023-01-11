package stashdb

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	once   sync.Once
	logger *zap.Logger
)

func getTestLogger() *zap.Logger {
	once.Do(func() {
		var err error
		logger, err = zap.NewProduction() // or NewProduction, or NewDevelopment,
		if err != nil {
			log.Fatal(err)
		}
	})

	return logger
}

func Test_redBlackTree_Put(t1 *testing.T) {
	sugar := getTestLogger().Sugar()

	tree := newRedBlackTree()
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.sizeof(), 0, "not empty")

	tree.put(NewKey(0, 1, 0))
	tree.put(NewKey(0, 2, 0))
	tree.put(NewKey(0, 1, 0))
	tree.put(NewKey(0, 3, 0))
	tree.put(NewKey(0, 4, 0))
	tree.put(NewKey(0, 5, 0))
	tree.put(NewKey(0, 6, 0))

	sugar.Debugln(tree)

	// redBlackTree
	// │           ┌── R 00 0000000000000006 0000
	// │       ┌── B 00 0000000000000005 0000
	// │   ┌── R 00 0000000000000004 0000
	// │   │   └── B 00 0000000000000003 0000
	// └── B 00 0000000000000002 0000
	//     └── B 00 0000000000000001 0000

	require.EqualValues(t1, 6, tree.sizeof(), "wrong size`")
	require.EqualValues(t1, 4, tree.get(NewKey(0, 4, 0)).sizeof(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.get(NewKey(0, 2, 0)).sizeof(tree), "wrong size`")
	require.EqualValues(t1, 0, tree.get(NewKey(0, 8, 0)).sizeof(tree), "wrong size`")

	key := NewKey(0, 3, 0)
	node := tree.get(key)
	require.NotNil(t1, node)
	require.Equal(t1, key, node.key)
}

func TestRedBlackTree_Remove(t1 *testing.T) {
	sugar := getTestLogger().Sugar()

	tree := newRedBlackTree()
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.sizeof(), 0, "not empty")

	tree.put(NewKey(0, 10, 0))
	tree.put(NewKey(0, 9, 0))
	tree.put(NewKey(0, 8, 0))
	tree.put(NewKey(0, 7, 0))
	tree.put(NewKey(0, 1, 0))
	tree.put(NewKey(0, 2, 0))
	tree.put(NewKey(0, 3, 0))
	tree.put(NewKey(0, 1, 0))
	tree.put(NewKey(0, 7, 0))
	tree.put(NewKey(0, 4, 0))
	tree.put(NewKey(0, 5, 0))
	tree.put(NewKey(0, 6, 0))

	sugar.Debugln(tree)

	// redBlackTree
	// │       ┌── B 00 000000000000000a 0000
	// │   ┌── B 00 0000000000000009 0000
	// │   │   └── B 00 0000000000000008 0000
	// └── B 00 0000000000000007 0000
	//     │           ┌── R 00 0000000000000006 0000
	//     │       ┌── B 00 0000000000000005 0000
	//     │   ┌── R 00 0000000000000004 0000
	//     │   │   └── B 00 0000000000000003 0000
	//     └── B 00 0000000000000002 0000
	//         └── B 00 0000000000000001 0000

	require.EqualValues(t1, 10, tree.sizeof(), "wrong size`")
	require.EqualValues(t1, 3, tree.get(NewKey(0, 9, 0)).sizeof(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.get(NewKey(0, 2, 0)).sizeof(tree), "wrong size`")

	tree.remove(NewKey(0, 10, 0))
	tree.remove(NewKey(0, 9, 0))
	tree.remove(NewKey(0, 8, 0))
	tree.remove(NewKey(0, 7, 0))
	tree.remove(NewKey(0, 9, 0))
	tree.remove(NewKey(0, 8, 0))

	sugar.Debugln(tree)

	// redBlackTree
	// │   ┌── B 00 0000000000000006 0000
	// │   │   └── R 00 0000000000000005 0000
	// └── B 00 0000000000000004 0000
	//     │   ┌── B 00 0000000000000003 0000
	//     └── R 00 0000000000000002 0000
	//         └── B 00 0000000000000001 0000

	require.EqualValues(t1, 6, tree.sizeof(), "wrong size`")
	require.EqualValues(t1, 3, tree.get(NewKey(0, 2, 0)).sizeof(tree), "wrong size`")
	require.EqualValues(t1, 2, tree.get(NewKey(0, 6, 0)).sizeof(tree), "wrong size`")
}
