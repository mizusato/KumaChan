package checker

import (
	"kumachan/compiler/loader"
)


func Box (
	to_be_boxed  SemiExpr,
	g_type       *GenericType,
	g_type_name  loader.Symbol,
	g_type_info  ExprInfo,
	given_args   [] Type,
	force_exact  bool,
	info         ExprInfo,
	ctx          ExprContext,
) (Expr, *ExprError) {
	var boxed, is_boxed = g_type.Definition.(*Boxed)
	if !is_boxed {
		var _, is_enum = g_type.Definition.(*Enum)
		if is_enum {
			var typed_expr, is_typed = to_be_boxed.Value.(TypedExpr)
			if is_typed {
				var named, is_named = typed_expr.Type.(*NamedType)
				if is_named {
					var named_g = ctx.ModuleInfo.Types[named.Name]
					var c = named_g.CaseInfo
					if c.IsCaseType && c.EnumName == g_type_name {
						var expr = Expr(typed_expr)
						return LiftCase(c, named, false, expr), nil
					}
				}
				return Expr{}, &ExprError {
					Point:    to_be_boxed.Info.ErrorPoint,
					Concrete: E_NotCaseType {
						Type:  ctx.DescribeCertainType(typed_expr.Type),
						Enum: g_type_name.String(),
					},
				}
			} else {
				return Expr{}, &ExprError {
					Point:    to_be_boxed.Info.ErrorPoint,
					Concrete: E_ExplicitTypeRequired {},
				}
			}
		} else {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_BoxNonBoxedType { g_type_name.String() },
			}
		}
	}
	var current_mod = ctx.GetModuleName()
	var type_mod = g_type_name.ModuleName
	if current_mod != type_mod {
		if boxed.Protected {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_BoxProtectedType { g_type_name.String() },
			}
		}
		if boxed.Opaque {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_BoxOpaqueType { g_type_name.String() },
			}
		}
	}
	var given_count = uint(len(given_args))
	var g_type_arity = uint(len(g_type.Params))
	var max = g_type_arity
	var min = g_type_arity - uint(len(g_type.Defaults))
	var node = info.ErrorPoint.Node
	if min <= given_count && given_count <= max {
		var inner_type = FillTypeArgsWithDefaults(boxed.InnerType, given_args, g_type.Defaults)
		var expr, err = AssignTo(inner_type, to_be_boxed, ctx)
		if err != nil { return Expr{}, err }
		var outer_type = &NamedType {
			Name: g_type_name,
			Args: given_args,
		}
		err = CheckTypeArgsBounds(given_args, g_type.Params, g_type.Defaults, g_type.Bounds, node, ctx)
		if err != nil { return Expr{}, err }
		var case_info = g_type.CaseInfo
		return LiftCase(case_info, outer_type, force_exact, Expr {
			Type:  outer_type,
			Value: expr.Value,
			Info:  info,
		}), nil
	} else if given_count == 0 {
		var inf_ctx = ctx.WithInferringEnabled(g_type.Params, g_type.Bounds)
		var marked_inner_type = MarkParamsAsBeingInferred(boxed.InnerType)
		var expr, err = AssignTo(marked_inner_type, to_be_boxed, inf_ctx)
		if err != nil { return Expr{}, err }
		if uint(len(inf_ctx.Inferring.Arguments)) != g_type_arity {
			return Expr{}, &ExprError {
				Point:    g_type_info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {},
			}
		}
		var inferred_args = make([] Type, g_type_arity)
		for i, active := range inf_ctx.Inferring.Arguments {
			inferred_args[i] = active.CurrentValue
		}
		var inferred_type = &NamedType {
			Name: g_type_name,
			Args: inferred_args,
		}
		var inner_type = FillTypeArgs(boxed.InnerType, inferred_args)
		var _, ok = AssignType(inner_type, expr.Type, ToInferred, ctx)
		if !(ok) {
			panic("type system internal error (likely a bug)")
		}
		err = CheckTypeArgsBounds(inferred_args, g_type.Params, g_type.Defaults, g_type.Bounds, node, ctx)
		if err != nil { return Expr{}, err }
		var case_info = g_type.CaseInfo
		return LiftCase(case_info, inferred_type, force_exact, Expr {
			Type:  inferred_type,
			Value: expr.Value,
			Info:  info,
		}), nil
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

func GetEnumArgs (
	case_args       [] Type,
	enum_arity      uint,
	enum_variance   [] TypeVariance,
	mapping         [] uint,
) ([] Type) {
	var mapped = make([] Type, enum_arity)
	for i := uint(0); i < enum_arity; i += 1 {
		var v = enum_variance[i]
		switch v {
		case Covariant:
			mapped[i] = &NeverType {}
		case Contravariant:
			mapped[i] = &AnyType {}
		default:
			mapped[i] = &AnonymousType { Unit {} }
		}
	}
	for i, j := range mapping {
		mapped[j] = case_args[i]
	}
	return mapped
}

func LiftCase (
	case_info    CaseInfo,
	case_t       *NamedType,
	force_exact  bool,
	expr         Expr,
) Expr {
	if force_exact || !(case_info.IsCaseType) {
		return expr
	} else {
		var enum = case_info.EnumName
		var enum_arity = case_info.EnumArity
		var enum_variance = case_info.EnumVariance
		var mapping = case_info.CaseParams
		var index = case_info.CaseIndex
		var args = GetEnumArgs(case_t.Args, enum_arity, enum_variance, mapping)
		return Expr {
			Type:  &NamedType {
				Name: enum,
				Args: args,
			},
			Value: Sum {
				Value: expr,
				Index: index,
			},
			Info:  expr.Info,
		}
	}
}

