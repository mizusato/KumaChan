package container

import (
	. "kumachan/runtime/common"
	"kumachan/stdlib"
	"reflect"
)


type Array struct {
	Length    uint
	GetItem   func(uint) Value
	ItemType  reflect.Type
}

func ArrayFromSlice(slice interface{}) Array {
	var slice_rv = reflect.ValueOf(slice)
	if slice_rv.Kind() != reflect.Slice {
		panic("cannot apply ArrayFromSlice() on non-slice value")
	}
	return Array {
		Length:  uint(slice_rv.Len()),
		GetItem: func(i uint) Value {
			return slice_rv.Index(int(i))
		},
		ItemType: slice_rv.Type().Elem(),
	}
}

func ArrayFrom(value Value) Array {
	var slice, ok = stdlib.AdaptSlice(value)
	if ok {
		return Array {
			Length:  uint(slice.Len()),
			GetItem: func(index uint) Value {
				return slice.Index(int(index))
			},
			ItemType: slice.Type().Elem(),
		}
	} else {
		return value.(Array)
	}
}

func (array Array) CopyAsSlice() Value {
	var L = array.Length
	var slice_rv = reflect.MakeSlice(array.ItemType, int(L), int(L))
	for i := uint(0); i < L; i += 1 {
		slice_rv.Index(int(i)).Set(reflect.ValueOf(array.GetItem(i)))
	}
	return slice_rv.Interface()
}

type ArrayIterator struct {
	Array      Array
	NextIndex  uint
}

type ArrayReversedIterator struct {
	Array      Array
	NextIndex  uint
}

func (it ArrayIterator) GetItemType() reflect.Type {
	return it.Array.ItemType
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

func (it ArrayReversedIterator) GetItemType() reflect.Type {
	return it.Array.ItemType
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

func (array Array) MapView(f func(Value)Value) Array {
	return Array {
		Length:  array.Length,
		GetItem: func(i uint) Value {
			return f(array.GetItem(i))
		},
	}
}

func (array Array) SliceView(low uint, high uint) Array {
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

func (array Array) Sort(cmp Compare) Seq {
	var L = array.Length
	if L == 0 {
		return EmptySeq { array.ItemType }
	} else if L == 1 {
		return OneShotSeq {
			ItemType: array.ItemType,
			Item:     array.GetItem(0),
		}
	} else {
		var M = (L / 2)
		var left = array.SliceView(0, M)
		var right = array.SliceView(M, L)
		return MergeSortIterator {
			Left:  left.Sort(cmp),
			Right: right.Sort(cmp),
			Cmp:   cmp,
		}
	}
}


type MergeSortIterator struct {
	Left  Seq
	Right Seq
	Cmp   Compare
}

func (m MergeSortIterator) GetItemType() reflect.Type {
	var lt = m.Left.GetItemType()
	var rt = m.Right.GetItemType()
	if rt.AssignableTo(lt) {
		return lt
	} else {
		panic("something went wrong")
	}
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
			case Smaller, Equal:
				order_preserved = true
			case Bigger:
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

