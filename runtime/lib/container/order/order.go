package order

import . "kumachan/runtime/common"

type Compare   func(Value,Value) Ordering
type Ordering  int
const (
	Equal  Ordering  =  iota
	Smaller
	Bigger
)
