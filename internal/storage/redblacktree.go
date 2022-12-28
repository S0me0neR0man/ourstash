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

func newRedBlackTree() *redBlackTree {
	return &redBlackTree{}
}

// put inserts key into the tree.
// thread safe
func (t *redBlackTree) put(key Key) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.root == nil {
		t.root = &redBlackNode{key: key, color: black}
		t.size++
		return
	}

	curNode := t.root
	for {
		switch key.Compare(curNode.key) {
		case KeyEqual:
			return
		case KeyLessThan:
			if curNode.left == nil {
				curNode.left = &redBlackNode{key: key, color: red}
				curNode.left.parent = curNode
				t.insertCase1(curNode.left)
				t.size++
				return
			}
			curNode = curNode.left
		case KeyMoreThan:
			if curNode.right == nil {
				curNode.right = &redBlackNode{key: key, color: red}
				curNode.right.parent = curNode
				t.insertCase1(curNode.right)
				t.size++
				return
			}
			curNode = curNode.right
		}
	}
}

// get searches the node in the tree, nil not found
//
// thread safe
func (t *redBlackTree) get(key Key) *redBlackNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.lookup(key)
}

// remove the node from the tree
//
// thread safe
func (t *redBlackTree) remove(key Key) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delNode := t.lookup(key)
	if delNode == nil {
		return
	}

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

// sizeof returns number of nodes
//
// IMPORTANT: does not provide thread safety
func (t *redBlackTree) sizeof() int {
	return t.size
}

// sizeof returns the number of elements in the subtree.
//
// IMPORTANT: does not provide thread safety
func (n *redBlackNode) sizeof(tree *redBlackTree) int {
	return n.size()
}

// size calc current size
func (n *redBlackNode) size() int {
	if n == nil {
		return 0
	}
	size := 1
	if n.left != nil {
		size += n.left.size()
	}
	if n.right != nil {
		size += n.right.size()
	}
	return size
}

// String implements Stringer interface
//
// IMPORTANT: does not provide thread safety
func (t *redBlackTree) String() string {
	str := "redBlackTree\n"
	if t.size != 0 {
		output(t.root, "", true, &str)
	}
	return str
}

// String implements Stringer interface
//
// IMPORTANT: does not provide thread safety
func (n *redBlackNode) String() string {
	color := "B"
	if nodeColor(n) == red {
		color = "R"
	}
	return fmt.Sprintf("%s %v", color, n.key)
}

func output(node *redBlackNode, prefix string, tail bool, str *string) {
	if node.right != nil {
		newPrefix := prefix
		if tail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.right, newPrefix, false, str)
	}

	*str += prefix
	if tail {
		*str += "└── "
	} else {
		*str += "┌── "
	}

	*str += node.String() + "\n"
	if node.left != nil {
		newPrefix := prefix
		if tail {
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
		switch key.Compare(curNode.key) {
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

func (t *redBlackTree) insertCase4(node *redBlackNode) {
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

func (t *redBlackTree) insertCase5(node *redBlackNode) {
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
	for n.right != nil {
		n = n.right
	}
	return n
}

func (t *redBlackTree) deleteCase1(node *redBlackNode) {
	if node.parent == nil {
		return
	}
	t.deleteCase2(node)
}

func (t *redBlackTree) deleteCase2(node *redBlackNode) {
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

func (t *redBlackTree) deleteCase3(node *redBlackNode) {
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

func (t *redBlackTree) deleteCase4(node *redBlackNode) {
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

func (t *redBlackTree) deleteCase5(node *redBlackNode) {
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

func (t *redBlackTree) deleteCase6(node *redBlackNode) {
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
