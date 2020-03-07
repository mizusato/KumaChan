package checker

import "kumachan/transformer/node"


func (impl SemiTypedBlock) SemiExprVal() {}
type SemiTypedBlock struct {
	Bindings  [] Binding
	Returned  SemiExpr
}

func (impl Block) ExprVal() {}
type Block struct {
	Bindings  [] Binding
	Returned  Expr
}
type Binding struct {
	Pattern  Pattern
	Value    Expr
}


func CheckBlock(block node.Block, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(block.Node)
	var type_ctx = ctx.GetTypeContext()
	var current_ctx = ctx
	var bindings = make([]Binding, len(block.Bindings))
	for i, b := range block.Bindings {
		var pattern = PatternFrom(b.Pattern, current_ctx)
		var t Type
		switch type_node := b.Type.(type) {
		case node.VariousType:
			var some_t, err = TypeFrom(type_node.Type, type_ctx)
			if err != nil { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(type_node.Node),
				Concrete: E_TypeErrorInExpr { err },
			}}
			t = some_t
		default:
			t = nil
		}
		if b.Recursive {
			if t == nil {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(b.Value.Node),
					Concrete: E_ExplicitTypeRequired {},
				}
			}
			var rec_ctx, err1 = current_ctx.WithPatternMatching (
				t, pattern, true,
			)
			if err1 != nil { return SemiExpr{}, err1 }
			var semi, err2 = Check(b.Value, rec_ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			switch semi.Value.(type) {
			case UntypedLambda:
				var typed, err = AssignTo(t, semi, rec_ctx)
				if err != nil { return SemiExpr{}, err }
				bindings[i] = Binding {
					Pattern: pattern,
					Value:   typed,
				}
				current_ctx = rec_ctx
			default:
				return SemiExpr{}, &ExprError {
					Point:    semi.Info.ErrorPoint,
					Concrete: E_RecursiveMarkUsedOnNonLambda {},
				}
			}
		} else {
			var semi, err1 = Check(b.Value, current_ctx)
			if err1 != nil { return SemiExpr{}, err1 }
			var typed, err2 = AssignTo(t, semi, current_ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			var final_t = typed.Type
			var next_ctx, err3 = current_ctx.WithPatternMatching (
				final_t, pattern, true,
			)
			if err3 != nil { return SemiExpr{}, err3 }
			bindings[i] = Binding {
				Pattern: pattern,
				Value:   typed,
			}
			current_ctx = next_ctx
		}
	}
	var ret, err = Check(block.Return, current_ctx)
	if err != nil { return SemiExpr{}, err }
	var ret_typed, is_typed = ret.Value.(TypedExpr)
	if is_typed {
		return LiftTyped(Expr {
			Type:  ret_typed.Type,
			Value: Block {
				Bindings: bindings,
				Returned: Expr(ret_typed),
			},
			Info:  info,
		}), nil
	} else {
		return SemiExpr {
			Value: SemiTypedBlock {
				Bindings: bindings,
				Returned: ret,
			},
			Info: info,
		}, nil
	}
}


func AssignBlockTo(expected Type, block SemiTypedBlock, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var bindings = block.Bindings
	var ret_semi = block.Returned
	var ret_typed, err = AssignTo(expected, ret_semi, ctx)
	if err != nil { return Expr{}, err }
	return Expr {
		Type:  expected,
		Info:  info,
		Value: Block {
			Bindings: bindings,
			Returned: ret_typed,
		},
	}, nil
}
