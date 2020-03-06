package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
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


func CheckMatch(match node.Match, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(match.Node) }
	var arg_semi, err = Check(match.Argument, ctx)
	if err != nil { return SemiExpr{}, err }
	var arg_typed, ok = arg_semi.Value.(TypedExpr)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(match.Argument.Node),
		Concrete: E_ExplicitTypeRequired {},
	} }
	var arg_type = arg_typed.Type
	var union, union_args, is_union = UnboxUnion(arg_type, ctx)
	if !is_union { return SemiExpr{}, &ExprError {
		Point:    arg_typed.Info.ErrorPoint,
		Concrete: E_InvalidMatchArgType {
			ArgType: ctx.DescribeType(arg_typed.Type),
		},
	} }
	var checked = make(map[loader.Symbol]bool)
	var has_default = false
	var branches = make([]SemiTypedBranch, len(match.Branches))
	for i, branch := range match.Branches {
		switch t := branch.Type.(type) {
		case node.Ref:
			if len(t.TypeArgs) > 0 {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(t.Node),
					Concrete: E_TypeParametersUnnecessary {},
				}
			}
			var maybe_type_sym = ctx.ModuleInfo.Module.SymbolFromRef(t)
			var maybe_pattern = MaybePatternFrom(branch.Pattern, ctx)
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
			var index, is_subtype = union.GetSubtypeIndex(type_sym)
			if !is_subtype { return SemiExpr{}, &ExprError{
				Point:    ctx.GetErrorPoint(t.Node),
				Concrete: E_NotSubtype {
					Union:    ctx.DescribeType(arg_type),
					TypeName: type_sym.String(),
				},
			} }
			if g.Arity != uint(len(union_args)) {
				panic("something went wrong")
			}
			var subtype = NamedType {
				Name: type_sym,
				Args: union_args,
			}
			var branch_ctx ExprContext
			switch pattern := maybe_pattern.(type) {
			case Pattern:
				var new_ctx, err = ctx.WithPatternMatching (
					subtype, pattern, false,
				)
				if err != nil { return SemiExpr{}, err }
				branch_ctx = new_ctx
			default:
				branch_ctx = ctx
			}
			var semi, err = Check(
				branch.Expr, branch_ctx,
			)
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
			switch branch.Pattern.(type) {
			case node.VariousPattern:
				panic("something went wrong")
			}
			var semi, err = Check(branch.Expr, ctx)
			if err != nil { return SemiExpr{}, nil }
			branches[i] = SemiTypedBranch {
				IsDefault: true,
				Index:     BadIndex,
				Pattern:   Pattern {},
				Value:     semi,
			}
			has_default = true
		}
	}
	if !has_default && len(checked) != len(union.SubTypes) {
		var missing = make([]string, 0)
		for _, subtype := range union.SubTypes {
			if !checked[subtype] {
				missing = append(missing, subtype.String())
			}
		}
		return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(match.Node),
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

func CheckIf(if_node node.If, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(if_node.Node) }
	var cond_semi, err = Check(if_node.Condition, ctx)
	if err != nil { return SemiExpr{}, err }
	var cond_typed, ok = cond_semi.Value.(TypedExpr)
	if !ok { return SemiExpr{}, &ExprError{
		Point:    ctx.GetErrorPoint(if_node.Condition.Node),
		Concrete: E_NonBooleanCondition { Typed:false },
	} }
	switch T := cond_typed.Type.(type) {
	case NamedType:
		if T.Name == __Bool {
			if len(T.Args) != 0 { panic("something went wrong") }
			var yes_semi, err1 = Check(if_node.YesBranch, ctx)
			if err1 != nil { return SemiExpr{}, err1 }
			var yes_branch = SemiTypedBranch {
				IsDefault: false,
				Index:     __Yes,
				Pattern:   nil,
				Value:     yes_semi,
			}
			var no_semi, err2 = Check(if_node.NoBranch, ctx)
			if err2 != nil { return SemiExpr{}, err2 }
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
	}
	return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(if_node.Condition.Node),
		Concrete: E_NonBooleanCondition {
			Typed: true,
			Type:  ctx.DescribeType(cond_typed.Type),
		},
	}
}


func AssignMatchTo(expected Type, match SemiTypedMatch, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
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
