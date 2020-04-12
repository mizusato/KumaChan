package container

import (
	. "kumachan/runtime/common"
	"reflect"
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

type RangeIterator struct {
	Current  uint
	Bound    uint
}
func (r RangeIterator) Next() (Value, Seq, bool) {
	if r.Current < r.Bound {
		return r.Current, RangeIterator {
			Current: r.Current + 1,
			Bound:   r.Bound,
		}, true
	} else {
		return (^uint(0)), RangeIterator{}, false
	}
}
func (r RangeIterator) GetItemType() reflect.Type {
	return reflect.TypeOf(uint(0))
}

func Collect(seq Seq) Value {
	var slice_rv = reflect.MakeSlice(seq.GetItemType(), 0, 0)
	for item,rest,ok := seq.Next(); ok; item,rest,ok = rest.Next() {
		slice_rv = reflect.Append(slice_rv, reflect.ValueOf(item))
	}
	return slice_rv.Interface()
}
