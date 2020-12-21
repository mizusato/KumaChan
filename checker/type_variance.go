package checker


type TypeVariance int
const (
	Invariant  TypeVariance  =  iota
	Covariant
	Contravariant
	Bivariant
)

type TypeVarianceContext struct {
	Registry    TypeRegistry
	Parameters  [] TypeParam
}
func (ctx TypeVarianceContext) Arity() uint {
	return uint(len(ctx.Parameters))
}
func (ctx TypeVarianceContext) DeduceVariance(params_v ([] TypeVariance), args_v ([][] TypeVariance)) ([] TypeVariance) {
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


func FilledVarianceVector(v TypeVariance, arity uint) ([] TypeVariance) {
	var draft = make([] TypeVariance, arity)
	for i, _ := range draft {
		draft[i] = v
	}
	return draft
}

func ParamsVarianceVector(params ([] TypeParam)) ([] TypeVariance) {
	var draft = make([] TypeVariance, len(params))
	for i, _ := range draft {
		draft[i] = params[i].Variance
	}
	return draft
}

func InverseVariance(v TypeVariance) TypeVariance {
	if v == Contravariant {
		return Covariant
	} else if v == Covariant {
		return Contravariant
	} else {
		return v
	}
}

func ApplyVariance(param TypeVariance, arg TypeVariance) TypeVariance {
	if arg == Covariant || arg == Contravariant {
		if param == Bivariant {
			return arg
		} else if arg == param {
			return Covariant
		} else if arg == InverseVariance(param) {
			return Contravariant
		} else {
			return Invariant
		}
	} else if arg == Bivariant {
		if param == Bivariant {
			return Bivariant
		} else if param == Covariant || param == Contravariant {
			return param
		} else {
			return Invariant
		}
	} else {
		return Invariant
	}
}

func CombineVariance(a TypeVariance, b TypeVariance) TypeVariance {
	if a == Invariant || b == Invariant {
		return Invariant
	} else if a == Bivariant {
		return b
	} else if b == Bivariant {
		return a
	} else if a == Covariant && b == Covariant {
		return Covariant
	} else if a == Contravariant && b == Contravariant {
		return Contravariant
	} else {
		return Invariant
	}
}

func MatchVariance(declared ([] TypeParam), deduced ([] TypeVariance)) (bool, ([] string)) {
	var bad_params = make([] string, 0)
	for i, _ := range declared {
		var v = deduced[i]
		var name = declared[i].Name
		var declared_v = declared[i].Variance
		var bad = false
		if declared_v == Covariant {
			if !(v == Covariant || v == Bivariant) {
				bad = true
			}
		} else if declared_v == Contravariant {
			if !(v == Contravariant || v == Bivariant) {
				bad = true
			}
		} else if declared_v == Bivariant {
			if v != Bivariant {
				bad = true
			}
		}
		if bad {
			bad_params = append(bad_params, name)
		}
	}
	if len(bad_params) > 0 {
		return false, bad_params
	} else {
		return true, nil
	}
}

func GetVariance(t Type, ctx TypeVarianceContext) ([] TypeVariance) {
	switch T := t.(type) {
	case *NeverType:
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

