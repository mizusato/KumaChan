package avl

import (
	. "kumachan/lang"
)

/* Functional AVL Tree: Underlying Implementation of Ordered Set and Map */
type AVL struct {
	Value   Value
	Left    *AVL
	Right   *AVL
	Size    uint64
	Height  uint64
}

func Node(v Value, left *AVL, right *AVL) *AVL {
	var node = &AVL {
		Value:  v,
		Left:   left,
		Right:  right,
		Size:   1 + left.GetSize() + right.GetSize(),
		Height: 1 + max(left.GetHeight(), right.GetHeight()),
	}
	return node.balanced()
}
func Leaf(v Value) *AVL {
	return &AVL {
		Value:  v,
		Left:   nil,
		Right:  nil,
		Size:   1,
		Height: 1,
	}
}

func (node *AVL) IsLeaf() bool {
	return (node.Left == nil && node.Right == nil)
}
func (node *AVL) GetSize() uint64 {
	if node == nil {
		return 0
	} else {
		return node.Size
	}
}
func (node *AVL) GetHeight() uint64 {
	if node == nil {
		return 0
	} else {
		return node.Height
	}
}

func (node *AVL) Lookup(target Value, cmp Compare) (Value, bool) {
	if node == nil {
		return nil, false
	} else {
		switch cmp(target, node.Value) {
		case Smaller:
			return node.Left.Lookup(target, cmp)
		case Bigger:
			return node.Right.Lookup(target, cmp)
		case Equal:
			return node.Value, true
		default:
			panic("impossible branch")
		}
	}
}

func (node *AVL) Inserted(inserted Value, cmp Compare) (*AVL, bool) {
	if node == nil {
		return Leaf(inserted), false
	} else {
		var value = node.Value
		var left = node.Left
		var right = node.Right
		switch cmp(inserted, value) {
		case Smaller:
			var left_inserted, override = left.Inserted(inserted, cmp)
			return Node(value, left_inserted, right), override
		case Bigger:
			var right_inserted, override = right.Inserted(inserted, cmp)
			return Node(value, left, right_inserted), override
		case Equal:
			return Node(inserted, left, right), true
		default:
			panic("impossible branch")
		}
	}
}

func (node *AVL) Deleted(target Value, cmp Compare) (Value, *AVL, bool) {
	if node == nil {
		return nil, nil, false
	} else {
		var value = node.Value
		var left = node.Left
		var right = node.Right
		switch cmp(target, value) {
		case Smaller:
			var deleted, rest, found = left.Deleted(target, cmp)
			if found {
				return deleted, Node(value, rest, right), true
			} else {
				return nil, node, false
			}
		case Bigger:
			var deleted, rest, found = right.Deleted(target, cmp)
			if found {
				return deleted, Node(value, left, rest), true
			} else {
				return nil, node, false
			}
		case Equal:
			if left == nil {
				return value, right, true
			} else if right == nil {
				return value, left, true
			} else {
				var node_state, _ = node.GetBalanceState()
				if node_state == RightTaller {
					var successor, rest_right, found = right.DeleteMin()
					assert(found, "right subtree should not be empty")
					return value, Node(successor, left, rest_right), true
				} else {
					var prior, rest_left, found = left.DeletedMax()
					assert(found, "left subtree should not be empty")
					return value, Node(prior, rest_left, right), true
				}
			}
		default:
			panic("impossible branch")
		}
	}
}

func (node *AVL) DeleteMin() (Value, *AVL, bool) {
	if node == nil {
		return nil, nil, false
	} else {
		var value = node.Value
		var left = node.Left
		var right = node.Right
		var deleted, rest, found = left.DeleteMin()
		if found {
			return deleted, Node(value, rest, right), true
		} else {
			return value, right, true
		}
	}
}

func (node *AVL) DeletedMax() (Value, *AVL, bool) {
	if node == nil {
		return nil, nil, false
	} else {
		var value = node.Value
		var left = node.Left
		var right = node.Right
		var deleted, rest, found = right.DeletedMax()
		if found {
			return deleted, Node(value, left, rest), true
		} else {
			return value, left, true
		}
	}
}

func (node *AVL) Walk(f func(Value)) {
	if node == nil { return }
	node.Left.Walk(f)
	f(node.Value)
	node.Right.Walk(f)
}

type BalanceState int
const (
	LeftTaller  BalanceState  =  iota
	RightTaller
	NeitherTaller
)
func (node *AVL) GetBalanceState() (BalanceState, uint) {
	var L = node.Left.GetHeight()
	var R = node.Right.GetHeight()
	if L > R {
		return LeftTaller, uint(L - R)
	} else if L < R {
		return RightTaller, uint(R - L)
	} else {
		return NeitherTaller, 0
	}
}
func (node *AVL) balanced() *AVL {
	var current = node
	var current_state, diff = current.GetBalanceState()
	if current_state == NeitherTaller || diff == 1 {
		return current
	} else {
		assert(diff == 2, "invalid usage of balanced()")
		switch current_state {
		case LeftTaller:
			var left = current.Left
			var left_state, _ = left.GetBalanceState()
			switch left_state {
			case LeftTaller:
				var new_right = Node(current.Value, left.Right, current.Right)
				var new_current = Node(left.Value, left.Left, new_right)
				return new_current
			case RightTaller:
				var middle = left.Right
				var new_left = Node(left.Value, left.Left, middle.Left)
				var new_right = Node(current.Value, middle.Right, current.Right)
				var new_current = Node(middle.Value, new_left, new_right)
				return new_current
			default:
				panic("invalid usage of balanced()")
			}
		case RightTaller:
			var right = current.Right
			var right_state, _ = right.GetBalanceState()
			switch right_state {
			case LeftTaller:
				var middle = right.Left
				var new_left = Node(current.Value, current.Left, middle.Left)
				var new_right = Node(right.Value, middle.Right, right.Right)
				var new_current = Node(middle.Value, new_left, new_right)
				return new_current
			case RightTaller:
				var new_left = Node(current.Value, current.Left, right.Left)
				var new_current = Node(right.Value, new_left, right.Right)
				return new_current
			default:
				panic("invalid usage of balanced()")
			}
		default:
			panic("impossible branch")
		}
	}
}


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}

func max(a uint64, b uint64) uint64 {
	if a >= b { return a } else { return b }
}
