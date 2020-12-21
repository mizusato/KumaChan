package checker

import (
	"kumachan/parser/ast"
)


func CheckCast(cast ast.Cast, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(cast.Node)
	var type_ctx = ctx.GetTypeContext()
	var type_ref, is_type_ref = cast.Target.Type.(ast.TypeRef)
	if is_type_ref {
		if len(type_ref.TypeArgs) == 0 &&
			len(type_ref.Module.Name) == 0 &&
			string(type_ref.Id.Name) == SuperTypeName {
			var semi, err = Check(cast.Object, ctx)
			if err != nil { return SemiExpr{}, err }
			var typed, is_typed = semi.Value.(TypedExpr)
			if !(is_typed) { return SemiExpr{}, &ExprError {
				Point:    typed.Info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			} }
			var ctx_mod = ctx.ModuleInfo.Module.Name  // TODO: --> method
			var reg = ctx.ModuleInfo.Types
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
						Type: ctx.DescribeType(typed.Type),
					},
				}
			case UnboxFailed:
				return SemiExpr{}, &ExprError {
					Point:    typed.Info.ErrorPoint,
					Concrete: E_UnboxFailed {
						Type: ctx.DescribeType(typed.Type),
					},
				}
			default:
				panic("impossible branch")
			}
		}
	}
	var target, err1 = TypeFrom(cast.Target, type_ctx)
	if err1 != nil { return SemiExpr{}, &ExprError {
		Point:    err1.Point,
		Concrete: E_TypeErrorInExpr { err1 },
	} }
	var semi, err2 = Check(cast.Object, ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	var typed, err3 = AssignTo(target, semi, ctx)
	if err3 != nil { return SemiExpr{}, err3 }
	if !(AreTypesEqualInSameCtx(typed.Type, target)) {
		panic("something went wrong")
	}
	return LiftTyped(Expr {
		Type:  typed.Type,
		Value: typed.Value,
		Info:  info,
	}), nil
}
