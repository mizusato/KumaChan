package typsys


type AssignContext struct {
	Module         string
	UseSubtyping   bool
	Inferring      *InferringState
}

func assertCertainType(t Type, s *InferringState) Type {
	if s != nil {
		return TypeOpMap(t, func(t Type) (Type, bool) {
			var T, is_param = t.(ParameterType)
			if !(is_param) { return nil, false }
			var _, is_target = s.targets[T.Parameter]
			if !(is_target) { return nil, false }
			panic("invalid assignment")
		})
	} else {
		return t
	}
}

func Assign(to Type, from Type, ctx AssignContext) (bool, *InferringState) {
	if ctx.Inferring != nil {
		var T, to_param = to.(ParameterType)
		var F, from_param = from.(ParameterType)
		var t_being_inferred = false
		if to_param {
			_, t_being_inferred = ctx.Inferring.targets[T.Parameter]
		}
		var from_being_inferred = false
		if from_param {
			_, from_being_inferred = ctx.Inferring.targets[F.Parameter]
		}
		if t_being_inferred && from_being_inferred {
			panic("invalid assignment")
		} else if !(t_being_inferred) && from_being_inferred {
			to = assertCertainType(to, ctx.Inferring)
			var ps, exists = ctx.Inferring.mapping[F.Parameter]
			if exists {
				var from_current = ps.currentInferred
				var ok, s = Assign(to, from_current, ctx)
				if s != ctx.Inferring { panic("something went wrong") }
				if ok {
					if ctx.UseSubtyping && ps.status == typeCanWiden {
						s = s.WithMappingCloned()
						ps.currentInferred = to
						s.mapping[F.Parameter] = ps
						return true, s
					} else {
						return true, s
					}
				} else {
					return false, ctx.Inferring
				}
			} else {
				var s = ctx.Inferring.WithMappingCloned()
				var init_status activeInferredTypeStatus
				if ctx.UseSubtyping {
					init_status = typeCanNarrow
				} else {
					init_status = typeFixed
				}
				s.mapping[F.Parameter] = beingInferredParameterState {
					status:          init_status,
					currentInferred: to,
				}
				return true, s
			}
		} else if t_being_inferred && !(from_being_inferred) {
			from = assertCertainType(from, ctx.Inferring)
			var ps, exists = ctx.Inferring.mapping[T.Parameter]
			if exists {
				var to_current = ps.currentInferred
				var ok, s = Assign(to_current, from, ctx)
				if s != ctx.Inferring { panic("something went wrong") }
				if ok {
					if ctx.UseSubtyping && ps.status == typeCanNarrow {
						s = s.WithMappingCloned()
						ps.currentInferred = from
						s.mapping[T.Parameter] = ps
						return true, s
					} else {
						return true, s
					}
				} else {
					return false, ctx.Inferring
				}
			} else {
				var s = ctx.Inferring.WithMappingCloned()
				var init_status activeInferredTypeStatus
				if ctx.UseSubtyping {
					init_status = typeCanWiden
				} else {
					init_status = typeFixed
				}
				s.mapping[T.Parameter] = beingInferredParameterState {
					status:          init_status,
					currentInferred: from,
				}
				return true, s
			}
		}
	}
	var _, to_top = to.(TopType)
	var _, from_bottom = from.(BottomType)
	if to_top || from_bottom {
		return true, ctx.Inferring
	}
	panic("not implemented") // TODO
}

