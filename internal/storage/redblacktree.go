package storage

import (
	"fmt"
	"sync"
)

type color bool

const (
	black, red color = true, false
)

// redBlackTree main index
type redBlackTree struct {
	mu   sync.RWMutex
	root *redBlackNode
	size int
}

// redBlackNode is a tree element
type redBlackNode struct {
	key    Key
	color  color
	left   *redBlackNode
	right  *redBlackNode
	parent *redBlackNode
}

// Put inserts key into the tree.
func (t *redBlackTree) Put(key Key) {
	if t.root == nil {
		t.mu.Lock()
		t.root = &redBlackNode{key: key, color: black}
		t.size++
		t.mu.Unlock()
		return
	}

	curNode := t.root
	for true {
		switch key.Compare(curNode.key) {
		case KeyEqual:
			return
		case KeyLessThan:
			if curNode.left == nil {
				t.mu.Lock()
				curNode.left = &redBlackNode{key: key, color: red}
				t.insertCase1(curNode.left)
				curNode.left.parent = curNode
				t.size++
				t.mu.Unlock()
				return
			}
			curNode = curNode.left
		case KeyMoreThan:
			if curNode.right == nil {
				t.mu.Lock()
				curNode.right = &redBlackNode{key: key, color: red}
				t.insertCase1(curNode.right)
				curNode.right.parent = curNode
				t.size++
				t.mu.Unlock()
				return
			}
			curNode = curNode.right
		}
	}
}

// GetNode searches the currentNode in the tree, nil not found
func (t *redBlackTree) GetNode(key Key) *redBlackNode {
	return t.lookup(key)
}

// Remove the currentNode from the tree
func (t *redBlackTree) Remove(key Key) {
	delNode := t.lookup(key)
	if delNode == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if delNode.left != nil && delNode.right != nil {
		replacementNode := delNode.left.maximumNode()
		delNode.key = replacementNode.key
		delNode = replacementNode
	}

	var childNode *redBlackNode
	if delNode.left == nil || delNode.right == nil {
		if delNode.right == nil {
			childNode = delNode.left
		} else {
			childNode = delNode.right
		}
		if delNode.color == black {
			delNode.color = nodeColor(childNode)
			t.deleteCase1(delNode)
		}
		t.replaceNode(delNode, childNode)
		if delNode.parent == nil && childNode != nil {
			childNode.color = black
		}
	}

	t.size--
}

// IsEmpty returns true if tree is empty
func (t *redBlackTree) IsEmpty() bool {
	return t.size == 0
}

// Size returns number of nodes
func (t *redBlackTree) Size() int {
	return t.size
}

// Size returns the number of elements in the subtree.
func (n *redBlackNode) Size() int {
	if n == nil {
		return 0
	}
	size := 1
	if n.left != nil {
		size += n.left.Size()
	}
	if n.right != nil {
		size += n.right.Size()
	}
	return size
}

// Keys returns all keys in-order
func (t *redBlackTree) Keys() []Key {
	keys := make([]Key, t.size)
	it := t.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// Values returns all values in-order based on the key.
func (t *redBlackTree) Values() []interface{} {
	values := make([]interface{}, t.size)
	it := t.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

// Left returns the minimal node or nil
func (t *redBlackTree) Left() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := t.root; curNode != nil; curNode = curNode.left {
		parentNode = curNode
	}
	return parentNode
}

// Right returns the max node or nil
func (t *redBlackTree) Right() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := t.root; curNode != nil; curNode = curNode.right{
		parentNode = curNode
	}
	return parentNode
}

// Floor Finds floor currentNode of the input key, return the floor currentNode or nil if no floor is found.
// Second return parameter is true if floor was found, otherwise false.
//
// Floor currentNode is defined as the largest currentNode that is smaller than or equal to the given currentNode.
// A floor currentNode may not be found, either because the tree is empty, or because
// all nodes in the tree are larger than the given currentNode.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (t *redBlackTree) Floor(key interface{}) (floor *redBlackNode, found bool) {
	found = false
	node := t.root
	for node != nil {
		compare := t.Comparator(key, node.key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.left
		case compare > 0:
			floor, found = node, true
			node = node.right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// Ceiling finds ceiling currentNode of the input key, return the ceiling currentNode or nil if no ceiling is found.
// Second return parameter is true if ceiling was found, otherwise false.
//
// Ceiling currentNode is defined as the smallest currentNode that is larger than or equal to the given currentNode.
// A ceiling currentNode may not be found, either because the tree is empty, or because
// all nodes in the tree are smaller than the given currentNode.
//
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (t *redBlackTree) Ceiling(key interface{}) (ceiling *redBlackNode, found bool) {
	found = false
	node := t.root
	for node != nil {
		compare := t.Comparator(key, node.key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.left
		case compare > 0:
			node = node.right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (t *redBlackTree) Clear() {
	t.root = nil
	t.size = 0
}

// String returns a string representation of container
func (t *redBlackTree) String() string {
	str := "redBlackTree\n"
	if !t.IsEmpty() {
		output(t.root, "", true, &str)
	}
	return str
}

func (n *redBlackNode) String() string {
	return fmt.Sprintf("%v", n.key)
}

func output(node *redBlackNode, prefix string, isTail bool, str *string) {
	if node.right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.left, newPrefix, true, str)
	}
}

func (t *redBlackTree) lookup(key Key) *redBlackNode {
	curNode := t.root
	for curNode != nil {
		switch curNode.key.Compare(key) {
		case KeyEqual:
			return curNode
		case KeyLessThan:
			curNode = curNode.left
		case KeyMoreThan:
			curNode = curNode.right
		}
	}
	return nil
}

func (n *redBlackNode) grandparent() *redBlackNode {
	if n != nil && n.parent != nil {
		return n.parent.parent
	}
	return nil
}

func (n *redBlackNode) uncle() *redBlackNode {
	if n == nil || n.parent == nil || n.parent.parent == nil {
		return nil
	}
	return n.parent.sibling()
}

func (n *redBlackNode) sibling() *redBlackNode {
	if n == nil || n.parent == nil {
		return nil
	}
	if n == n.parent.left {
		return n.parent.right
	}
	return n.parent.left
}

func (t *redBlackTree) rotateLeft(node *redBlackNode) {
	right := node.right
	t.replaceNode(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
	}
	right.left = node
	node.parent = right
}

func (t *redBlackTree) rotateRight(node *redBlackNode) {
	left := node.left
	t.replaceNode(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
	}
	left.right = node
	node.parent = left
}

func (t *redBlackTree) replaceNode(old *redBlackNode, new *redBlackNode) {
	if old.parent == nil {
		t.root = new
	} else {
		if old == old.parent.left {
			old.parent.left = new
		} else {
			old.parent.right = new
		}
	}
	if new != nil {
		new.parent = old.parent
	}
}

func (t *redBlackTree) insertCase1(node *redBlackNode) {
	if node.parent == nil {
		node.color = black
	} else {
		t.insertCase2(node)
	}
}

func (t *redBlackTree) insertCase2(node *redBlackNode) {
	if nodeColor(node.parent) == black {
		return
	}
	t.insertCase3(node)
}

func (t *redBlackTree) insertCase3(node *redBlackNode) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.parent.color = black
		uncle.color = black
		node.grandparent().color = red
		t.insertCase1(node.grandparent())
	} else {
		t.insertCase4(node)
	}
}

func (t *redBlackTree) insertCase4(node *redBlackNode) {
	grandparent := node.grandparent()
	if node == node.parent.right && node.parent == grandparent.left {
		t.rotateLeft(node.parent)
		node = node.left
	} else if node == node.parent.left && node.parent == grandparent.right {
		t.rotateRight(node.parent)
		node = node.right
	}
	t.insertCase5(node)
}

func (t *redBlackTree) insertCase5(node *redBlackNode) {
	node.parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.parent.left && node.parent == grandparent.left {
		t.rotateRight(grandparent)
	} else if node == node.parent.right && node.parent == grandparent.right {
		t.rotateLeft(grandparent)
	}
}

func (n *redBlackNode) maximumNode() *redBlackNode {
	if n == nil {
		return nil
	}
	var curNode *redBlackNode
	for curNode = n.right; curNode.right != nil; curNode = curNode.right {
	}
	return curNode
}

func (t *redBlackTree) deleteCase1(node *redBlackNode) {
	if node.parent == nil {
		return
	}
	t.deleteCase2(node)
}

func (t *redBlackTree) deleteCase2(node *redBlackNode) {
	sibling := node.sibling()
	if nodeColor(sibling) == red {
		node.parent.color = red
		sibling.color = black
		if node == node.parent.left {
			t.rotateLeft(node.parent)
		} else {
			t.rotateRight(node.parent)
		}
	}
	t.deleteCase3(node)
}

func (t *redBlackTree) deleteCase3(node *redBlackNode) {
	sibling := node.sibling()
	if nodeColor(node.parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == black &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		t.deleteCase1(node.parent)
	} else {
		t.deleteCase4(node)
	}
}

func (t *redBlackTree) deleteCase4(node *redBlackNode) {
	sibling := node.sibling()
	if nodeColor(node.parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == black &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		node.parent.color = black
	} else {
		t.deleteCase5(node)
	}
}

func (t *redBlackTree) deleteCase5(node *redBlackNode) {
	sibling := node.sibling()
	if node == node.parent.left &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == red &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		sibling.left.color = black
		t.rotateRight(sibling)
	} else if node == node.parent.right &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.right) == red &&
		nodeColor(sibling.left) == black {
		sibling.color = red
		sibling.right.color = black
		t.rotateLeft(sibling)
	}
	t.deleteCase6(node)
}

func (t *redBlackTree) deleteCase6(node *redBlackNode) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.parent)
	node.parent.color = black
	if node == node.parent.left && nodeColor(sibling.right) == red {
		sibling.right.color = black
		t.rotateLeft(node.parent)
	} else if nodeColor(sibling.left) == red {
		sibling.left.color = black
		t.rotateRight(node.parent)
	}
}

func nodeColor(node *redBlackNode) color {
	if node == nil {
		return black
	}
	return node.color
}
