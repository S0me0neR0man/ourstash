package storage

type BNode struct {
	left  *BNode
	right *BNode
	data  int64
}

type BTree struct {
	root *BNode
}
