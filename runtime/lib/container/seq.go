package container

import (
	"reflect"
	. "kumachan/lang"
)


type Seq interface {
	Next() (Value, Seq, bool)
	GetItemType() reflect.Type
}

type EmptySeq struct {
	ItemType  reflect.Type
}
func (_ EmptySeq) Next() (Value, Seq, bool) {
	return nil, nil, false
}
func (e EmptySeq) GetItemType() reflect.Type {
	return e.ItemType
}

type OneShotSeq struct {
	ItemType  reflect.Type
	Item      Value
}
func (o OneShotSeq) Next() (Value, Seq, bool) {
	return o.ItemType, EmptySeq { o.ItemType }, true
}
func (o OneShotSeq) GetItemType() reflect.Type {
	return o.ItemType
}

type ConsSeq struct {
	Head  Value
	Tail  Seq
}
func (p ConsSeq) Next() (Value, Seq, bool) {
	return p.Head, p.Tail, true
}
func (p ConsSeq) GetItemType() reflect.Type {
	return p.Tail.GetItemType()
}

type RangeSeq struct {
	Current  uint
	Bound    uint
}
func (r RangeSeq) Next() (Value, Seq, bool) {
	if r.Current < r.Bound {
		return r.Current, RangeSeq {
			Current: r.Current + 1,
			Bound:   r.Bound,
		}, true
	} else {
		return (^uint(0)), RangeSeq {}, false
	}
}
func (_ RangeSeq) GetItemType() reflect.Type {
	return reflect.TypeOf(uint(0))
}

type MappedSeq struct {
	Input   Seq
	Mapper  func(Value) Value
}
func (m MappedSeq) Next() (Value, Seq, bool) {
	var v, rest, ok = m.Input.Next()
	if ok {
		var f = m.Mapper
		return f(v), MappedSeq { Input: rest, Mapper: f }, true
	} else {
		return nil, nil, false
	}
}
func (_ MappedSeq) GetItemType() reflect.Type {
	return ValueReflectType()
}

type OptMappedSeq struct {
	Input      Seq
	MapFilter  func(Value) Value
}
func (o OptMappedSeq) Next() (Value, Seq, bool) {
	var v, rest, ok = o.Input.Next()
	if ok {
		var filtered_rest = OptMappedSeq {
			Input:      rest,
			MapFilter:  o.MapFilter,
		}
		var ok_v, ok = Unwrap(o.MapFilter(v).(SumValue))
		if ok {
			return ok_v, filtered_rest, true
		} else {
			return filtered_rest.Next()
		}
	} else {
		return nil, nil, false
	}
}
func (o OptMappedSeq) GetItemType() reflect.Type {
	return o.Input.GetItemType()
}


type FilteredSeq struct {
	Input   Seq
	Filter  func(Value) bool
}
func (f FilteredSeq) Next() (Value, Seq, bool) {
	var v, rest, ok = f.Input.Next()
	if ok {
		var filtered_rest = FilteredSeq { Input: rest, Filter: f.Filter }
		if f.Filter(v) {
			return v, filtered_rest, true
		} else {
			return filtered_rest.Next()
		}
	} else {
		return nil, nil, false
	}
}
func (f FilteredSeq) GetItemType() reflect.Type {
	return f.Input.GetItemType()
}

type FlatMappedSeq struct {
	Input    Seq
	Mapper   func(Value) Value
	Current  Seq
}
func (f FlatMappedSeq) Next() (Value, Seq, bool) {
	if f.Current == nil { panic("something went wrong") }
	var v, rest, ok = f.Current.Next()
	if ok {
		return v, FlatMappedSeq {
			Input:   f.Input,
			Mapper:  f.Mapper,
			Current: rest,
		}, true
	} else {
		var v, rest, ok = f.Input.Next()
		if ok {
			var inner_seq = f.Mapper(v).(Seq)
			var t = FlatMappedSeq {
				Input:   rest,
				Mapper:  f.Mapper,
				Current: inner_seq,
			}
			return t.Next()
		} else {
			return nil, nil, false
		}
	}
}
func (f FlatMappedSeq) GetItemType() reflect.Type {
	return f.Input.GetItemType()
}

type ScannedSeq struct {
	Previous  Value
	Rest      Seq
	Reducer   func(Value,Value) Value
}
func (s ScannedSeq) Next() (Value, Seq, bool) {
	var v, rest, ok = s.Rest.Next()
	if ok {
		var f = s.Reducer
		var new_previous = f(s.Previous, v)
		return new_previous, ScannedSeq {
			Previous: new_previous,
			Rest:     rest,
			Reducer:  f,
		}, true
	} else {
		return nil, nil, false
	}
}
func (_ ScannedSeq) GetItemType() reflect.Type {
	return ValueReflectType()
}

type ChunkedSeq struct {
	ChunkSize  uint
	Remaining  Seq
}
func (c ChunkedSeq) Next() (Value, Seq, bool) {
	if c.ChunkSize == 0 {
		return nil, nil, false
	} else {
		var item_t = c.Remaining.GetItemType()
		var slice_t = reflect.SliceOf(item_t)
		var chunk_rv = reflect.MakeSlice(slice_t, 0, int(c.ChunkSize))
		var v Value
		var remaining = c.Remaining
		var ok bool
		for i := uint(0); i < c.ChunkSize; i += 1 {
			v, remaining, ok = remaining.Next()
			if ok {
				chunk_rv = reflect.Append(chunk_rv, reflect.ValueOf(v))
			} else {
				break
			}
		}
		var chunk = chunk_rv.Interface()
		var L = uint(chunk_rv.Len())
		if L == c.ChunkSize {
			return chunk, ChunkedSeq {
				ChunkSize: c.ChunkSize,
				Remaining: remaining,
			}, true
		} else if L > 0 {
			return chunk, EmptySeq { ItemType: item_t}, true
		} else {
			return nil, nil, false
		}
	}
}
func (c ChunkedSeq) GetItemType() reflect.Type {
	var item_t = c.Remaining.GetItemType()
	var slice_t = reflect.SliceOf(item_t)
	return slice_t
}

func SeqCollect(seq Seq) interface{} {
	var t = reflect.SliceOf(seq.GetItemType())
	var slice_rv = reflect.MakeSlice(t, 0, 0)
	for item,rest,ok := seq.Next(); ok; item,rest,ok = rest.Next() {
		slice_rv = reflect.Append(slice_rv, reflect.ValueOf(item))
	}
	return slice_rv.Interface()
}

func SeqReduce(seq Seq, init Value, f func(Value,Value)Value) Value {
	var v = init
	for item,rest,ok := seq.Next(); ok; item,rest,ok = rest.Next() {
		v = f(v, item)
	}
	return v
}

func SeqSome(seq Seq, f func(Value)bool) bool {
	for item,rest,ok := seq.Next(); ok; item,rest,ok = rest.Next() {
		if f(item) {
			return true
		}
	}
	return false
}

func SeqEvery(seq Seq, f func(Value)bool) bool {
	return !(SeqSome(seq, func(item Value) bool { return !(f(item)) }))
}
