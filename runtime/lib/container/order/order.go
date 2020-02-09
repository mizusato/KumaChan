package order

import . "kumachan/runtime/common"

type Compare   func(Value,Value) Ordering
type Ordering  int
const (
	Equal  Ordering  =  iota
	Smaller
	Bigger
)

func (cmp Compare) ReversedOrder() Compare {
	return func(a Value, b Value) Ordering {
		switch cmp(a, b) {
		case Smaller:
			return Bigger
		case Bigger:
			return Smaller
		case Equal:
			return Equal
		default:
			panic("impossible branch")
		}
	}
}