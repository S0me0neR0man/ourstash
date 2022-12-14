package storage

// Iterator holding the iterator's state
type Iterator struct {
	tree        *RedBlackTree
	currentNode *redBlackNode
	pos         position
}

type position byte

const (
	begin, onmyway, end position = 0, 1, 2
)

// Iterator returns a iterator
func (t *RedBlackTree) Iterator() Iterator {
	return Iterator{tree: t, currentNode: nil, pos: begin}
}

// IteratorAt returns a iterator
func (t *RedBlackTree) IteratorAt(node *redBlackNode) Iterator {
	return Iterator{tree: t, currentNode: node, pos: onmyway}
}

// Next moves the iterator to the next element
func (it *Iterator) Next() bool {
	if it.pos == end {
		it.currentNode = nil
		return false
	}

	if it.pos == begin {
		minNode := it.tree.Left()
		if minNode == nil {
			it.currentNode = nil
			it.pos = end
			return false
		}
		it.currentNode = minNode
		it.pos = onmyway
		return true
	}

	if it.currentNode.right != nil {
		it.currentNode = it.currentNode.right
		for it.currentNode.left != nil {
			it.currentNode = it.currentNode.left
		}
		it.pos = onmyway
		return true
	}

	for it.currentNode.parent != nil {
		node := it.currentNode
		it.currentNode = it.currentNode.parent
		if node == it.currentNode.left {
			break
		}
	}

	it.pos = onmyway
	return true
}

// Prev moves the iterator to the previous element
func (it *Iterator) Prev() bool {
	if it.pos == begin {
		it.currentNode = nil
		return false
	}

	if it.pos == end {
		maxNode := it.tree.Right()
		if maxNode == nil {
			it.currentNode = nil
			it.pos = begin
			return false
		}
		it.currentNode = maxNode
		it.pos = onmyway
		return true
	}

	if it.currentNode.left != nil {
		it.currentNode = it.currentNode.left
		for it.currentNode.right != nil {
			it.currentNode = it.currentNode.right
		}
		it.pos = onmyway
		return true
	}

	for it.currentNode.parent != nil {
		curNode := it.currentNode
		it.currentNode = it.currentNode.parent
		if curNode == it.currentNode.right {
			break
		}
	}
	it.pos = onmyway
	return true
}

// Key returns the current element's key.
func (it *Iterator) Key() Key {
	return it.currentNode.key
}

// Node returns the current element's currentNode.
func (it *Iterator) Node() *redBlackNode {
	return it.currentNode
}

// Begin resets the iterator to one-before-first
func (it *Iterator) Begin() {
	it.currentNode = nil
	it.pos = begin
}

// End moves the iterator to one-past-the-end
func (it *Iterator) End() {
	it.currentNode = nil
	it.pos = end
}

// First moves the iterator to the first element
func (it *Iterator) First() bool {
	it.Begin()
	return it.Next()
}

// Last moves the iterator to the last element
func (it *Iterator) Last() bool {
	it.End()
	return it.Prev()
}