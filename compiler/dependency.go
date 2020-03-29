package compiler

import (
	"fmt"
	. "kumachan/error"
	c "kumachan/runtime/common"
	ch "kumachan/checker"
)


type Index  map[string] *CompiledModule
type CompiledModule struct {
	Functions   map[string] ([] FuncNode)
	Constants   map[string] FuncNode
	Effects     [] FuncNode
}
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

func CompileModule (
	mod       *ch.CheckedModule,
	idx       Index,
	data      *([] c.DataValue),
	closures  *([] FuncNode),
) *Error {
	var _, exists = idx[mod.Name]
	if exists {
		return nil
	}
	for _, imported := range mod.Imported {
		var err = CompileModule(imported, idx, data, closures)
		if err != nil { return err }
	}
	var functions = make(map[string] ([] FuncNode))
	var constants = make(map[string] FuncNode)
	var effects = make([] FuncNode, 0)
	for name, instances := range mod.Functions {
		for _, item := range instances {
			var f_raw, refs, err = CompileFunction(item.Body, name, item.Point)
			if err != nil { return err }
			var f = FuncNodeFrom(f_raw, refs, idx, data, closures)
			var existing, exists = functions[name]
			if exists {
				functions[name] = append(existing, f)
			} else {
				functions[name] = [] FuncNode { f }
			}
		}
	}
	for name, item := range mod.Constants {
		var f, refs, err = CompileFunction(item.Value, name, item.Point)
		if err != nil { return err }
		constants[name] = FuncNodeFrom(f, refs, idx, data, closures)
	}
	for _, item := range mod.Effects {
		var value = ch.ExprExpr(item.Value)
		var f, refs, err = CompileFunction(value, "(do)", item.Point)
		if err != nil { return err }
		effects = append(effects, FuncNodeFrom(f, refs, idx, data, closures))
	}
	idx[mod.Name] = &CompiledModule {
		Functions: functions,
		Constants: constants,
		Effects:   effects,
	}
	return nil
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

func CreateProgram (
	idx       Index,
	data      [] c.DataValue,
	closures  [] FuncNode,
) (c.Program, *Error) {
	var function_index_map = make(map[DepFunction] uint)
	var functions = make([] FuncNode, 0)
	for mod_name, mod := range idx {
		for f_name, items := range mod.Functions {
			for f_index, item := range items {
				var global_index = uint(len(functions))
				functions = append(functions, item)
				function_index_map[DepFunction {
					Module: mod_name,
					Name:   f_name,
					Index:  uint(f_index),
				}] = global_index
			}
		}
	}
	var constant_index_map = make(map[DepConstant] uint)
	var constant_names = make([] string, 0)
	var constants = make([] FuncNode, 0)
	for mod_name, mod := range idx {
		for item_name, item := range mod.Constants {
			var global_index = uint(len(constants))
			constants = append(constants, item)
			constant_names = append(constant_names,
				fmt.Sprintf("%s::%s", mod_name, item_name))
			constant_index_map[DepConstant {
				Module: mod_name,
				Name:   item_name,
			}] = global_index
		}
	}
	var effects = make([] FuncNode, 0)
	for _, mod := range idx {
		effects = append(effects, mod.Effects...)
	}
	var get_function_index = func(d DepFunction) uint {
		var index, exists = function_index_map[d]
		if exists {
			return index
		} else {
			panic("something went wrong")
		}
	}
	var get_constant_index = func(d DepConstant) uint {
		var index, exists = constant_index_map[d]
		if exists {
			return index
		} else {
			panic("something went wrong")
		}
	}
	var constant_dep_map = make([][] uint, len(constants))
	for constant_index, constant := range constants {
		var dep_indexes = make([] uint, 0)
		for _, dep := range constant.Dependencies {
			switch d := dep.(type) {
			case DepConstant:
				dep_indexes = append(dep_indexes, get_constant_index(d))
			case DepFunction:
				var visited_index_map = make(map[uint] bool)
				var collect_indirect func(FuncNode)
				collect_indirect = func(f FuncNode) {
					for _, f_dep := range f.Dependencies {
						switch concrete_f_dep := f_dep.(type) {
						case DepConstant:
							dep_indexes = append(dep_indexes,
								get_constant_index(concrete_f_dep))
						case DepFunction:
							var f_index = get_function_index(concrete_f_dep)
							var _, exists = visited_index_map[f_index]
							if exists { return }
							visited_index_map[f_index] = true
							collect_indirect(functions[f_index])
						case DepClosure:
							var closure = closures[concrete_f_dep.Index]
							collect_indirect(closure)
						default:
							// do nothing
						}
					}
				}
				var f_index = get_function_index(d)
				visited_index_map[f_index] = true
				collect_indirect(functions[f_index])
			}
		}
		constant_dep_map[constant_index] = dep_indexes
	}
	var L = uint(len(constants))
	var in_degrees = make([] uint, L)
	var inv_map = make([][] uint, L)
	for i := uint(0); i < L; i += 1 {
		inv_map[i] = make([] uint, 0)
	}
	for i := uint(0); i < L; i += 1 {
		var deps = constant_dep_map[i]
		in_degrees[i] = uint(len(deps))
		for _, dep := range deps {
			inv_map[dep] = append(inv_map[dep], i)
		}
	}
	var queue = make([] uint, 0)
	for i := uint(0); i < L; i += 1 {
		if in_degrees[i] == 0 {
			queue = append(queue, i)
		}
	}
	var sorted = make([] uint, 0, L)
	for len(queue) > 0 {
		var i = queue[0]
		queue = queue[1:]
		sorted = append(sorted, i)
		for _, j := range inv_map[i] {
			if in_degrees[j] < 1 { panic("something went wrong") }
			in_degrees[j] -= 1
			if in_degrees[j] == 0 {
				queue = append(queue, j)
			}
		}
	}
	if uint(len(sorted)) < L {
		var rest_names = make([] string, 0)
		var point ErrorPoint
		for i := uint(0); i < L; i += 1 {
			if in_degrees[i] > 0 {
				rest_names = append(rest_names, constant_names[i])
				point = constants[i].Underlying.Info.DeclPoint
			}
		}
		if len(rest_names) == 0 { panic("something went wrong") }
		return c.Program{}, &Error{
			Point:    point,
			Concrete: E_CircularConstantDependency { rest_names },
		}
	}
	var sorted_constants = make([] FuncNode, L)
	for i := uint(0); i < L; i += 1 {
		sorted_constants[i] = constants[sorted[i]]
	}
	var base_data = uint(0)
	var base_function = base_data + uint(len(data))
	var base_closure = base_function + uint(len(functions))
	var base_constant = base_closure + uint(len(closures))
	var get_dep_addr = func(dep Dependency) uint {
		switch d := dep.(type) {
		case DepData:
			return base_data + d.Index
		case DepFunction:
			return base_function + function_index_map[d]
		case DepClosure:
			return base_closure + d.Index
		case DepConstant:
			return base_constant + constant_index_map[d]
		default:
			panic("impossible branch")
		}
	}
	var relocate_code = func(f *FuncNode) {
		var inst_seq = f.Underlying.Code
		for i, _ := range inst_seq {
			if inst_seq[i].OpCode == c.GLOBAL {
				var relative_index = inst_seq[i].GetGlobalIndex()
				var dep = f.Dependencies[relative_index]
				var absolute_index = get_dep_addr(dep)
				ValidateGlobalIndex(absolute_index)
				inst_seq[i].Arg1 = c.Long(absolute_index)
			}
		}
	}
	for i, _ := range functions {
		relocate_code(&(functions[i]))
	}
	for i, _ := range closures {
		relocate_code(&(closures[i]))
	}
	for i, _ := range sorted_constants {
		relocate_code(&(sorted_constants[i]))
	}
	for i, _ := range effects {
		relocate_code(&(effects[i]))
	}
	var unwrap = func(list []FuncNode) []*c.Function {
		var raw = make([]*c.Function, len(list))
		for i, item := range list {
			raw[i] = item.Underlying
		}
		return raw
	}
	return c.Program {
		DataValues: data,
		Functions:  unwrap(functions),
		Closures:   unwrap(closures),
		Constants:  unwrap(sorted_constants),
		Effects:    unwrap(effects),
	}, nil
}
