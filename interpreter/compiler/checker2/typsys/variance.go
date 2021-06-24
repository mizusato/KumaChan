package typsys



func InverseVariance(v Variance) Variance {
	if v == Contravariant {
		return Covariant
	} else if v == Covariant {
		return Contravariant
	} else {
		return v
	}
}
func ApplyVariance(param Variance, arg Variance) Variance {
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
		return Bivariant
	} else {
		return Invariant
	}
}
func CombineVariance(a Variance, b Variance) Variance {
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

func MatchVariance(declared ([] Parameter), deduced ([] Variance)) (bool, ([] string)) {
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
func DeduceVariance(N uint, p ([] Variance), a ([][] Variance)) ([] Variance) {
	var result = make([] Variance, N)
	var M = uint(len(p))
	if uint(len(a)) != M { panic("invalid argument") }
	for i := uint(0); i < N; i += 1 {
		var v = Bivariant
		for j := uint(0); j < M; j += 1 {
			if uint(len(a[j])) != N { panic("invalid argument") }
			v = CombineVariance(v, ApplyVariance(p[j], a[j][i]))
		}
		result[i] = v
	}
	return result
}

func FilledVarianceVector(v Variance, arity uint) ([] Variance) {
	var vec = make([] Variance, arity)
	for i := uint(0); i < arity; i += 1 {
		vec[i] = v
	}
	return vec
}
func ParametersVarianceVector(params ([] Parameter)) ([] Variance) {
	var draft = make([] Variance, len(params))
	for i, _ := range draft {
		draft[i] = params[i].Variance
	}
	return draft
}
func GetVariance(t Type, p ([] Parameter)) ([] Variance) {
	var arity = uint(len(p))
	switch T := t.(type) {
	case *UnknownType:
		return FilledVarianceVector(Bivariant, arity)
	case UnitType:
		return FilledVarianceVector(Bivariant, arity)
	case TopType:
		return FilledVarianceVector(Bivariant, arity)
	case BottomType:
		return FilledVarianceVector(Bivariant, arity)
	case ParameterType:
		var vec = make([] Variance, arity)
		for i := uint(0); i < arity; i += 1 {
			if &(p[i]) == T.Parameter {
				vec[i] = Covariant
			} else {
				vec[i] = Bivariant
			}
		}
		return vec
	case *NestedType:
		switch T := T.Content.(type) {
		case Ref:
			var v_params = ParametersVarianceVector(T.Def.Parameters)
			var v_args = make([][] Variance, len(v_params))
			for i := 0; i < len(v_args); i += 1 {
				v_args[i] = GetVariance(T.Args[i], p)
			}
			return DeduceVariance(arity, v_params, v_args)
		case Tuple:
			var n = uint(len(T.Elements))
			var v_tuple = FilledVarianceVector(Covariant, n)
			var v_elements = make([][] Variance, n)
			for i, el := range T.Elements {
				v_elements[i] = GetVariance(el, p)
			}
			return DeduceVariance(arity, v_tuple, v_elements)
		case Record:
			var n = uint(len(T.Fields))
			var v_record = FilledVarianceVector(Covariant, n)
			var v_fields = make([][] Variance, n)
			for i, field := range T.Fields {
				v_fields[i] = GetVariance(field.Type, p)
			}
			return DeduceVariance(arity, v_record, v_fields)
		case Lambda:
			var v_input = GetVariance(T.Input, p)
			var v_output = GetVariance(T.Output, p)
			var v_lambda = [] Variance { Contravariant, Covariant }
			var v_io = [][] Variance { v_input, v_output }
			return DeduceVariance(arity, v_lambda, v_io)
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}


