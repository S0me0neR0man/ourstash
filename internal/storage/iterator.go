package storage

// Iterator holding the iterator's state
type Iterator struct {
	tree        *redBlackTree
	currentNode *redBlackNode
	pos         position
}

type position byte

const (
	begin, onmyway, end position = 0, 1, 2
)

// Iterator returns a iterator
func (t *redBlackTree) Iterator() Iterator {
	return Iterator{tree: t, currentNode: nil, pos: begin}
}

// IteratorAt returns a iterator
func (t *redBlackTree) IteratorAt(node *redBlackNode) Iterator {
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
// Does not modify the state of the iterator.
func (it *Iterator) Key() Key {
	return it.currentNode.key
}

// redBlackNode returns the current element's currentNode.
// Does not modify the state of the iterator.
func (it *Iterator) Node() *redBlackNode {
	return it.currentNode
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (it *Iterator) Begin() {
	it.currentNode = nil
	it.pos = begin
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (it *Iterator) End() {
	it.currentNode = nil
	it.pos = end
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator
func (it *Iterator) First() bool {
	it.Begin()
	return it.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (it *Iterator) Last() bool {
	it.End()
	return it.Prev()
}

// NextTo moves the iterator to the next element from current pos that satisfies the condition given by the
// passed function, and returns true if there was a next element in the container.
// If NextTo() returns true, then next element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (it *Iterator) NextTo(f func(key interface{}, value interface{}) bool) bool {
	for it.Next() {
		key, value := it.Key(), it.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}

// PrevTo moves the iterator to the previous element from current pos that satisfies the condition given by the
// passed function, and returns true if there was a next element in the container.
// If PrevTo() returns true, then next element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (it *Iterator) PrevTo(f func(key interface{}, value interface{}) bool) bool {
	for it.Prev() {
		key, value := it.Key(), it.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}
