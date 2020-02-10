package container

import (
	. "kumachan/runtime/common"
)

type Seq interface {
	Next() (Value, Seq, bool)
}

type EmptySeq struct {}
func (_ EmptySeq) Next() (Value, Seq, bool) {
	return nil, nil, false
}

func SeqFrom(values []Value) Seq {
	return ArrayFrom(values).Iterate()
}

func SeqOf(v Value) Seq {
	return SeqFunc(func() (Value, Seq, bool) {
		return v, EmptySeq{}, true
	})
}

type SeqFunc  func() (Value, Seq, bool)

func (sf SeqFunc) Next() (Value, Seq, bool) {
	return sf()
}
