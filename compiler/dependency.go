package compiler

import (
	c "kumachan/runtime/common"
)


type FuncNode struct {
	Underlying    *c.Function
	Dependencies  [] Dependency
}
type Dependency interface { Dependency() }

func (impl DepFunction) Dependency() {}
type DepFunction struct {
	Module  string
	Name    string
	Index   uint
}
func (impl DepConstant) Dependency() {}
type DepConstant struct {
	Module  string
	Name    string
}
func (impl DepData) Dependency() {}
type DepData struct {
	Index  uint
}
func (impl DepClosure) Dependency() {}
type DepClosure struct {
	Index  uint
}

func FuncNodeFrom (
	f         *c.Function,
	refs      [] GlobalRef,
	idx       Index,
	data      *([] c.DataValue),
	closures  *([] FuncNode),
) FuncNode {
	var deps = RefsToDeps(refs, idx, data, closures)
	return FuncNode {
		Underlying:   f,
		Dependencies: deps,
	}
}

func RefsToDeps (
	refs      [] GlobalRef,
	idx       Index,
	data      *([] c.DataValue),
	closures  *([] FuncNode),
) []Dependency {
	var deps = make([]Dependency, len(refs))
	for i, ref := range refs {
		switch r := ref.(type) {
		case RefData:
			var index = uint(len(*data))
			*data = append(*data, r.DataValue)
			deps[i] = DepData {
				Index: index,
			}
		case RefClosure:
			var index = uint(len(*closures))
			var cl = FuncNodeFrom (
				r.Function,
				r.GlobalRefs,
				idx,
				data,
				closures,
			)
			*closures = append(*closures, cl)
			deps[i] = DepClosure {
				Index: index,
			}
		case RefFun:
			deps[i] = DepFunction {
				Module: r.Module,
				Name:   r.Name,
				Index:  r.Index,
			}
		case RefConst:
			deps[i] = DepConstant {
				Module: r.Name.ModuleName,
				Name:   r.Name.SymbolName,
			}
		default:
			panic("impossible branch")
		}
	}
	return deps
}
