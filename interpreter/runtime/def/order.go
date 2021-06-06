package def


type Compare   func(Value,Value) Ordering
type Ordering  int
const (
	Equal  Ordering  =  iota
	Smaller
	Bigger
)
type LessThanOperator  func(Value,Value) bool

func (o Ordering) Reversed() Ordering {
	switch o {
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

func (cmp Compare) ReversedOrdering() Compare {
	return func(a Value, b Value) Ordering {
		return cmp(a, b).Reversed()
	}
}

