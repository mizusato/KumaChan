package checker

import (
	"kumachan/interpreter/lang/ast"
)


func CheckCast(object SemiExpr, target ast.VariousType, info ExprInfo, ctx ExprContext) (SemiExpr, *ExprError) {
	var type_ctx = ctx.GetTypeContext()
	var type_ref, is_type_ref = target.Type.(ast.TypeRef)
	if is_type_ref {
		if len(type_ref.TypeArgs) == 0 &&
			len(type_ref.Module.Name) == 0 &&
			string(type_ref.Id.Name) == SuperTypeName {
			var typed, is_typed = object.Value.(TypedExpr)
			if !(is_typed) { return SemiExpr{}, &ExprError {
				Point:    typed.Info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			} }
			var ctx_mod = ctx.GetModuleName()
			var reg = ctx.GetTypeRegistry()
			switch result := Unbox(typed.Type, ctx_mod, reg).(type) {
			case Unboxed:
				var t = result.Type
				return LiftTyped(Expr {
					Type:  t,
					Value: typed.Value,
					Info:  typed.Info,
				}), nil
			case UnboxedButOpaque:
				return SemiExpr{}, &ExprError {
					Point:    typed.Info.ErrorPoint,
					Concrete: E_UnboxOpaqueType {
						Type: ctx.DescribeCertainType(typed.Type),
					},
				}
			case UnboxFailed:
				return SemiExpr{}, &ExprError {
					Point:    typed.Info.ErrorPoint,
					Concrete: E_UnboxFailed {
						Type: ctx.DescribeCertainType(typed.Type),
					},
				}
			default:
				panic("impossible branch")
			}
		}
	}
	var target_t, err1 = TypeFrom(target, type_ctx)
	if err1 != nil { return SemiExpr{}, &ExprError {
		Point:    err1.Point,
		Concrete: E_TypeErrorInExpr { err1 },
	} }
	var typed, err2 = AssignTo(target_t, object, ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	return LiftTyped(Expr {
		Type:  target_t,
		Value: typed.Value,
		Info:  info,
	}), nil
}
