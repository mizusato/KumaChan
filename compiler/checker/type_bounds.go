package checker

import (
	"kumachan/compiler/loader/parser/ast"
	. "kumachan/util/error"
)


type TypeBounds struct {
	Sub    map[uint] Type
	Super  map[uint] Type
}

type TypeBoundsContext struct {
	TypeValidationContext
	Bounds  TypeBounds
}

type TypeBoundKind rune
const (
	SubBound    TypeBoundKind  =  '>'
	SuperBound  TypeBoundKind  =  '<'
)

func CheckTypeValBounds(val TypeDef, info TypeNodeInfo, ctx TypeBoundsContext) *TypeError {
	switch V := val.(type) {
	case *Enum:
		var union_b = ctx.Bounds
		for _, case_t := range V.CaseTypes {
			var case_g = ctx.Registry[case_t.Name]
			var case_b = case_g.Bounds
			var info = case_g.CaseInfo
			if !(info.IsCaseType) { panic("something went wrong") }
			for i_, j := range info.CaseParams {
				var i = uint(i_)
				if case_b.Sub[i] != union_b.Sub[j] {
					return &TypeError {
						Point:    ErrorPointFrom(case_g.Node),
						Concrete: E_CaseBadBounds {
							CaseName:  case_t.Name.String(),
							UnionName: info.UnionName.String(),
						},
					}
				}
				if case_b.Super[i] != union_b.Super[j] {
					return &TypeError {
						Point:    ErrorPointFrom(case_g.Node),
						Concrete: E_CaseBadBounds {
							CaseName:  case_t.Name.String(),
							UnionName: info.UnionName.String(),
						},
					}
				}
			}
		}
		return nil
	case *Boxed:
		return CheckTypeBounds(V.InnerType, info.TypeNodeMap, ctx)
	case *Native:
		return nil
	default:
		panic("impossible branch")
	}
}

func CheckTypeBounds(t Type, nodes (map[Type] ast.Node), ctx TypeBoundsContext) *TypeError {
	var get_node = func(t Type) ast.Node {
		return nodes[t]
	}
	switch T := t.(type) {
	case *NeverType:
		return nil
	case *AnyType:
		return nil
	case *ParameterType:
		return nil
	case *NamedType:
		var g, exists = ctx.Registry[T.Name]
		if !(exists) {
			// refers to a type that does not exist
			return nil
		}
		var L = uint(len(T.Args))
		for i, super := range g.Bounds.Super {
			if i < L {
				var arg = T.Args[i]
				var err = CheckTypeArgBound(arg, super, SuperBound, get_node, ctx)
				if err != nil { return err }
			}
		}
		for i, sub := range g.Bounds.Sub {
			if i < L {
				var arg = T.Args[i]
				var err = CheckTypeArgBound(arg, sub, SubBound, get_node, ctx)
				if err != nil { return err }
			}
		}
		return nil
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Unit:
			return nil
		case Tuple:
			for _, el := range R.Elements {
				var err = CheckTypeBounds(el, nodes, ctx)
				if err != nil { return err }
			}
			return nil
		case Bundle:
			for _, f := range R.Fields {
				var err = CheckTypeBounds(f.Type, nodes, ctx)
				if err != nil { return err }
			}
			return nil
		case Func:
			var err = CheckTypeBounds(R.Input, nodes, ctx)
			if err != nil { return err }
			err = CheckTypeBounds(R.Output, nodes, ctx)
			if err != nil { return err }
			return nil
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func CheckTypeArgBound(arg Type, bound Type, kind TypeBoundKind, nodes func(Type)(ast.Node), ctx TypeBoundsContext) *TypeError {
	var ctx_param_names = TypeParamsNames(ctx.Parameters)
	var ctx_mod = ctx.Module.Name
	switch kind {
	case SuperBound:
		var super = bound
		var checked = make(map[Type] bool)
		var ok = CheckBound(arg, super, checked, ctx)
		if !(ok) { return &TypeError {
			Point:    ErrorPointFrom(nodes(arg)),
			Concrete: E_BoundNotSatisfied {
				Kind:  kind,
				Bound: DescribeTypeWithParams(super, ctx_param_names, ctx_mod),
			},
		} }
	case SubBound:
		var sub = bound
		var checked = make(map[Type] bool)
		var ok = CheckBound(sub, arg, checked, ctx)
		if !(ok) { return &TypeError {
			Point:    ErrorPointFrom(nodes(arg)),
			Concrete: E_BoundNotSatisfied {
				Kind:  kind,
				Bound: DescribeTypeWithParams(sub, ctx_param_names, ctx_mod),
			},
		} }
	default:
		panic("impossible branch")
	}
	return nil
}

func CheckTypeArgsBounds(args ([] Type), params ([] TypeParam), defaults (map[uint] Type), bounds TypeBounds, node ast.Node, ctx ExprContext) *ExprError {
	var bound_ctx = ctx.GetTypeContext().TypeBoundsContext
	var get_node = func(_ Type) ast.Node { return node }
	var bad_type_arg = func(index uint, err *TypeError) *ExprError {
		return &ExprError {
			Point:    ErrorPointFrom(node),
			Concrete: E_BadTypeArg {
				Index:  index,
				Name:   params[index].Name,
				Detail: err,
			},
		}
	}
	for index, raw_super := range bounds.Super {
		var super = FillTypeArgsWithDefaults(raw_super, args, defaults)
		var arg = args[index]
		var err = CheckTypeArgBound(arg, super, SuperBound, get_node, bound_ctx)
		if err != nil { return bad_type_arg(index, err) }
	}
	for index, raw_sub := range bounds.Sub {
		var sub = FillTypeArgsWithDefaults(raw_sub, args, defaults)
		var arg = args[index]
		var err = CheckTypeArgBound(arg, sub, SubBound, get_node, bound_ctx)
		if err != nil { return bad_type_arg(index, err) }
	}
	return nil
}

func CheckBound(sub Type, super Type, checked (map[Type] bool), ctx TypeBoundsContext) bool {
	var sub_param, sub_is_param = sub.(*ParameterType)
	if sub_is_param && !(checked[sub]) {
		checked[sub] = true
		var sub_super, sub_has_super = ctx.Bounds.Super[sub_param.Index]
		if sub_has_super {
			return CheckBound(sub_super, super, checked, ctx)
		}
	}
	var super_param, super_is_param = super.(*ParameterType)
	if super_is_param && !(checked[super]) {
		checked[super] = true
		var super_sub, super_has_sub = ctx.Bounds.Sub[super_param.Index]
		if super_has_sub {
			return CheckBound(sub, super_sub, checked, ctx)
		}
	}
	if AreTypesEqualInSameCtx(sub, super) {
		return true
	}
	var unboxed, ok = Unbox(sub, ctx.Module.Name, ctx.Registry).(Unboxed)
	if ok {
		return CheckBound(unboxed.Type, super, checked, ctx)
	} else {
		return false
	}
}
