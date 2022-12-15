package storage

import (
	"fmt"
	"log"
	"sync"
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

	// RedBlackTree
	// │           ┌── R 00 0000000000000006 0000 0000
	// │       ┌── B 00 0000000000000005 0000 0000
	// │   ┌── R 00 0000000000000004 0000 0000
	// │   │   └── B 00 0000000000000003 0000 0000
	// └── B 00 0000000000000002 0000 0000
	//     └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 6, tree.Size(), "wrong size`")
	require.EqualValues(t1, 4, tree.GetNode(NewKey(0, 4, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 0, tree.GetNode(NewKey(0, 8, 0, 0)).Size(tree), "wrong size`")
}

func TestRedBlackTree_Remove(t1 *testing.T) {
	tree := NewRedBlackTree()
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	tree.Put(NewKey(0, 10, 0, 0))
	tree.Put(NewKey(0, 9, 0, 0))
	tree.Put(NewKey(0, 8, 0, 0))
	tree.Put(NewKey(0, 7, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 2, 0, 0))
	tree.Put(NewKey(0, 3, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 7, 0, 0))
	tree.Put(NewKey(0, 4, 0, 0))
	tree.Put(NewKey(0, 5, 0, 0))
	tree.Put(NewKey(0, 6, 0, 0))

	fmt.Println(tree)

	// RedBlackTree
	// │       ┌── B 00 000000000000000a 0000 0000
	// │   ┌── B 00 0000000000000009 0000 0000
	// │   │   └── B 00 0000000000000008 0000 0000
	// └── B 00 0000000000000007 0000 0000
	//     │           ┌── R 00 0000000000000006 0000 0000
	//     │       ┌── B 00 0000000000000005 0000 0000
	//     │   ┌── R 00 0000000000000004 0000 0000
	//     │   │   └── B 00 0000000000000003 0000 0000
	//     └── B 00 0000000000000002 0000 0000
	//         └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 10, tree.Size(), "wrong size`")
	require.EqualValues(t1, 3, tree.GetNode(NewKey(0, 9, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")

	tree.Remove(NewKey(0, 10, 0, 0))
	tree.Remove(NewKey(0, 9, 0, 0))
	tree.Remove(NewKey(0, 8, 0, 0))
	tree.Remove(NewKey(0, 7, 0, 0))
	tree.Remove(NewKey(0, 9, 0, 0))
	tree.Remove(NewKey(0, 8, 0, 0))

	fmt.Println(tree)

	// RedBlackTree
	// │   ┌── B 00 0000000000000006 0000 0000
	// │   │   └── R 00 0000000000000005 0000 0000
	// └── B 00 0000000000000004 0000 0000
	//     │   ┌── B 00 0000000000000003 0000 0000
	//     └── R 00 0000000000000002 0000 0000
	//         └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 6, tree.Size(), "wrong size`")
	require.EqualValues(t1, 3, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 2, tree.GetNode(NewKey(0, 6, 0, 0)).Size(tree), "wrong size`")
}

func TestRedBlackTree_inGoroutines(t1 *testing.T) {
	tree := NewRedBlackTree()
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	goroutinesCount := 5
	var wg sync.WaitGroup
	wg.Add(goroutinesCount *2)

	func1 := func() {
		tree.Put(NewKey(0, 10, 0, 0))
		tree.Put(NewKey(0, 9, 0, 0))
		tree.Put(NewKey(0, 8, 0, 0))
		tree.Put(NewKey(0, 7, 0, 0))
		tree.Put(NewKey(0, 1, 0, 0))
		tree.Put(NewKey(0, 2, 0, 0))
		tree.Put(NewKey(0, 3, 0, 0))
		tree.Put(NewKey(0, 1, 0, 0))
		tree.Put(NewKey(0, 7, 0, 0))
		tree.Put(NewKey(0, 4, 0, 0))
		tree.Put(NewKey(0, 5, 0, 0))
		tree.Put(NewKey(0, 6, 0, 0))
		wg.Done()
		log.Println("func1 done")
	}
	log.Printf("-- %p\n", func1)

	func2 := func() {
		tree.Remove(NewKey(0, 10, 0, 0))
		tree.Remove(NewKey(0, 9, 0, 0))
		tree.Remove(NewKey(0, 8, 0, 0))
		tree.Remove(NewKey(0, 7, 0, 0))
		tree.Remove(NewKey(0, 9, 0, 0))
		tree.Remove(NewKey(0, 8, 0, 0))
		wg.Done()
		log.Println("func2 done")
	}
	log.Printf("-- %p\n", func2)

	func3 := func() {
		it := tree.Iterator()
		log.Println("func3", it)
		it.First()
		//for res := it.First(); res; it.Next() {}
		wg.Done()
		log.Println("func3 done")
	}
	log.Printf("-- %p\n", func3)

	func4 := func() {
		it := tree.Iterator()
		log.Println("func4", it)
		for res := it.Last(); res; it.Prev() {}
		wg.Done()
		log.Println("func4 done")
	}
	log.Printf("-- %p\n", func4)

	log.Println("-- start")
	// todo: deadlock bug
	for i := 0; i < goroutinesCount; i++ {
		go func1()
		//go func2()
		go func3()
		//go func4()
	}

	wg.Wait()
}
