package compiler

import (
	ch "kumachan/checker"
    . "kumachan/error"
)


type Scope struct {
	Bindings     [] Binding
	BindingMap   map[string] uint
	BindingPeek  *uint
	Children     [] *Scope
	NextId       *uint
}

type Binding struct {
	Name   string
	Used   bool
	Point  ErrorPoint
	Id     uint
}

func MakeScope() *Scope {
	return &Scope {
		Bindings:    make([] Binding, 0),
		BindingMap:  make(map[string] uint),
		BindingPeek: new(uint),
		Children:    make([] *Scope, 0),
		NextId:      new(uint),
	}
}

func MakeClosureScope(outer *Scope) *Scope {
	var bindings = make([] Binding, len(outer.Bindings))
	for i, b := range outer.Bindings {
		bindings[i] = Binding {
			Name:  b.Name,
			Used:  false,
			Point: b.Point,
		}
	}
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	var child = &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: new(uint),
		Children:    make([] *Scope, 0),
		NextId:      outer.NextId,
	}
	outer.Children = append(outer.Children, child)
	return child
}

func MakeBranchScope(outer *Scope) *Scope {
	var bindings = make([] Binding, len(outer.Bindings))
	for i, b := range outer.Bindings {
		bindings[i] = b
	}
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	var child = &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: outer.BindingPeek,
		Children:    make([] *Scope, 0),
		NextId:      outer.NextId,
	}
	outer.Children = append(outer.Children, child)
	return child
}

func (scope *Scope) AddBinding(name string, point ErrorPoint) uint {
	var _, exists = scope.BindingMap[name]
	if exists {
		// shadowing: do nothing
	}
	var list = &(scope.Bindings)
	var offset = uint(len(*list))
	*list = append(*list, Binding {
		Name:  name,
		Used:  false,
		Point: point,
		Id:    *(scope.NextId),
	})
	*(scope.NextId) += 1
	var max = func(a uint, b uint) uint {
		if (a > b) { return a } else { return b }
	}
	scope.BindingMap[name] = offset
	*(scope.BindingPeek) = max(*(scope.BindingPeek), uint(len(*list)))
	return offset
}

func (scope *Scope) CollectUnused() ([] Binding) {
	var all = make(map[uint] *Binding)
	var collect func(*Scope)
	collect = func(scope *Scope) {
		for _, b := range scope.Bindings {
			var existing, exists = all[b.Id]
			if exists {
				existing.Used = (existing.Used || b.Used)
			} else {
				var b_copied = b  // make copy more clearly
				all[b.Id] = &b_copied
			}
		}
		for _, child := range scope.Children {
			collect(child)
		}
	}
	collect(scope)
	var unused = make([] Binding, 0)
	for _, b := range all {
		if !(b.Used) {
			unused = append(unused, *b)
		}
	}
	return unused
}

func (scope *Scope) CollectUnusedAsErrors() ([] *Error) {
	var unused = scope.CollectUnused()
	if len(unused) == 0 {
		return nil
	} else {
		var errs = make([] *Error, 0)
		for _, b := range unused {
			if b.Name == ch.IgnoreMark {
				continue
			}
			errs = append(errs, &Error {
				Point:    b.Point,
				Concrete: E_UnusedBinding { b.Name },
			})
		}
		return errs
	}
}
