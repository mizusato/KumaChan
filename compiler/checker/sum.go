package checker

import (
	"fmt"
	"strings"
	. "kumachan/util/error"
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
)


func (impl Sum) ExprVal() {}
type Sum struct {
	Value  Expr
	Index  uint
}

func (impl UnitValue) ExprVal() {}
type UnitValue struct {}

func (impl SemiTypedSwitch) SemiExprVal() {}
type SemiTypedSwitch struct {
	Argument  Expr
	Branches  [] SemiTypedBranch
	Reactive  bool
}
type SemiTypedBranch struct {
	IsDefault  bool
	Index      uint
	Pattern    MaybePattern
	Value      SemiExpr
}
func (impl SemiTypedMultiSwitch) SemiExprVal() {}
type SemiTypedMultiSwitch struct {
	Arguments  [] Expr
	Branches   [] SemiTypedMultiBranch
}
type SemiTypedMultiBranch struct {
	IsDefault  bool
	Indexes    [] MaybeDefaultIndex
	Pattern    MaybePattern   // can only be TuplePattern or nil
	Value      SemiExpr
}
type MaybeDefaultIndex struct {
	IsDefault  bool
	Index      uint
}

func (impl Switch) ExprVal() {}
type Switch struct {
	Argument  Expr
	Branches  [] Branch
}
type Branch struct {
	IsDefault  bool
	Index      uint
	Pattern    MaybePattern
	Value      Expr
}
func (impl ReactiveSwitch) ExprVal() {}
type ReactiveSwitch struct {
	Argument  Expr
	Branches  Product
}
func (impl MultiSwitch) ExprVal() {}
type MultiSwitch struct {
	Arguments  [] Expr
	Branches   [] MultiBranch
}
type MultiBranch struct {
	IsDefault  bool
	Indexes    [] MaybeDefaultIndex
	Pattern    MaybePattern   // can only be TuplePattern or nil
	Value      Expr
}


func CheckSwitch(sw ast.Switch, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(sw.Node)
	var arg_semi, err1 = Check(sw.Argument, ctx)
	if err1 != nil { return SemiExpr{}, err1 }
	var arg_typed, err2 = AssignTo(nil, arg_semi, ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	var arg_type = arg_typed.Type
	var enum, enum_args, across_reactive, ok =
		ExtractEnum(arg_type, ctx, true)
	if !(ok) { return SemiExpr{}, &ExprError {
		Point:    arg_typed.Info.ErrorPoint,
		Concrete: E_InvalidSwitchArgType {
			ArgType: ctx.DescribeCertainType(arg_typed.Type),
		},
	} }
	var checked = make(map[loader.Symbol] bool)
	var has_default = false
	var default_node ast.Node
	var ast_branches = DesugarBranches(sw.Branches)
	var semi_branches = make([] SemiTypedBranch, len(ast_branches))
	for i, branch := range ast_branches {
		switch t := branch.Type.(type) {
		case ast.TypeRef:
			if len(t.TypeArgs) > 0 {
				return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(t.Node),
					Concrete: E_TypeParametersUnnecessary {},
				}
			}
			var maybe_type_sym = ctx.ModuleInfo.Module.SymbolFromTypeRef(t)
			var type_sym, ok = maybe_type_sym.(loader.Symbol)
			if !ok { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(t.Module.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ErrorPointFrom(t.Module.Node),
					Concrete: E_ModuleOfTypeRefNotFound {
						Name: ast.Id2String(t.Module),
					},
				} },
			}}
			var _, exists = ctx.ModuleInfo.Types[type_sym]
			if !exists { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(t.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ErrorPointFrom(t.Node),
					Concrete: E_TypeNotFound {
						Name: type_sym,
					},
				} },
			} }
			var index, case_args, is_case = GetCaseInfo (
				enum, enum_args, type_sym,
			)
			if !is_case { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(t.Node),
				Concrete: E_NotBranchType {
					Enum:     ctx.DescribeCertainType(arg_type),
					TypeName: type_sym.String(),
				},
			} }
			if checked[type_sym] { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(branch.Node),
				Concrete: E_CheckedBranch {},
			} }
			var case_type Type = &NamedType {
				Name: type_sym,
				Args: case_args,
			}
			if across_reactive {
				case_type = Reactive(case_type)
			}
			var maybe_pattern MaybePattern
			var branch_ctx ExprContext
			switch pattern_node := branch.Pattern.(type) {
			case ast.VariousPattern:
				var pattern, err = PatternFrom(pattern_node, case_type, ctx)
				if err != nil { return SemiExpr{}, err }
				maybe_pattern = pattern
				branch_ctx = ctx.WithPatternMatching(pattern)
			default:
				maybe_pattern = nil
				branch_ctx = ctx
			}
			var semi, err = Check(branch.Expr, branch_ctx)
			if err != nil { return SemiExpr{}, err }
			semi_branches[i] = SemiTypedBranch {
				IsDefault: false,
				Index:     index,
				Pattern:   maybe_pattern,
				Value:     semi,
			}
			checked[type_sym] = true
		default:
			if has_default {
				return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(branch.Node),
					Concrete: E_DuplicateDefaultBranch {},
				}
			}
			if branch.Pattern != nil { panic("something went wrong") }
			var value, err = Check(branch.Expr, ctx)
			if err != nil { return SemiExpr{}, err }
			semi_branches[i] = SemiTypedBranch {
				IsDefault: true,
				Index:     BadIndex,
				Pattern:   nil,
				Value:     value,
			}
			has_default = true
			default_node = branch.Node
		}
	}
	if !has_default && len(checked) != len(enum.CaseTypes) {
		var missing = make([]string, 0)
		for _, case_type := range enum.CaseTypes {
			if !checked[case_type.Name] {
				missing = append(missing, case_type.Name.String())
			}
		}
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(sw.Node),
			Concrete: E_IncompleteMatch { missing },
		}
	} else if has_default && len(checked) == len(enum.CaseTypes) {
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(default_node),
			Concrete: E_SuperfluousDefaultBranch {},
		}
	} else {
		return SemiExpr {
			Value: SemiTypedSwitch {
				Argument: Expr(arg_typed),
				Branches: semi_branches,
				Reactive: across_reactive,
			},
			Info: info,
		}, nil
	}
}

func CheckMultiSwitch(msw ast.MultiSwitch, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(msw.Node)
	var A = uint(len(msw.Arguments))
	var args = make([] Expr, A)
	var enums = make([] *Enum, A)
	var enums_args = make([][] Type, A)
	for i, arg_node := range msw.Arguments {
		var semi, err1 = Check(arg_node, ctx)
		if err1 != nil { return SemiExpr{}, err1 }
		var arg_typed, err2 = AssignTo(nil, semi, ctx)
		if err2 != nil { return SemiExpr{}, err2 }
		var arg_type = arg_typed.Type
		var enum, enum_args, _, is_enum = ExtractEnum(arg_type, ctx, false)
		if !is_enum { return SemiExpr{}, &ExprError {
			Point:    arg_typed.Info.ErrorPoint,
			Concrete: E_InvalidSwitchArgType {
				ArgType: ctx.DescribeCertainType(arg_typed.Type),
			},
		} }
		args[i] = Expr(arg_typed)
		enums[i] = enum
		enums_args[i] = enum_args
	}
	var checked = make(map[string] bool)
	var N = uint(1)
	for _, u := range enums {
		N *= uint(len(u.CaseTypes))
	}
	var has_default = false
	var default_node ast.Node
	var semi_branches = make([] SemiTypedMultiBranch, len(msw.Branches))
	for branch_index, branch := range msw.Branches {
		if len(branch.Types) > 0 {
			var num_types = uint(len(branch.Types))
			if num_types != A { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(branch.Node),
				Concrete: E_WrongMultiBranchTypeQuantity {
					Required: A,
					Given:    num_types,
				},
			} }
			var indexes = make([]uint, A)
			var types = make([]Type, A)
			var is_default = make([]bool, A)
			for i, t := range branch.Types {
				if len(t.TypeArgs) > 0 {
					return SemiExpr{}, &ExprError {
						Point:    ErrorPointFrom(t.Node),
						Concrete: E_TypeParametersUnnecessary {},
					}
				}
				var maybe_type_sym = ctx.ModuleInfo.Module.SymbolFromTypeRef(t)
				var type_sym, ok = maybe_type_sym.(loader.Symbol)
				if !ok { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(t.Module.Node),
					Concrete: E_TypeErrorInExpr { &TypeError {
						Point:    ErrorPointFrom(t.Module.Node),
						Concrete: E_ModuleOfTypeRefNotFound {
							Name: ast.Id2String(t.Module),
						},
					} },
				}}
				if (type_sym.SymbolName == IgnoreMark) {
					indexes[i] = BadIndex
					types[i] = &AnonymousType { Unit {} }
					is_default[i] = true
					continue
				}
				var _, exists = ctx.ModuleInfo.Types[type_sym]
				if !exists { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(t.Node),
					Concrete: E_TypeErrorInExpr { &TypeError {
						Point:    ErrorPointFrom(t.Node),
						Concrete: E_TypeNotFound {
							Name: type_sym,
						},
					} },
				} }
				var index, case_args, is_case = GetCaseInfo (
					enums[i], enums_args[i], type_sym,
				)
				if !is_case { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(t.Node),
					Concrete: E_NotBranchType {
						Enum:    ctx.DescribeCertainType(args[i].Type),
						TypeName: type_sym.String(),
					},
				} }
				var el_case_type = &NamedType {
					Name: type_sym,
					Args: case_args,
				}
				indexes[i] = index
				types[i] = el_case_type
				is_default[i] = false
			}
			var all_default = true
			for _, is_def := range is_default {
				var is_not_def = !(is_def)
				if is_not_def {
					all_default = false
					break
				}
			}
			if all_default { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(branch.Node),
				Concrete: E_MultiBranchTypesAllDefault {},
			} }
			var keys = make([]string, 0)
			var collect_keys func([]uint)
			collect_keys = func(path []uint) {
				if len(path) == len(indexes) {
					var buf strings.Builder
					for _, index := range path {
						buf.WriteString(fmt.Sprint(index))
						buf.WriteRune(' ')
					}
					var key = buf.String()
					keys = append(keys, key)
				} else {
					var i = len(path)
					if is_default[i] {
						var num_cases = uint(len(enums[i].CaseTypes))
						for j := uint(0); j < num_cases; j += 1 {
							var copied = make([]uint, len(path))
							copy(copied, path)
							collect_keys(append(copied, j))
						}
					} else {
						var copied = make([]uint, len(path))
						copy(copied, path)
						collect_keys(append(copied, indexes[i]))
					}
				}
			}
			collect_keys([]uint {})
			for _, key := range keys {
				if checked[key] { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(branch.Node),
					Concrete: E_CheckedBranch {},
				} }
				checked[key] = true
			}
			var case_type = &AnonymousType { Tuple { types } }
			var maybe_pattern MaybePattern
			var branch_ctx ExprContext
			switch pattern_node := branch.Pattern.(type) {
			case ast.PatternTuple:
				var adapted = ast.VariousPattern {
					Node:    pattern_node.Node,
					Pattern: pattern_node,
				}
				var pattern, err = PatternFrom(adapted, case_type, ctx)
				if err != nil { return SemiExpr{}, err }
				maybe_pattern = pattern
				branch_ctx = ctx.WithPatternMatching(pattern)
			default:
				maybe_pattern = nil
				branch_ctx = ctx
			}
			var indexes_info = make([] MaybeDefaultIndex, A)
			for i := uint(0); i < A; i += 1 {
				indexes_info[i] = MaybeDefaultIndex {
					IsDefault: is_default[i],
					Index:     indexes[i],
				}
			}
			var value, err = Check(branch.Expr, branch_ctx)
			if err != nil { return SemiExpr{}, err }
			semi_branches[branch_index] = SemiTypedMultiBranch {
				IsDefault: false,
				Indexes:   indexes_info,
				Pattern:   maybe_pattern,
				Value:     value,
			}
		} else {
			if has_default {
				return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(branch.Node),
					Concrete: E_DuplicateDefaultBranch {},
				}
			}
			if branch.Pattern != nil { panic("something went wrong") }
			var value, err = Check(branch.Expr, ctx)
			if err != nil { return SemiExpr{}, err }
			semi_branches[branch_index] = SemiTypedMultiBranch {
				IsDefault: true,
				Indexes:   nil,
				Pattern:   nil,
				Value:     value,
			}
			has_default = true
			default_node = branch.Node
		}
	}
	var num_checked = uint(len(checked))
	if num_checked > N { panic("something went wrong") }
	if !has_default && num_checked != N {
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(msw.Node),
			Concrete: E_IncompleteMultiMatch {
				MissingQuantity: (N - num_checked),
			},
		}
	} else if has_default && num_checked == N {
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(default_node),
			Concrete: E_SuperfluousDefaultBranch {},
		}
	} else {
		return SemiExpr {
			Value: SemiTypedMultiSwitch {
				Arguments: args,
				Branches:  semi_branches,
			},
			Info: info,
		}, nil
	}
}

func CheckIf(raw ast.If, ctx ExprContext) (SemiExpr, *ExprError) {
	var cond_assign = func(semi SemiExpr, ctx ExprContext) (Expr, bool, *ExprError) {
		var typed, err1 = AssignTo(__T_Bool, semi, ctx)
		if err1 == nil {
			return typed, false, nil
		} else {
			var typed, err = AssignTo(Reactive(__T_Bool), semi, ctx)
			if err == nil {
				return typed, true, nil
			} else {
				return Expr{}, false, err1
			}
		}
	}
	var if_node = DesugarElseIf(raw)
	var info = ctx.GetExprInfo(if_node.Node)
	var cond_semi, err1 = Check(if_node.Condition, ctx)
	if err1 != nil { return SemiExpr{}, err1 }
	var cond_typed, reactive, err2 = cond_assign(cond_semi, ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	var yes_semi, err3 = Check(if_node.YesBranch, ctx)
	if err3 != nil { return SemiExpr{}, err3 }
	var yes_branch = SemiTypedBranch {
		IsDefault: false,
		Index:     __Yes,
		Pattern:   nil,
		Value:     yes_semi,
	}
	var no_semi, err4 = Check(if_node.NoBranch, ctx)
	if err4 != nil { return SemiExpr{}, err4 }
	var no_branch = SemiTypedBranch {
		IsDefault: true,
		Index:     BadIndex,
		Pattern:   nil,
		Value:     no_semi,
	}
	return SemiExpr {
		Info: info,
		Value: SemiTypedSwitch {
			Argument: Expr(cond_typed),
			Branches: []SemiTypedBranch {
				yes_branch, no_branch,
			},
			Reactive: reactive,
		},
	}, nil
}

func DesugarBranches(raw_branches ([] ast.Branch)) ([] ast.Branch) {
	var branches = make([] ast.Branch, 0)
	for _, raw_branch := range raw_branches {
		if len(raw_branch.Types) == 0 {
			branches = append(branches, raw_branch)
			continue
		}
		for _, t := range raw_branch.Types {
			var branch = ast.Branch {
				Node:    raw_branch.Node,
				Type:    t,
				Types:   nil,
				Pattern: raw_branch.Pattern,
				Expr:    raw_branch.Expr,
			}
			branches = append(branches, branch)
		}
	}
	return branches
}


func AssignSwitchTo(expected Type, sw SemiTypedSwitch, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err1 = RequireExplicitType(expected, info)
	if err1 != nil { return Expr{}, err1 }
	if sw.Reactive {
		var _, ok = AssignType(expected, __AnyActionType, FromInferred, ctx)
		if !(ok) {
			return Expr{}, &ExprError {
				Point: info.ErrorPoint,
				Concrete: E_InvalidTypeForReactiveSwitch {
					Type: ctx.DescribeInferredType(expected),
				},
			}
		}
	}
	var branches = make([] Branch, len(sw.Branches))
	for i, branch_semi := range sw.Branches {
		var typed, err = AssignTo(expected, branch_semi.Value, ctx)
		if err != nil { return Expr{}, err }
		branches[i] = Branch {
			IsDefault: branch_semi.IsDefault,
			Index:     branch_semi.Index,
			Pattern:   branch_semi.Pattern,
			Value:     typed,
		}
	}
	var t, err2 = GetCertainType(expected, info.ErrorPoint, ctx)
	if err2 != nil { return Expr{}, err2 }
	if sw.Reactive {
		var elements = make([] Expr, len(branches))
		for i, b := range branches {
			var index Expr
			if b.IsDefault {
				index = Expr {
					Type:  nil,
					Value: UnitValue {},
					Info:  b.Value.Info,
				}
			} else {
				index = Expr {
					Type:  nil,
					Value: SmallIntLiteral { Value: b.Index },
					Info:  b.Value.Info,
				}
			}
			var pattern, ok = b.Pattern.(Pattern)
			if !(ok) {
				pattern = Pattern {
					Point:    b.Value.Info.ErrorPoint,
					Concrete: TrivialPattern {
						ValueName: IgnoreMark,
						ValueType: nil,
						Point:     b.Value.Info.ErrorPoint,
					},
				}
			}
			var pair = Product { Values: [] Expr {
				index, {
					Type: nil,
					Value: Lambda {
						Input:  pattern,
						Output: b.Value,
					},
				},
			} }
			elements[i] = Expr {
				Type:  nil,
				Value: pair,
				Info:  b.Value.Info,
			}
		}
		return Expr {
			Type:  t,
			Value: ReactiveSwitch {
				Argument: sw.Argument,
				Branches: Product { Values: elements },
			},
		}, nil
	} else {
		return Expr {
			Type:  t,
			Value: Switch {
				Argument: sw.Argument,
				Branches: branches,
			},
			Info:  info,
		}, nil
	}
}

func AssignMultiSwitchTo(expected Type, msw SemiTypedMultiSwitch, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err1 = RequireExplicitType(expected, info)
	if err1 != nil { return Expr{}, err1 }
	var branches = make([] MultiBranch, len(msw.Branches))
	for i, branch_semi := range msw.Branches {
		var typed, err = AssignTo(expected, branch_semi.Value, ctx)
		if err != nil { return Expr{}, err }
		branches[i] = MultiBranch {
			IsDefault: branch_semi.IsDefault,
			Indexes:   branch_semi.Indexes,
			Pattern:   branch_semi.Pattern,
			Value:     typed,
		}
	}
	var t, err2 = GetCertainType(expected, info.ErrorPoint, ctx)
	if err2 != nil { return Expr{}, err2 }
	return Expr {
		Type:  t,
		Value: MultiSwitch {
			Arguments: msw.Arguments,
			Branches:  branches,
		},
		Info:  info,
	}, nil
}


func ExtractEnum(t Type, ctx ExprContext, cross_reactive bool) (*Enum, []Type, bool, bool) {
	switch T := t.(type) {
	case *NamedType:
		if cross_reactive && IsReactive(T) {
			if !(len(T.Args) == 1) { panic("something went wrong") }
			var enum, args, _, ok = ExtractEnum(T.Args[0], ctx, false)
			if ok {
				return enum, args, true, true
			} else {
				return nil, nil, false, false
			}
		}
		var reg = ctx.GetTypeRegistry()
		var g = reg[T.Name]
		switch gv := g.Definition.(type) {
		case *Enum:
			return gv, T.Args, false, true
		case *Boxed:
			var ctx_mod = ctx.GetModuleName()
			var unboxed, can_unbox = Unbox(t, ctx_mod, reg).(Unboxed)
			if can_unbox {
				return ExtractEnum(unboxed.Type, ctx, cross_reactive)
			}
		}
	}
	return nil, nil, false, false
}

func DesugarElseIf(raw ast.If) ast.If {
	var no_branch = raw.NoBranch
	var elifs = raw.ElIfs
	for i, _ := range elifs {
		var elif = elifs[len(elifs)-1-i]
		var t = ast.If {
			Node:      elif.Node,
			Condition: elif.Condition,
			YesBranch: elif.YesBranch,
			NoBranch:  no_branch,
			ElIfs:     nil,
		}
		no_branch = ast.WrapTermAsExpr(ast.VariousTerm {
			Node: t.Node,
			Term: t,
		})
	}
	return ast.If {
		Node:      raw.Node,
		Condition: raw.Condition,
		YesBranch: raw.YesBranch,
		NoBranch:  no_branch,
		ElIfs:     nil,
	}
}

func GetMultiSwitchArgumentTuple(msw MultiSwitch, info ExprInfo) Expr {
	var el_types = make([] Type, len(msw.Arguments))
	for i, arg := range msw.Arguments {
		el_types[i] = arg.Type
	}
	return Expr {
		Type:  &AnonymousType { Tuple { el_types } },
		Value: Product { msw.Arguments },
		Info:  info,
	}
}

func GetCaseInfo(u *Enum, args []Type, sym loader.Symbol) (uint, []Type, bool) {
	for index, case_type := range u.CaseTypes {
		if case_type.Name == sym {
			var case_args = make([]Type, len(case_type.Params))
			for i, which_arg := range case_type.Params {
				case_args[i] = args[which_arg]
			}
			return uint(index), case_args, true
		}
	}
	return BadIndex, nil, false
}
