package checker

import (
	"kumachan/compiler/loader/parser/ast"
	"kumachan/runtime/common"
	. "kumachan/util/error"
)


type TypeValidationContext struct {
	TypeConstructContext
	Registry  TypeRegistry
}
func (ctx TypeValidationContext) GetVarianceContext() TypeVarianceContext {
	return TypeVarianceContext {
		Registry:   ctx.Registry,
		Parameters: ctx.Parameters,
	}
}

func ValidateTypeVal(val TypeVal, info TypeNodeInfo, ctx TypeValidationContext) *TypeError {
	var val_point = ErrorPointFrom(info.ValNodeMap[val])
	switch V := val.(type) {
	case *Union:
		var case_amount = uint(len(V.CaseTypes))
		var max = uint(common.SumMaxBranches)
		if case_amount > max { return &TypeError {
			Point:    val_point,
			Concrete: E_TooManyUnionItems {
				Defined: case_amount,
				Limit:   max,
			},
		} }
		var union_v = ParamsVarianceVector(ctx.Parameters)
		for _, case_t := range V.CaseTypes {
			var case_g = ctx.Registry[case_t.Name]
			var case_v = ParamsVarianceVector(case_g.Params)
			var info = case_g.CaseInfo
			if !(info.IsCaseType) { panic("something went wrong") }
			for i, j := range info.CaseParams {
				if case_v[i] != union_v[j] {
					return &TypeError {
						Point:    ErrorPointFrom(case_g.Node),
						Concrete: E_CaseBadVariance {
							CaseName:  case_t.Name.String(),
							UnionName: info.UnionName.String(),
						},
					}
				}
			}
		}
		return nil
	case *Boxed:
		var err = ValidateType(V.InnerType, info.TypeNodeMap, ctx)
		if err != nil { return err }
		var inner_v = GetVariance(V.InnerType, ctx.GetVarianceContext())
		var v_ok, bad_params = MatchVariance(ctx.Parameters, inner_v)
		if !(v_ok) { return &TypeError {
			Point:    val_point,
			Concrete: E_BoxedBadVariance { bad_params },
		} }
		return nil
	case *Native:
		return nil
	default:
		panic("impossible branch")
	}
}

func ValidateType(t Type, nodes (map[Type] ast.Node), ctx TypeValidationContext) *TypeError {
	var t_point = ErrorPointFrom(nodes[t])
	switch T := t.(type) {
	case *NeverType:
		return nil
	case *AnyType:
		return nil
	case *ParameterType:
		return nil
	case *NamedType:
		var g, exists = ctx.Registry[T.Name]
		if !exists { return &TypeError {
			Point:    t_point,
			Concrete: E_TypeNotFound {
				Name: T.Name,
			},
		} }
		var arity = uint(len(g.Params))
		var given_arity = uint(len(T.Args))
		if arity != given_arity { return &TypeError {
			Point:    t_point,
			Concrete: E_WrongParameterQuantity {
				TypeName: T.Name,
				Required: arity,
				Given:    given_arity,
			},
		} }
		for _, arg := range T.Args {
			var err = ValidateType(arg, nodes, ctx)
			if err != nil { return err }
		}
		return nil
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Unit:
			return nil
		case Tuple:
			var count = uint(len(R.Elements))
			var max = uint(common.ProductMaxSize)
			if count > max { return &TypeError {
				Point:    t_point,
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			} }
			for _, el := range R.Elements {
				err := ValidateType(el, nodes, ctx)
				if err != nil { return err }
			}
			return nil
		case Bundle:
			var count = uint(len(R.Fields))
			var max = uint(common.ProductMaxSize)
			if count > max { return &TypeError {
				Point:    t_point,
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			} }
			for _, field := range R.Fields {
				err := ValidateType(field.Type, nodes, ctx)
				if err != nil { return err }
			}
			return nil
		case Func:
			err := ValidateType(R.Input, nodes, ctx)
			if err != nil { return err }
			err = ValidateType(R.Output, nodes, ctx)
			if err != nil { return err }
			return nil
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}
