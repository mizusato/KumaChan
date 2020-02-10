package container

import . "kumachan/runtime/common"

type Seq interface {
	Next() (Value, Seq, bool)
}

type EmptySeq struct {}
func (_ EmptySeq) Next() (Value, Seq, bool) {
	return nil, nil, false
}

func SeqFrom(values []Value) Seq {
	return (Array {
		Length: uint(len(values)),
		GetItem: func(i uint) Value {
			return values[i]
		},
	}).Iterate()
}

func SeqOf(values ...Value) Seq {
	return SeqFrom(values)
}
