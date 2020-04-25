package checker

import "kumachan/transformer/ast"
import . "kumachan/error"


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
	Pattern    Pattern
	Value      Expr
	Recursive  bool
}


func CheckBlock(block ast.Block, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(block.Node)
	var type_ctx = ctx.GetTypeContext()
	var current_ctx = ctx
	var bindings = make([]Binding, len(block.Bindings))
	for i, b := range block.Bindings {
		var t Type
		switch type_node := b.Type.(type) {
		case ast.VariousType:
			var some_t, err = TypeFrom(type_node.Type, type_ctx)
			if err != nil { return SemiExpr{}, &ExprError {
				Point:    err.Point,
				Concrete: E_TypeErrorInExpr { err },
			}}
			t = some_t
		default:
			t = nil
		}
		if b.Recursive {
			if t == nil {
				return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(b.Value.Node),
					Concrete: E_ExplicitTypeRequired {},
				}
			}
			var pattern, err1 = PatternFrom(b.Pattern, t, current_ctx)
			if err1 != nil { return SemiExpr{}, err1 }
			var rec_ctx = current_ctx.WithShadowingPatternMatching(pattern)
			var semi, err2 = Check(b.Value, rec_ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			var typed, err3 = AssignTo(t, semi, rec_ctx)
			if err3 != nil { return SemiExpr{}, err3 }
			switch typed.Value.(type) {
			case Lambda:
				bindings[i] = Binding {
					Pattern:   pattern,
					Value:     typed,
					Recursive: true,
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
			var non_nil_t = typed.Type
			var pattern, err3 = PatternFrom(b.Pattern, non_nil_t, current_ctx)
			if err3 != nil { return SemiExpr{}, err3 }
			var next_ctx = current_ctx.WithShadowingPatternMatching(pattern)
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
		Type:  ret_typed.Type,
		Info:  info,
		Value: Block {
			Bindings: bindings,
			Returned: ret_typed,
		},
	}, nil
}
