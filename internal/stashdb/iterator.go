package stashdb

// iterator holding the iterators state
type iterator struct {
	tree *redBlackTree
	node *redBlackNode
	pos  position
}

type position byte

const (
	begin, onmyway, end position = 0, 1, 2
)

// iterator returns an iterator
//
// IMPORTANT: iterator does not provide thread safety
func (t *redBlackTree) iterator() iterator {
	return iterator{tree: t, node: nil, pos: begin}
}

// iteratorAt returns an iterator at node
//
// IMPORTANT: iterator does not provide thread safety
func (t *redBlackTree) iteratorAt(node *redBlackNode) iterator {
	if node == nil {
		return iterator{tree: t, node: nil, pos: begin}
	}
	return iterator{tree: t, node: node, pos: onmyway}
}

// next moves the iterator to the next element
func (it *iterator) next() bool {
	if it.pos == end {
		it.node = nil
		return false
	}

	if it.pos == begin {
		minNode := it.min()
		if minNode == nil {
			it.node = nil
			it.pos = end
			return false
		}
		it.node = minNode
		it.pos = onmyway
		return true
	}

	if it.node.right != nil {
		it.node = it.node.right
		for it.node.left != nil {
			it.node = it.node.left
		}
		it.pos = onmyway
		return true
	}

	for it.node.parent != nil {
		node := it.node
		it.node = it.node.parent
		if node == it.node.left {
			it.pos = onmyway
			return true
		}
	}

	it.pos = end
	it.node = nil
	return false
}

// prev moves the iterator to the previous element
func (it *iterator) prev() bool {
	if it.pos == begin {
		it.node = nil
		return false
	}

	if it.pos == end {
		maxNode := it.max()
		if maxNode == nil {
			it.node = nil
			it.pos = begin
			return false
		}
		it.node = maxNode
		it.pos = onmyway
		return true
	}

	if it.node.left != nil {
		it.node = it.node.left
		for it.node.right != nil {
			it.node = it.node.right
		}
		it.pos = onmyway
		return true
	}

	for it.node.parent != nil {
		curNode := it.node
		it.node = it.node.parent
		if curNode == it.node.right {
			it.pos = onmyway
			return true
		}
	}

	it.node = nil
	it.pos = begin
	return false
}

// begin resets the iterator to one-before-first
func (it *iterator) begin() {
	it.node = nil
	it.pos = begin
}

// end moves the iterator to one-past-the-end
func (it *iterator) end() {
	it.node = nil
	it.pos = end
}

// min returns the minimal current or nil
func (it *iterator) min() *redBlackNode {
	var minNode *redBlackNode
	for curNode := it.tree.root; curNode != nil; curNode = curNode.left {
		minNode = curNode
	}
	return minNode
}

// max returns the max current or nil
func (it *iterator) max() *redBlackNode {
	var maxNode *redBlackNode
	for curNode := it.tree.root; curNode != nil; curNode = curNode.right {
		maxNode = curNode
	}
	return maxNode
}
