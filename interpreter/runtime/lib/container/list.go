package container

import (
	"fmt"
	"reflect"
	"kumachan/standalone/rx"
	. "kumachan/interpreter/base"
	. "kumachan/standalone/util/error"
	"strings"
)


type List struct {
	head  uint
	tail  uint
	rev   bool  // whether head and tail are reversed when they are equal
	data  reflect.Value  // [] T
}
func (l List) EmptyListOfSameType() List {
	return List {
		head: 0,
		tail: 0,
		data: reflect.MakeSlice(l.data.Type(), 0, 0),
	}
}
func (l List) Length() uint {
	if l.data.Len() > 0 {
		if l.head <= l.tail {
			return (l.tail - l.head + 1)
		} else {
			return (l.head - l.tail + 1)
		}
	} else {
		return 0
	}
}
func (l List) at(index uint) Value {
	if !(index < l.Length()) {
		panic("list index out of range")
	}
	if l.head <= l.tail {
		return l.data.Index(int(l.head + index)).Interface()
	} else {
		return l.data.Index(int(l.head - index)).Interface()
	}
}

func ListFrom(value Value) List {
	var get_slice = func (v interface{}) (reflect.Value, bool) {
		var rv = reflect.ValueOf(v)
		var t = rv.Type()
		if t.Kind() == reflect.Slice {
			return rv, true
		} else {
			return reflect.ValueOf(nil), false
		}
	}
	var slice, ok = get_slice(value)
	if ok {
		return List {
			head: 0,
			tail: uint(slice.Len() - 1),
			rev:  false,
			data: slice,
		}
	} else {
		return value.(List)
	}
}

func (l List) ItemType() reflect.Type {
	return l.data.Type().Elem()
}

func (l List) ForEach(f func(uint, Value)) {
	var L = l.Length()
	for i := uint(0); i < L; i += 1 {
		f(i, l.at(i))
	}
}

func (l List) ForEachWithError(f func(uint, Value) error) error {
	var L = l.Length()
	for i := uint(0); i < L; i += 1 {
		var err = f(i, l.at(i))
		if err != nil { return err }
	}
	return nil
}

func (l List) CopyAsSlice() Value {
	var L = l.Length()
	var slice_rv = reflect.MakeSlice(l.data.Type(), int(L), int(L))
	for i := uint(0); i < L; i += 1 {
		slice_rv.Index(int(i)).Set(reflect.ValueOf(l.at(i)))
	}
	var slice = slice_rv.Interface()
	return slice
}

func (l List) CopyAsString() string {
	var L = l.Length()
	var buf strings.Builder
	for i := uint(0); i < L; i += 1 {
		buf.WriteRune(l.at(i).(rune))
	}
	return buf.String()
}

func (l List) CopyAsObservables() ([] rx.Observable) {
	var slice = make([] rx.Observable, l.Length())
	l.ForEach(func(i uint, item Value) {
		slice[i] = item.(rx.Observable)
	})
	return slice
}

func (l List) CopyAsStringSlice() ([] string) {
	var slice = make([] string, l.Length())
	l.ForEach(func(i uint, item Value) {
		slice[i] = item.(string)
	})
	return slice

}

func (l List) Reversed() Value {
	return List {
		head: l.tail,
		tail: l.head,
		rev:  !(l.rev),
		data: l.data,
	}
}

func (l List) Iterate() Seq {
	return ListIterator {
		List:      l,
		NextIndex: 0,
	}
}

func (l List) Sort(lt LessThanOperator) Seq {
	var slice = l.CopyAsSlice()
	return mergeSort(reflect.ValueOf(slice), lt)
}

func (l List) Shifted() (Value, List, bool) {
	if l.head < l.tail {
		return l.at(0), List {
			head: (l.head + 1),
			tail: l.tail,
			data: l.data,
		}, true
	} else if l.head > l.tail {
		return l.at(0), List {
			head: (l.head - 1),
			tail: l.tail,
			rev:  true,
			data: l.data,
		}, true
	} else {
		if l.Length() > 0 {
			return l.at(0), l.EmptyListOfSameType(), true
		} else {
			return nil, List{}, false
		}
	}
}

func (l List) Popped() (Value, List, bool) {
	if l.head < l.tail {
		var last = l.at(l.Length() - 1)
		return last, List {
			head: l.head,
			tail: (l.tail - 1),
			data: l.data,
		}, true
	} else if l.head > l.tail {
		var last = l.at(l.Length() - 1)
		return last, List {
			head: l.head,
			tail: (l.tail + 1),
			rev:  true,
			data: l.data,
		},true
	} else {
		if l.Length() > 0 {
			var last = l.at(l.Length() - 1)
			return last, l.EmptyListOfSameType(), true
		} else {
			return nil, List{}, false
		}
	}
}

func (l List) Unshifted() List {
	if l.head < l.tail || (l.head == l.tail && !(l.rev)) {
		if l.head > 0 {
			return List {
				head: (l.head - 1),
				tail: l.tail,
				data: l.data,
			}
		} else {
			return l
		}
	} else if l.head > l.tail || (l.head == l.tail && l.rev) {
		if (l.head + 1) < uint(l.data.Len()) {
			return List {
				head: (l.head + 1),
				tail: l.tail,
				data: l.data,
			}
		} else {
			return l
		}
	} else {
		panic("impossible branch")
	}
}

func (l List) Unpopped() List {
	if l.head < l.tail || (l.head == l.tail && !(l.rev)) {
		if (l.tail + 1) < uint(l.data.Len()) {
			return List {
				head: l.head,
				tail: (l.tail + 1),
				data: l.data,
			}
		} else {
			return l
		}
	} else if l.head > l.tail || (l.head == l.tail && l.rev) {
		if l.tail > 0 {
			return List {
				head: l.head,
				tail: (l.tail - 1),
				data: l.data,
			}
		} else {
			return l
		}
	} else {
		panic("impossible branch")
	}
}

func (l List) Inspect(inspect func(Value)(ErrorMessage)) ErrorMessage {
	var L = l.data.Len()
	var head = l.head
	var tail = l.tail
	var items = make([] ErrorMessage, 0)
	for i := 0; i < L; i += 1 {
		var entry_msg = make(ErrorMessage, 0)
		if head == tail && l.rev {
			entry_msg.WriteText(TS_BOLD, "(reversed)")
		}
		if uint(i) == head {
			entry_msg.WriteText(TS_BOLD, "[head] --> ")
		}
		if uint(i) == tail {
			entry_msg.WriteText(TS_BOLD, "[tail] --> ")
		}
		entry_msg.WriteText(TS_NORMAL, fmt.Sprintf("[%d]", i))
		entry_msg.Write(T_SPACE)
		entry_msg.WriteAll(inspect(l.data.Index(i).Interface()))
		items = append(items, entry_msg)
	}
	var title = fmt.Sprintf("List(%d)[%d,%d]", L, head, tail)
	return ListErrMsgItems(items, title)
}

type ListIterator struct {
	List       List
	NextIndex  uint
}
func (it ListIterator) GetItemType() reflect.Type {
	return it.List.ItemType()
}
func (it ListIterator) Next() (Value, Seq, bool) {
	var list = it.List
	var next = it.NextIndex
	if next < list.Length() {
		return list.at(next), ListIterator {
			List:      list,
			NextIndex: (next + 1),
		}, true
	} else {
		return nil, nil, false
	}
}

type MergeSortIterator struct {
	Left   Seq
	Right  Seq
	LtOp   LessThanOperator
}
func mergeSort(slice_rv reflect.Value, lt LessThanOperator) Seq {
	var item_type = slice_rv.Type().Elem()
	var L = slice_rv.Len()
	if L == 0 {
		return EmptySeq { item_type }
	} else if L == 1 {
		return OneShotSeq {
			ItemType: item_type,
			Item:     slice_rv.Index(0).Interface(),
		}
	} else {
		var M = (L / 2)
		var left = slice_rv.Slice(0, M)
		var right = slice_rv.Slice(M, L)
		return MergeSortIterator {
			Left:  mergeSort(left, lt),
			Right: mergeSort(right, lt),
			LtOp:  lt,
		}
	}
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
	var lt = m.LtOp
	var l, l_rest, l_exists = left.Next()
	var r, r_rest, r_exists = right.Next()
	if !l_exists && !r_exists {
		return nil, nil, false
	} else {
		var order_preserved bool
		if l_exists && r_exists {
			if !(lt(r, l)) {
				// `l` is smaller than or equal to `r`
				order_preserved = true
			} else {
				// `l` is greater than `r`
				order_preserved = false
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
				LtOp:  lt,
			}, true
		} else {
			return r, MergeSortIterator {
				Left:  left,
				Right: r_rest,
				LtOp:  lt,
			}, true
		}
	}
}

