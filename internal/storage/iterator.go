package storage

import "fmt"

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
	t.sugar.Infow("Iterator .", "t", fmt.Sprintf("%p", t))
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.sugar.Infow("Iterator R", "t", fmt.Sprintf("%p", t))

	return Iterator{tree: t, currentNode: nil, pos: begin}
}

// Next moves the iterator to the next element
func (it *Iterator) Next() bool {
	it.tree.sugar.Infow("Next .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("Next R", "it", fmt.Sprintf("%p", it))

	if it.pos == end {
		it.currentNode = nil
		return false
	}

	if it.pos == begin {
		minNode := it.min()
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
	it.tree.sugar.Infow("Prev .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("Prev R", "it", fmt.Sprintf("%p", it))

	if it.pos == begin {
		it.currentNode = nil
		return false
	}

	if it.pos == end {
		maxNode := it.max()
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
	it.tree.sugar.Infow("Key .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("Key R", "it", fmt.Sprintf("%p", it))

	return it.currentNode.key
}

// Node returns the current element's currentNode.
func (it *Iterator) Node() *redBlackNode {
	it.tree.sugar.Infow("Node .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("Node R", "it", fmt.Sprintf("%p", it))

	return it.currentNode
}

// Begin resets the iterator to one-before-first
func (it *Iterator) Begin() {
	it.tree.sugar.Infow("Begin .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("Begin R", "it", fmt.Sprintf("%p", it))

	it.currentNode = nil
	it.pos = begin
}

// End moves the iterator to one-past-the-end
func (it *Iterator) End() {
	it.tree.sugar.Infow("End .", "it", fmt.Sprintf("%p", it))
	it.tree.mu.RLock()
	defer it.tree.mu.RUnlock()
	it.tree.sugar.Infow("End R", "it", fmt.Sprintf("%p", it))

	it.currentNode = nil
	it.pos = end
}

// min returns the minimal node or nil
func (it *Iterator) min() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := it.tree.root; curNode != nil; curNode = curNode.left {
		parentNode = curNode
	}
	return parentNode
}

// max returns the max node or nil
func (it *Iterator) max() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := it.tree.root; curNode != nil; curNode = curNode.right {
		parentNode = curNode
	}
	return parentNode
}
