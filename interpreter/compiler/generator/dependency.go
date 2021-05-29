package generator

import (
	"kumachan/interpreter/base"
	ch "kumachan/interpreter/compiler/checker"
)


type FuncNode struct {
	Underlying    *base.Function
	Dependencies  [] Dependency
	ch.FunctionGeneratorFlags
	ch.FunctionKmdInfo
}
var __DefaultFlags = ch.FunctionGeneratorFlags {}
var __DefaultKmdInfo = ch.FunctionKmdInfo {}

type Dependency interface { Dependency() }
type DepLocator struct {
	Locate  func(Dependency) (uint, bool)
	Offset  uint
}

func (impl DepFunction) Dependency() {}
type DepFunction struct {
	Module  string
	Name    string
	Index   uint
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
	f         *base.Function,
	refs      [] GlobalRef,
	data      *([] base.DataValue),
	closures  *([] FuncNode),
	flags     ch.FunctionGeneratorFlags,
	kmd_info  ch.FunctionKmdInfo,
) FuncNode {
	var deps = RefsToDeps(refs, data, closures)
	return FuncNode {
		Underlying: f,
		Dependencies: deps,
		FunctionGeneratorFlags: flags,
		FunctionKmdInfo: kmd_info,
	}
}

func RefsToDeps (
	refs      [] GlobalRef,
	data      *([] base.DataValue),
	closures  *([] FuncNode),
) ([] Dependency) {
	var deps = make([] Dependency, len(refs))
	for i, ref := range refs {
		switch r := ref.(type) {
		case RefData:
			var index = uint(len(*data))
			*data = append(*data, r.DataValue)
			deps[i] = DepData {
				Index: index,
			}
		case RefClosure:
			var cl = FuncNodeFrom (
				r.Function,
				r.GlobalRefs,
				data,
				closures,
				__DefaultFlags,
				__DefaultKmdInfo,
			)
			var index = uint(len(*closures))
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
		default:
			panic("impossible branch")
		}
	}
	return deps
}

func RelocateCode(f *FuncNode, locator DepLocator, extra ExtraDepLocator) {
	var inst_seq = f.Underlying.Code
	for i, _ := range inst_seq {
		switch inst_seq[i].OpCode {
		case base.GLOBAL, base.ARRAY:
			var relative_index = inst_seq[i].GetGlobalIndex()
			var dep = f.Dependencies[relative_index]
			var absolute_index uint
			if extra != nil {
				var extra_index, is_extra = extra(dep)
				if is_extra {
					absolute_index = (locator.Offset + extra_index)
				} else {
					var index, exists = locator.Locate(dep)
					if !(exists) { panic("something went wrong") }
					absolute_index = index
				}
			} else {
				var index, exists = locator.Locate(dep)
				if !(exists) { panic("something went wrong") }
				absolute_index = index
			}
			ValidateGlobalIndex(absolute_index)
			var a0, a1 = base.GlobalIndex(absolute_index)
			inst_seq[i].Arg0 = a0
			inst_seq[i].Arg1 = a1
		}
	}
}

