package container

import (
	"reflect"
	. "kumachan/error"
	. "kumachan/runtime/common"
)


type Seq interface {
	Inspectable
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
func (_ EmptySeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq empty]")
	return msg
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
func (_ OneShotSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq one-shot]")
	return msg
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
func (_ ConsSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq cons]")
	return msg
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
func (_ RangeSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq range]")
	return msg
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
func (_ MappedSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq mapped]")
	return msg
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
func (_ FilteredSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq filtered]")
	return msg
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
func (_ ScannedSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq scanned]")
	return msg
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
func (_ ChunkedSeq) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq chunked]")
	return msg
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
