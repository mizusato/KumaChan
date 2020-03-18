package checker

import (
	"kumachan/loader"
)


func Box (
	to_be_boxed  SemiExpr,
	g_type       *GenericType,
	g_type_name  loader.Symbol,
	g_type_info  ExprInfo,
	given_args   [] Type,
	info         ExprInfo,
	ctx          ExprContext,
) (Expr, *ExprError) {
	var wrapped, ok = g_type.Value.(Boxed)
	if !ok { return Expr{}, &ExprError {
		Point:    g_type_info.ErrorPoint,
		Concrete: E_BoxNonBoxedType { g_type_name.String() },
	} }
	var current_mod = ctx.GetModuleName()
	var type_mod = g_type_name.ModuleName
	if current_mod != type_mod {
		if wrapped.Protected {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_BoxProtectedType { g_type_name.String() },
			}
		}
		if wrapped.Opaque {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_BoxOpaqueType { g_type_name.String() },
			}
		}
	}
	var given_count = uint(len(given_args))
	var g_type_arity = uint(len(g_type.Params))
	if given_count == g_type_arity {
		var inner_type = FillTypeArgs(wrapped.InnerType, given_args)
		var expr, err = AssignTo(inner_type, to_be_boxed, ctx)
		if err != nil { return Expr{}, err }
		var outer_type = NamedType {
			Name: g_type_name,
			Args: given_args,
		}
		return Expr {
			Type:  outer_type,
			Value: expr.Value,
			Info:  info,
		}, nil
	} else if given_count == 0 {
		var inf_ctx = ctx.WithTypeArgsInferringEnabled(g_type.Params)
		var marked_inner_type = MarkParamsAsBeingInferred(wrapped.InnerType)
		var expr, err = AssignTo(marked_inner_type, to_be_boxed, inf_ctx)
		if err != nil { return Expr{}, err }
		if uint(len(inf_ctx.Inferred)) != g_type_arity {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {},
			}
		}
		var inferred_args = make([]Type, g_type_arity)
		for i, t := range inf_ctx.Inferred {
			inferred_args[i] = t
		}
		var inferred_type = NamedType {
			Name: g_type_name,
			Args: inferred_args,
		}
		var inner_type = FillTypeArgs(wrapped.InnerType, inferred_args)
		if !(AreTypesEqualInSameCtx(inner_type, expr.Type)) {
			panic("something went wrong")
		}
		return Expr {
			Type:  inferred_type,
			Value: expr.Value,
			Info:  info,
		}, nil
	} else {
		return Expr{}, &ExprError {
			Point:    g_type_info.ErrorPoint,
			Concrete: E_TypeErrorInExpr { &TypeError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_WrongParameterQuantity {
					TypeName: g_type_name,
					Required: g_type_arity,
					Given:    given_count,
				},
			} },
		}
	}
}
