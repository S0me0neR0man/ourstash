package storage

import (
	"fmt"
	"sync"
)

type color bool

const (
	black, red color = true, false
)

// RedBlackTree main index
type RedBlackTree struct {
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

func NewRedBlackTree() *RedBlackTree {
	return &RedBlackTree{}
}

// Put inserts key into the tree.
func (t *RedBlackTree) Put(key Key) {
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
				curNode.right.parent = curNode
				t.insertCase1(curNode.right)
				t.size++
				t.mu.Unlock()
				return
			}
			curNode = curNode.right
		}
	}
}

// GetNode searches the currentNode in the tree, nil not found
func (t *RedBlackTree) GetNode(key Key) *redBlackNode {
	return t.lookup(key)
}

// Remove the currentNode from the tree
func (t *RedBlackTree) Remove(key Key) {
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
func (t *RedBlackTree) IsEmpty() bool {
	return t.size == 0
}

// Size returns number of nodes
func (t *RedBlackTree) Size() int {
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
func (t *RedBlackTree) Keys() []Key {
	keys := make([]Key, t.size)
	it := t.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// Left returns the minimal node or nil
func (t *RedBlackTree) Left() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := t.root; curNode != nil; curNode = curNode.left {
		parentNode = curNode
	}
	return parentNode
}

// Right returns the max node or nil
func (t *RedBlackTree) Right() *redBlackNode {
	var parentNode *redBlackNode
	for curNode := t.root; curNode != nil; curNode = curNode.right {
		parentNode = curNode
	}
	return parentNode
}

// Floor Finds floor currentNode of the input key, return the floor currentNode or nil if no floor is found.
func (t *RedBlackTree) Floor(key Key) (*redBlackNode, bool) {
	var foundNode *redBlackNode
	for curNode := t.root; curNode != nil; {
		switch curNode.key.Compare(key) {
		case KeyEqual:
			return curNode, true
		case KeyLessThan:
			curNode = curNode.left
		case KeyMoreThan:
			foundNode = curNode
			curNode = curNode.right
		}
	}
	if foundNode != nil {
		return foundNode, true
	}
	return nil, false
}

// Ceiling finds ceiling currentNode of the input key, return the ceiling currentNode or nil if no ceiling is found.
func (t *RedBlackTree) Ceiling(key Key) (*redBlackNode, bool) {
	var foundNode *redBlackNode
	for curNode := t.root; curNode != nil; {
		switch curNode.key.Compare(key) {
		case KeyEqual:
			return curNode, true
		case KeyLessThan:
			foundNode = curNode
			curNode = curNode.left
		case KeyMoreThan:
			curNode = curNode.right
		}
	}
	if foundNode != nil {
		return foundNode, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (t *RedBlackTree) Clear() {
	t.root = nil
	t.size = 0
}

// String implements Stringer interface
func (t *RedBlackTree) String() string {
	str := "RedBlackTree\n"
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

func (t *RedBlackTree) lookup(key Key) *redBlackNode {
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

func (t *RedBlackTree) rotateLeft(node *redBlackNode) {
	right := node.right
	t.replaceNode(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
	}
	right.left = node
	node.parent = right
}

func (t *RedBlackTree) rotateRight(node *redBlackNode) {
	left := node.left
	t.replaceNode(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
	}
	left.right = node
	node.parent = left
}

func (t *RedBlackTree) replaceNode(old *redBlackNode, new *redBlackNode) {
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

func (t *RedBlackTree) insertCase1(node *redBlackNode) {
	if node.parent == nil {
		node.color = black
	} else {
		t.insertCase2(node)
	}
}

func (t *RedBlackTree) insertCase2(node *redBlackNode) {
	if nodeColor(node.parent) == black {
		return
	}
	t.insertCase3(node)
}

func (t *RedBlackTree) insertCase3(node *redBlackNode) {
	uncleNode := node.uncle()
	if nodeColor(uncleNode) == red {
		node.parent.color = black
		uncleNode.color = black
		node.grandparent().color = red
		t.insertCase1(node.grandparent())
	} else {
		t.insertCase4(node)
	}
}

func (t *RedBlackTree) insertCase4(node *redBlackNode) {
	grandparentNode := node.grandparent()
	if node == node.parent.right && node.parent == grandparentNode.left {
		t.rotateLeft(node.parent)
		node = node.left
	} else if node == node.parent.left && node.parent == grandparentNode.right {
		t.rotateRight(node.parent)
		node = node.right
	}
	t.insertCase5(node)
}

func (t *RedBlackTree) insertCase5(node *redBlackNode) {
	node.parent.color = black
	grandparentNode := node.grandparent()
	grandparentNode.color = red
	if node == node.parent.left && node.parent == grandparentNode.left {
		t.rotateRight(grandparentNode)
	} else if node == node.parent.right && node.parent == grandparentNode.right {
		t.rotateLeft(grandparentNode)
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

func (t *RedBlackTree) deleteCase1(node *redBlackNode) {
	if node.parent == nil {
		return
	}
	t.deleteCase2(node)
}

func (t *RedBlackTree) deleteCase2(node *redBlackNode) {
	siblingNode := node.sibling()
	if nodeColor(siblingNode) == red {
		node.parent.color = red
		siblingNode.color = black
		if node == node.parent.left {
			t.rotateLeft(node.parent)
		} else {
			t.rotateRight(node.parent)
		}
	}
	t.deleteCase3(node)
}

func (t *RedBlackTree) deleteCase3(node *redBlackNode) {
	siblingNode := node.sibling()
	if nodeColor(node.parent) == black &&
		nodeColor(siblingNode) == black &&
		nodeColor(siblingNode.left) == black &&
		nodeColor(siblingNode.right) == black {
		siblingNode.color = red
		t.deleteCase1(node.parent)
	} else {
		t.deleteCase4(node)
	}
}

func (t *RedBlackTree) deleteCase4(node *redBlackNode) {
	siblingNode := node.sibling()
	if nodeColor(node.parent) == red &&
		nodeColor(siblingNode) == black &&
		nodeColor(siblingNode.left) == black &&
		nodeColor(siblingNode.right) == black {
		siblingNode.color = red
		node.parent.color = black
	} else {
		t.deleteCase5(node)
	}
}

func (t *RedBlackTree) deleteCase5(node *redBlackNode) {
	siblingNode := node.sibling()
	if node == node.parent.left &&
		nodeColor(siblingNode) == black &&
		nodeColor(siblingNode.left) == red &&
		nodeColor(siblingNode.right) == black {
		siblingNode.color = red
		siblingNode.left.color = black
		t.rotateRight(siblingNode)
	} else if node == node.parent.right &&
		nodeColor(siblingNode) == black &&
		nodeColor(siblingNode.right) == red &&
		nodeColor(siblingNode.left) == black {
		siblingNode.color = red
		siblingNode.right.color = black
		t.rotateLeft(siblingNode)
	}
	t.deleteCase6(node)
}

func (t *RedBlackTree) deleteCase6(node *redBlackNode) {
	siblingNode := node.sibling()
	siblingNode.color = nodeColor(node.parent)
	node.parent.color = black
	if node == node.parent.left && nodeColor(siblingNode.right) == red {
		siblingNode.right.color = black
		t.rotateLeft(node.parent)
	} else if nodeColor(siblingNode.left) == red {
		siblingNode.left.color = black
		t.rotateRight(node.parent)
	}
}

func nodeColor(node *redBlackNode) color {
	if node == nil {
		return black
	}
	return node.color
}
