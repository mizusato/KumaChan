package checker

import (
	"kumachan/parser/ast"
	. "kumachan/error"
)


type TypeBounds struct {
	Sub    map[uint] Type
	Super  map[uint] Type
}

type TypeBoundsContext struct {
	TypeValidationContext
	Bounds  TypeBounds
}

func CheckTypeValBounds(val TypeVal, info TypeNodeInfo, ctx TypeBoundsContext) *TypeError {
	switch V := val.(type) {
	case *Union:
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
	// TODO: detect circular reference of parameter types
	const kind_super = "<"
	const kind_sub = ">"
	var ctx_param_names = TypeParamsNames(ctx.Parameters)
	switch T := t.(type) {
	case *WildcardRhsType:
		return nil
	case *ParameterType:
		return nil
	case *NamedType:
		// TODO: debug
		var g = ctx.Registry[T.Name]
		for i, super := range g.Bounds.Super {
			var arg = T.Args[i]
			var ok = CheckBound(arg, super, ctx)
			if !(ok) { return &TypeError {
				Point:    ErrorPointFrom(nodes[arg]),
				Concrete: E_BoundNotSatisfied {
					Kind:  kind_super,
					Bound: DescribeTypeWithParams(super, ctx_param_names),
				},
			} }
		}
		for i, sub := range g.Bounds.Sub {
			var arg = T.Args[i]
			var ok = CheckBound(sub, arg, ctx)
			if !(ok) { return &TypeError {
				Point:    ErrorPointFrom(nodes[arg]),
				Concrete: E_BoundNotSatisfied {
					Kind:  kind_sub,
					Bound: DescribeTypeWithParams(sub, ctx_param_names),
				},
			} }
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

func CheckBound(sub Type, super Type, ctx TypeBoundsContext) bool {
	var sub_param, sub_is_param = sub.(*ParameterType)
	if sub_is_param {
		var sub_super, sub_has_super = ctx.Bounds.Super[sub_param.Index]
		if sub_has_super {
			return CheckBound(sub_super, super, ctx)
		}
	}
	var super_param, super_is_param = super.(*ParameterType)
	if super_is_param {
		var super_sub, super_has_sub = ctx.Bounds.Sub[super_param.Index]
		if super_has_sub {
			return CheckBound(sub, super_sub, ctx)
		}
	}
	if AreTypesEqualInSameCtx(sub, super) {
		return true
	}
	var unboxed, ok = Unbox(sub, ctx.Module.Name, ctx.Registry).(Unboxed)
	if ok {
		return CheckBound(unboxed.Type, super, ctx)
	} else {
		return false
	}
}
