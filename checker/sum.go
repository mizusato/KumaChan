package checker

import (
	"kumachan/loader"
	"kumachan/transformer/ast"
)


func (impl Sum) ExprVal() {}
type Sum struct {
	Value  Expr
	Index  uint
}

func (impl UnitValue) ExprVal() {}
type UnitValue struct {}

func (impl SemiTypedMatch) SemiExprVal() {}
type SemiTypedMatch struct {
	Argument  Expr
	Branches  [] SemiTypedBranch
}
type SemiTypedBranch struct {
	IsDefault  bool
	Index      uint
	Pattern    MaybePattern
	Value      SemiExpr
}

func (impl Match) ExprVal() {}
type Match struct {
	Argument  Expr
	Branches  [] Branch
}
type Branch struct {
	IsDefault  bool
	Index      uint
	Pattern    MaybePattern
	Value      Expr
}


func CheckSwitch(sw ast.Switch, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(sw.Node)
	var arg_semi, err = CheckTerm(sw.Argument, ctx)
	if err != nil { return SemiExpr{}, err }
	var arg_typed, ok = arg_semi.Value.(TypedExpr)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(sw.Argument.Node),
		Concrete: E_ExplicitTypeRequired {},
	} }
	var arg_type = arg_typed.Type
	var union, union_args, is_union = ExtractUnion(arg_type, ctx)
	if !is_union { return SemiExpr{}, &ExprError {
		Point:    arg_typed.Info.ErrorPoint,
		Concrete: E_InvalidMatchArgType {
			ArgType: ctx.DescribeType(arg_typed.Type),
		},
	} }
	var checked = make(map[loader.Symbol] bool)
	var has_default = false
	var branches = make([]SemiTypedBranch, len(sw.Branches))
	for i, branch := range sw.Branches {
		switch t := branch.Type.(type) {
		case ast.Ref:
			if len(t.TypeArgs) > 0 {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(t.Node),
					Concrete: E_TypeParametersUnnecessary {},
				}
			}
			var maybe_type_sym = ctx.ModuleInfo.Module.SymbolFromRef(t)
			var type_sym, ok = maybe_type_sym.(loader.Symbol)
			if !ok { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(t.Module.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ctx.GetErrorPoint(t.Module.Node),
					Concrete: E_ModuleOfTypeRefNotFound {
						Name: loader.Id2String(t.Module),
					},
				} },
			}}
			var g, exists = ctx.ModuleInfo.Types[type_sym]
			if !exists { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(t.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ctx.GetErrorPoint(t.Node),
					Concrete: E_TypeNotFound {
						Name: type_sym,
					},
				} },
			} }
			var index, is_case = GetCaseIndex(union, type_sym)
			if !is_case { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(t.Node),
				Concrete: E_NotBranchType {
					Union:    ctx.DescribeType(arg_type),
					TypeName: type_sym.String(),
				},
			} }
			if len(g.Params) != len(union_args) {
				panic("something went wrong")
			}
			var case_type = NamedType {
				Name: type_sym,
				Args: union_args,
			}
			var maybe_pattern MaybePattern
			var branch_ctx ExprContext
			switch pattern_node := branch.Pattern.(type) {
			case ast.VariousPattern:
				var pattern, err = PatternFrom(pattern_node, case_type, ctx)
				if err != nil { return SemiExpr{}, err }
				maybe_pattern = pattern
				branch_ctx = ctx.WithShadowingPatternMatching(pattern)
			default:
				maybe_pattern = nil
				branch_ctx = ctx
			}
			var semi, err = Check(branch.Expr, branch_ctx)
			if err != nil { return SemiExpr{}, err }
			branches[i] = SemiTypedBranch {
				IsDefault: false,
				Index:     index,
				Pattern:   maybe_pattern,
				Value:     semi,
			}
			checked[type_sym] = true
		default:
			if has_default {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(branch.Node),
					Concrete: E_DuplicateDefaultBranch {},
				}
			}
			if branch.Pattern != nil { panic("something went wrong") }
			var semi, err = Check(branch.Expr, ctx)
			if err != nil { return SemiExpr{}, err }
			branches[i] = SemiTypedBranch {
				IsDefault: true,
				Index:     BadIndex,
				Pattern:   Pattern {},
				Value:     semi,
			}
			has_default = true
		}
	}
	if !has_default && len(checked) != len(union.CaseTypes) {
		var missing = make([]string, 0)
		for _, case_type := range union.CaseTypes {
			if !checked[case_type] {
				missing = append(missing, case_type.String())
			}
		}
		return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(sw.Node),
			Concrete: E_IncompleteMatch { missing },
		}
	} else {
		return SemiExpr {
			Value: SemiTypedMatch {
				Argument: Expr(arg_typed),
				Branches: branches,
			},
			Info: info,
		}, nil
	}
}

func CheckIf(raw ast.If, ctx ExprContext) (SemiExpr, *ExprError) {
	var if_node = DesugarElseIf(raw)
	var info = ctx.GetExprInfo(if_node.Node)
	var cond_semi, err1 = CheckTerm(if_node.Condition, ctx)
	if err1 != nil { return SemiExpr{}, err1 }
	var cond_typed, err2 = AssignTo(__T_Bool, cond_semi, ctx)
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
		Value: SemiTypedMatch {
			Argument: Expr(cond_typed),
			Branches: []SemiTypedBranch {
				yes_branch, no_branch,
			},
		},
	}, nil
}


func AssignMatchTo(expected Type, match SemiTypedMatch, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	var branches = make([]Branch, len(match.Branches))
	for i, branch_semi := range match.Branches {
		var typed, err = AssignTo(expected, branch_semi.Value, ctx)
		if err != nil { return Expr{}, err }
		branches[i] = Branch {
			IsDefault: branch_semi.IsDefault,
			Index:     branch_semi.Index,
			Pattern:   branch_semi.Pattern,
			Value:     typed,
		}
	}
	return Expr {
		Type:  expected,
		Value: Match {
			Argument: match.Argument,
			Branches: branches,
		},
		Info:  info,
	}, nil
}



func ExtractUnion(t Type, ctx ExprContext) (Union, []Type, bool) {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Union:
			return gv, T.Args, true
		}
	}
	return Union{}, nil, false
}

func ExtractUnionTuple(t Type, ctx ExprContext) ([]Union, [][]Type, bool) {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch tuple := inner_type.Repr.(type) {
				case Tuple:
					var union_types = make([]Union, len(tuple.Elements))
					var union_args = make([][]Type, len(tuple.Elements))
					for i, el := range tuple.Elements {
						switch el_t := el.(type) {
						case NamedType:
							var el_g = ctx.ModuleInfo.Types[el_t.Name]
							switch el_gv := el_g.Value.(type) {
							case Union:
								union_types[i] = el_gv
								union_args[i] = el_t.Args
								continue
							}
						}
						return nil, nil, false
					}
					return union_types, union_args, true
				}
			}
		}
	}
	return nil, nil, false
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
		no_branch = ast.Expr {
			Node:     t.Node,
			Call:     ast.Call {
				Node: t.Node,
				Func: ast.VariousTerm {
					Node: t.Node,
					Term: t,
				},
				Arg:  nil,
			},
			Pipeline: nil,
		}
	}
	return ast.If {
		Node:      raw.Node,
		Condition: raw.Condition,
		YesBranch: raw.YesBranch,
		NoBranch:  no_branch,
		ElIfs:     nil,
	}
}
