package typsys


type AssignContext struct {
	module     string
	subtyping  bool
	inferring  *InferringState
}
func MakeAssignContext(mod string, s *InferringState) AssignContext {
	return AssignContext {
		module:    mod,
		subtyping: true,
		inferring: s,
	}
}
func MakeAssignContextWithoutSubtyping(s *InferringState) AssignContext {
	return AssignContext {
		module:    "",
		subtyping: false,
		inferring: s,
	}
}
func ApplyNewInferringState(ctx *AssignContext, s *InferringState) {
	if s != nil {
		ctx.inferring = s
	}
}

func Assign(to Type, from Type, ctx AssignContext) (bool, *InferringState) {
	if to == nil || from == nil {
		panic("something went wrong")
	}
	// 1. Deal with parameter inferring
	if ctx.inferring != nil {
		var T, to_param = to.(ParameterType)
		var F, from_param = from.(ParameterType)
		var to_being_inferred = ctx.inferring.IsTargetType(T, to_param)
		var from_being_inferred = ctx.inferring.IsTargetType(F, from_param)
		if to_being_inferred && from_being_inferred {
			if T == F {
				return true, nil
			} else {
				return false, nil
			}
		} else if !(to_being_inferred) && from_being_inferred {
			return assignFromBeingInferred(to, F.Parameter, ctx)
		} else if to_being_inferred && !(from_being_inferred) {
			return assignToBeingInferred(T.Parameter, from, ctx)
		} else {
			goto direct
		}
	}
	// 2. Check structural equality and recurse into nested cases
	direct:
	{
		var ok, s = assignDirect(to, from, ctx)
		if ok {
			return true, s
		} else {
			goto subtyping
		}
	}
	// 3. Apply subtyping rules
	subtyping:
	if ctx.subtyping {
		// 3.1. TopType and BottomType
		var _, to_top = to.(TopType)
		var _, from_bottom = from.(BottomType)
		if to_top || from_bottom {
			return true, nil
		} else {
			goto bound
		}
		// 3.2. ParameterType (Bound)
		bound:
		{
			var T, to_param = to.(ParameterType)
			var F, from_param = from.(ParameterType)
			if to_param && from_param {
				return (T == F), nil
			} else if !(to_param) && from_param {
				if F.Parameter.Bound.Kind == SupBound {
					var from_sup = F.Parameter.Bound.Value
					return Assign(to, from_sup, ctx)
				} else {
					return false, nil
				}
			} else if to_param && !(from_param) {
				if T.Parameter.Bound.Kind == InfBound {
					var to_inf = T.Parameter.Bound.Value
					return Assign(to_inf, from, ctx)
				} else {
					return false, nil
				}
			} else {
				goto unbox
			}
		}
		// 3.3. Ref of Box (Unbox)
		unbox:
		var from_sup, ok = Unbox(from, ctx.module)
		if ok {
			return Assign(to, from_sup, ctx)
		} else {
			goto final
		}
	}
	final:
	return false, nil
}

func assignFromBeingInferred(to Type, p *Parameter, ctx AssignContext) (bool, *InferringState) {
	var ps, exists = ctx.inferring.currentInferredParameterState(p)
	if exists {
		if assignWithoutInferring(to, ps.currentInferred, ctx) {
			// 1. update condition
			if ctx.subtyping && ps.constraint == typeCanWiden {
				return true, ctx.inferring.withInferredTypeUpdate(p, ps, to)
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	} else {
		var c inferredTypeConstraint
		if ctx.subtyping {
			// 2. constraint of new state
			c = typeCanNarrow
		} else {
			c = typeFixed
		}
		return true, ctx.inferring.withNewParameterState(p, c, to)
	}
}

func assignToBeingInferred(p *Parameter, from Type, ctx AssignContext) (bool, *InferringState) {
	var ps, exists = ctx.inferring.currentInferredParameterState(p)
	if exists {
		if assignWithoutInferring(ps.currentInferred, from, ctx) {
			// 1. update condition
			if ctx.subtyping && ps.constraint == typeCanNarrow {
				return true, ctx.inferring.withInferredTypeUpdate(p, ps, from)
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	} else {
		var c inferredTypeConstraint
		if ctx.subtyping {
			// 2. constraint of new state
			c = typeCanWiden
		} else {
			c = typeFixed
		}
		return true, ctx.inferring.withNewParameterState(p, c, from)
	}
}

func assignWithoutInferring(to Type, from Type, ctx AssignContext) bool {
	var ok, _ = Assign(to, from, AssignContext {
		module:    ctx.module,
		subtyping: ctx.subtyping,
		inferring: nil,
	})
	return ok
}

func assignDirect(to Type, from Type, ctx AssignContext) (bool, *InferringState) {
	// 1. UnknownType
	var _, to_unknown = to.(*UnknownType)
	var _, from_unknown = from.(*UnknownType)
	if to_unknown || from_unknown {
		return false, nil
	}
	// 2. UnitType
	var _, to_unit = to.(UnitType)
	var _, from_unit = from.(UnitType)
	if to_unit || from_unit {
		return (to_unit && from_unit), nil
	}
	// 3. TopType
	var _, to_top = to.(TopType)
	var _, from_top = from.(TopType)
	if to_top || from_top {
		return (to_top && from_top), nil
	}
	// 4. BottomType
	var _, to_bottom = to.(BottomType)
	var _, from_bottom = from.(BottomType)
	if to_bottom || from_bottom {
		return (to_bottom && from_bottom), nil
	}
	// 5. ParameterType
	{
		var T, to_param = to.(ParameterType)
		var F, from_param = from.(ParameterType)
		if to_param || from_param {
			if to_param && from_param {
				return (T == F), nil
			} else {
				return false, nil
			}
		}
	}
	// 6. NestedType
	{
		var T, to_nested = to.(*NestedType)
		var F, from_nested = from.(*NestedType)
		if to_nested || from_nested {
			if !(to_nested && from_nested) {
				return false, nil
			}
			// 6.1. Ref
			{
				var T, to_ref = T.Content.(Ref)
				var F, from_ref = F.Content.(Ref)
				if to_ref || from_ref {
					if !(to_ref && from_ref) {
						return false, nil
					}
					if T.Def == F.Def {
						var d = T.Def
						if len(T.Args) != len(F.Args) {
							panic("something went wrong")
						}
						var v = ParametersVarianceVector(d.Parameters)
						if len(T.Args) != len(v) {
							panic("something went wrong")
						}
						return assignVector(T.Args, F.Args, v, ctx)
					} else {
						return false, nil
					}
				}
			}
			// 6.2. Tuple
			{
				var T, to_tuple = T.Content.(Tuple)
				var F, from_tuple = F.Content.(Tuple)
				if to_tuple || from_tuple {
					if !(to_tuple && from_tuple) {
						return false, nil
					}
					return assignVector(T.Elements, F.Elements, nil, ctx)
				}
			}
			// 6.3. Record
			{
				var T, to_record = T.Content.(Record)
				var F, from_record = F.Content.(Record)
				if to_record || from_record {
					if !(to_record && from_record) {
						return false, nil
					}
					return assignFields(T.Fields, F.Fields, ctx)
				}
			}
			// 6.4. Lambda
			{
				var T, to_lambda = T.Content.(Lambda)
				var F, from_lambda = F.Content.(Lambda)
				if to_lambda || from_lambda {
					if !(to_lambda && from_lambda) {
						return false, nil
					}
					var to_io = [] Type { T.Input, T.Output }
					var from_io = [] Type { F.Input, F.Output }
					var v = [] Variance { Contravariant, Covariant }
					return assignVector(to_io, from_io, v, ctx)
				}
			}
		}
	}
	return false, nil
}

func assignVector(to ([] Type), from ([] Type), v ([] Variance), ctx AssignContext) (bool, *InferringState) {
	if len(to) != len(from) {
		return false, nil
	}
	var L = len(to)
	for i := 0; i < L; i += 1 {
		var this_ctx = ctx
		var this_to Type
		var this_from Type
		var this_v Variance
		if i < len(v) {
			this_v = v[i]
		} else {
			this_v = Covariant
		}
		switch this_v {
		case Invariant:
			this_ctx.subtyping = false
			this_to = to[i]
			this_from = from[i]
		case Covariant:
			this_ctx.subtyping = ctx.subtyping
			this_to = to[i]
			this_from = from[i]
		case Contravariant:
			this_ctx.subtyping = ctx.subtyping
			this_to = from[i]
			this_from = to[i]
		default:
			panic("something went wrong")
		}
		var ok, s = Assign(this_to, this_from, this_ctx)
		if !(ok) {
			return false, nil
		}
		ApplyNewInferringState(&ctx, s)
	}
	return true, ctx.inferring
}

func assignFields(to ([] Field), from ([] Field), ctx AssignContext) (bool, *InferringState) {
	if len(to) != len(from) {
		return false, nil
	}
	var L = len(to)
	var to_types = make([] Type, L)
	var from_types = make([] Type, L)
	for i := 0; i < L; i += 1 {
		if to[i].Name != from[i].Name {
			return false, nil
		}
		to_types[i] = to[i].Type
		from_types[i] = from[i].Type
	}
	return assignVector(to_types, from_types, nil, ctx)
}


