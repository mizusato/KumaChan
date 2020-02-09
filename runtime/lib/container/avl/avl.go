package avl

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container/order"
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
	return &AVL {
		Value:  v,
		Left:   left,
		Right:  right,
		Size:   1 + left.GetSize() + right.GetSize(),
		Height: 1 + max(left.GetHeight(), right.GetHeight()),
	}
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

func (node *AVL) Inserted(inserted Value, cmp order.Compare) *AVL {
	if node == nil {
		return Leaf(inserted)
	} else {
		var value = node.Value
		var left = node.Left
		var right = node.Right
		switch cmp(inserted, node.Value) {
		case order.Smaller:
			return Node(value, left.Inserted(inserted, cmp), right).balanced()
		case order.Bigger:
			return Node(value, left, right.Inserted(inserted, cmp)).balanced()
		case order.Equal:
			return Node(inserted, left, right)
		default:
			panic("impossible branch")
		}
	}
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
