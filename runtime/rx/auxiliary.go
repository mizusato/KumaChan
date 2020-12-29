package rx


type Pair struct {
	First   Object
	Second  Object
}

type Optional struct {
	HasValue  bool
	Value     Object
}

type Stack struct {
	top   Object
	rest  *Stack
}

func (s *Stack) Pushed(obj Object) *Stack {
	return &Stack {
		top:  obj,
		rest: s,
	}
}

func (s *Stack) Popped() (Object, *Stack, bool) {
	if s != nil {
		return s.top, s.rest, true
	} else {
		return nil, nil, false
	}
}

func (s *Stack) Top() (Object, bool) {
	if s != nil {
		return s.top, true
	} else {
		return nil, false
	}
}

func (s *Stack) Empty() bool {
	return (s == nil)
}

func (s *Stack) NotEmpty() bool {
	return (s != nil)
}

