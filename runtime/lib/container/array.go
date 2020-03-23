package container

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container/order"
)

type Array struct {
	Length   uint
	GetItem  func(uint) Value
}

func ArrayFrom(values []Value) Array {
	return Array {
		Length:  uint(len(values)),
		GetItem: func(i uint) Value {
			return values[i]
		},
	}
}

type ArrayIterator struct {
	Array      Array
	NextIndex  uint
}

type ArrayReversedIterator struct {
	Array      Array
	NextIndex  uint
}

func (it ArrayIterator) Next() (Value, Seq, bool) {
	var array = it.Array
	var next = it.NextIndex
	if next < array.Length {
		return array.GetItem(next), ArrayIterator {
			Array:     array,
			NextIndex: (next + 1),
		}, true
	} else {
		return nil, nil, false
	}
}

func (it ArrayReversedIterator) Next() (Value, Seq, bool) {
	var array = it.Array
	var next = it.NextIndex
	if next >= 0 {
		return array.GetItem(next), ArrayIterator {
			Array:     array,
			NextIndex: (next - 1),
		}, true
	} else {
		return nil, nil, false
	}
}

func (array Array) Iterate() Seq {
	return ArrayIterator {
		Array:     array,
		NextIndex: 0,
	}
}

func (array Array) IterateReversed() Seq {
	return ArrayReversedIterator {
		Array:     array,
		NextIndex: (array.Length - 1),
	}
}

func (array Array) Map(f func(Value)Value) Array {
	return Array {
		Length:  array.Length,
		GetItem: func(i uint) Value {
			return f(array.GetItem(i))
		},
	}
}

func (array Array) CarefullySlice(low uint, high uint) Array {
	var L = array.Length
	if !(low <= high && low < L && high <= L) {
		panic("invalid slice bounds")
	}
	return Array {
		Length:  (high - low),
		GetItem: func(i uint) Value {
			return array.GetItem(i + low)
		},
	}
}

func (array Array) Sort(cmp order.Compare) Seq {
	var L = array.Length
	if L == 0 {
		return EmptySeq {}
	} else if L == 1 {
		return SeqOf(array.GetItem(0))
	} else {
		var M = (L / 2)
		var left = array.CarefullySlice(0, M)
		var right = array.CarefullySlice(M, L)
		return MergeSortIterator {
			Left:  left.Sort(cmp),
			Right: right.Sort(cmp),
			Cmp:   cmp,
		}
	}
}

type MergeSortIterator struct {
	Left   Seq
	Right  Seq
	Cmp    order.Compare
}

func (m MergeSortIterator) Next() (Value, Seq, bool) {
	var left = m.Left
	var right = m.Right
	var cmp = m.Cmp
	var l, l_rest, l_exists = left.Next()
	var r, r_rest, r_exists = right.Next()
	if !l_exists && !r_exists {
		return nil, nil, false
	} else {
		var order_preserved bool
		if l_exists && r_exists {
			switch cmp(l, r) {
			case order.Smaller, order.Equal:
				order_preserved = true
			case order.Bigger:
				order_preserved = false
			default:
				panic("impossible branch")
			}
		} else if l_exists {
			order_preserved = true
		} else if r_exists {
			order_preserved = false
		} else {
			panic("impossible branch")
		}
		if order_preserved {
			return l, MergeSortIterator {
				Left:  l_rest,
				Right: right,
				Cmp:   cmp,
			}, true
		} else {
			return r, MergeSortIterator {
				Left:  left,
				Right: r_rest,
				Cmp:   cmp,
			}, true
		}
	}
}

