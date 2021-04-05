package generator

import (
	"fmt"
	"kumachan/misc/rpc"
	"kumachan/misc/rpc/kmd"
	"kumachan/lang"
	. "kumachan/misc/util/error"
)


func CreateProgram (
	metadata  lang.ProgramMetaData,
	idx       Index,
	data      [] lang.DataValue,
	closures  [] FuncNode,
	schema    kmd.SchemaTable,
	services  rpc.ServiceIndex,
) (lang.Program, DepLocator, E) {
	var no_locator = DepLocator {}
	var kmd_info = lang.KmdInfo {
		SchemaTable:       schema,
		KmdAdapterTable:   make(lang.KmdAdapterTable),
		KmdValidatorTable: make(lang.KmdValidatorTable),
	}
	var rpc_info = lang.RpcInfo{
		ServiceIndex: services,
	}
	var function_index_map = make(map[DepFunction] uint)
	var functions = make([] FuncNode, 0)
	var thunk_index_map = make(map[DepFunction] uint)
	var thunk_names = make([] string, 0)
	var thunks = make([] FuncNode, 0)
	for mod_name, mod := range idx {
		for f_name, items := range mod.Functions {
			for f_index, item := range items {
				var global_index = uint(len(functions))
				functions = append(functions, item)
				var dep = DepFunction {
					Module: mod_name,
					Name:   f_name,
					Index:  uint(f_index),
				}
				function_index_map[dep] = global_index
				if item.ConsideredThunk {
					var name = fmt.Sprintf("%s::%s", mod_name, f_name)
					thunk_index_map[dep] = uint(len(thunk_index_map))
					thunk_names = append(thunk_names, name)
					thunks = append(thunks, item)
				}
				if item.IsAdapter {
					kmd_info.KmdAdapterTable[item.AdapterId] = lang.KmdAdapterInfo {
						Index: global_index,
					}
				}
				if item.IsValidator {
					kmd_info.KmdValidatorTable[item.ValidatorId] = lang.KmdValidatorInfo {
						Index: global_index,
					}
				}
			}
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
	var get_thunk_index = func(d DepFunction) (uint, bool) {
		var index, exists = thunk_index_map[d]
		return index, exists
	}
	var thunk_dep_map = make([][] uint, len(thunks))
	for thunk_index, thunk := range thunks {
		var dep_indexes = make([] uint, 0)
		var visited_index_map = make(map[uint] bool)
		var collect_deps_from func(FuncNode)
		collect_deps_from = func(f FuncNode) {
			for _, f_dep := range f.Dependencies {
				switch concrete_f_dep := f_dep.(type) {
				case DepFunction:
					{
						var dep_index, this_is_thunk =
							get_thunk_index(concrete_f_dep)
						if this_is_thunk {
							dep_indexes = append(dep_indexes, dep_index)
							break
						}
					}
					var dep_f_index = get_function_index(concrete_f_dep)
					var _, exists = visited_index_map[dep_f_index]
					if !(exists) {
						visited_index_map[dep_f_index] = true
						var dep_f = functions[dep_f_index]
						if !(dep_f.HideFromCircularCheck) {
							collect_deps_from(dep_f)
						}
					}
				case DepClosure:
					var closure = closures[concrete_f_dep.Index]
					collect_deps_from(closure)
				default:
					// do nothing
				}
			}
		}
		collect_deps_from(thunk)
		thunk_dep_map[thunk_index] = dep_indexes
	}
	var L = uint(len(thunks))
	var in_degrees = make([] uint, L)
	var inv_map = make([][] uint, L)
	for i := uint(0); i < L; i += 1 {
		inv_map[i] = make([] uint, 0)
	}
	for i := uint(0); i < L; i += 1 {
		var deps = thunk_dep_map[i]
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
	var sorted_count = uint(0)
	for len(queue) > 0 {
		var i = queue[0]
		queue = queue[1:]
		sorted_count += 1
		for _, j := range inv_map[i] {
			if in_degrees[j] < 1 { panic("something went wrong") }
			in_degrees[j] -= 1
			if in_degrees[j] == 0 {
				queue = append(queue, j)
			}
		}
	}
	if sorted_count < L {
		var rest_names = make([] string, 0)
		var point ErrorPoint
		for i := uint(0); i < L; i += 1 {
			if in_degrees[i] > 0 {
				rest_names = append(rest_names, thunk_names[i])
				point = thunks[i].Underlying.Info.DeclPoint
			}
		}
		if len(rest_names) == 0 { panic("something went wrong") }
		return lang.Program{}, no_locator, &Error {
			Point:    point,
			Concrete: E_CircularThunkDependency { rest_names },
		}
	}
	var base_data = uint(0)
	var base_function = base_data + uint(len(data))
	var base_closure = base_function + uint(len(functions))
	var base_extra = base_closure + uint(len(closures))
	var get_dep_addr = func(dep Dependency) (uint, bool) {
		switch d := dep.(type) {
		case DepData:
			return base_data + d.Index, true
		case DepFunction:
			return base_function + function_index_map[d], true
		case DepClosure:
			return base_closure + d.Index, true
		default:
			return ^uint(0), false
		}
	}
	var locator = DepLocator {
		Locate: get_dep_addr,
		Offset: base_extra,
	}
	var relocate_code = func(f *FuncNode) {
		RelocateCode(f, locator, nil)
	}
	for i, _ := range functions {
		relocate_code(&(functions[i]))
	}
	for i, _ := range closures {
		relocate_code(&(closures[i]))
	}
	for i, _ := range effects {
		relocate_code(&(effects[i]))
	}
	var unwrap = func(list []FuncNode) []*lang.Function {
		var raw = make([]*lang.Function, len(list))
		for i, item := range list {
			raw[i] = item.Underlying
		}
		return raw
	}
	return lang.Program {
		MetaData:   metadata,
		DataValues: data,
		Functions:  unwrap(functions),
		Closures:   unwrap(closures),
		Effects:    unwrap(effects),
		KmdInfo:    kmd_info,
		RpcInfo:    rpc_info,
	}, locator, nil
}

