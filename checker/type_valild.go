package checker

import (
	"kumachan/parser/ast"
	"kumachan/runtime/common"
	. "kumachan/error"
)


type TypeContext struct {
	TypeConstructContext
	Registry  TypeRegistry
}
type TypeNodeInfo struct {
	ValNodeMap   map[TypeVal] ast.Node
	TypeNodeMap  map[Type] ast.Node
}
func (ctx TypeContext) Arity() uint {
	return uint(len(ctx.Parameters))
}
func (ctx TypeContext) DeduceVariance(params_v ([] TypeVariance), args_v ([][] TypeVariance)) ([] TypeVariance) {
	var ctx_arity = ctx.Arity()
	var n = uint(len(params_v))
	var result = make([] TypeVariance, ctx_arity)
	for i := uint(0); i < ctx_arity; i += 1 {
		var v = Bivariant
		for j := uint(0); j < n; j += 1 {
			v = CombineVariance(v, ApplyVariance(params_v[j], args_v[j][i]))
		}
		result[i] = TypeVariance(v)
	}
	return result
}

func TypeFrom(ast_type ast.VariousType, ctx TypeContext) (Type, *TypeError) {
	var info = make(map[Type] ast.Node)
	var t, err = RawTypeFrom(ast_type, info, ctx.TypeConstructContext)
	if err != nil { return nil, err }
	err = ValidateType(t, info, ctx)
	if err != nil { return nil, err }
	return t, nil
}

func TypeFromRepr(ast_repr ast.VariousRepr, ctx TypeContext) (Type, *TypeError) {
	var info = make(map[Type] ast.Node)
	var t, err = RawTypeFromRepr(ast_repr, info, ctx.TypeConstructContext)
	if err != nil { return nil, err }
	err = ValidateType(t, info, ctx)
	if err != nil { return nil, err }
	return t, nil
}

func ValidateTypeVal(val TypeVal, info TypeNodeInfo, ctx TypeContext) *TypeError {
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
		var inner_v = GetVariance(V.InnerType, ctx)
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

func ValidateType(t Type, nodes (map[Type] ast.Node), ctx TypeContext) *TypeError {
	var t_point = ErrorPointFrom(nodes[t])
	switch T := t.(type) {
	case *WildcardRhsType:
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

func GetVariance(t Type, ctx TypeContext) ([] TypeVariance) {
	switch T := t.(type) {
	case *WildcardRhsType:
		return FilledVarianceVector(Bivariant, ctx.Arity())
	case *ParameterType:
		var v_draft = FilledVarianceVector(Bivariant, ctx.Arity())
		v_draft[T.Index] = Covariant
		var v = v_draft
		return v
	case *NamedType:
		var g = ctx.Registry[T.Name]
		var arity = uint(len(g.Params))
		if uint(len(T.Args)) != arity { panic("something went wrong") }
		var params_v = ParamsVarianceVector(g.Params)
		var args_v = make([][] TypeVariance, arity)
		for i, arg := range T.Args {
			args_v[i] = GetVariance(arg, ctx)
		}
		return ctx.DeduceVariance(params_v, args_v)
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Unit:
			return FilledVarianceVector(Bivariant, ctx.Arity())
		case Tuple:
			var n = uint(len(R.Elements))
			var tuple_v = FilledVarianceVector(Covariant, n)
			var elements_v = make([][] TypeVariance, n)
			for i, el := range R.Elements {
				elements_v[i] = GetVariance(el, ctx)
			}
			return ctx.DeduceVariance(tuple_v, elements_v)
		case Bundle:
			var n = uint(len(R.Fields))
			var bundle_v = FilledVarianceVector(Covariant, n)
			var fields_v = make([][] TypeVariance, n)
			for _, field := range R.Fields {
				fields_v[field.Index] = GetVariance(field.Type, ctx)
			}
			return ctx.DeduceVariance(bundle_v, fields_v)
		case Func:
			var input_v = GetVariance(R.Input, ctx)
			var output_v = GetVariance(R.Output, ctx)
			var func_v = [] TypeVariance { Contravariant, Covariant }
			var io_v = [][] TypeVariance { input_v, output_v }
			return ctx.DeduceVariance(func_v, io_v)
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

