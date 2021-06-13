package typsys


type AssignContext struct {
	module     string
	subtyping  bool
	inferring  *InferringState
}
func (ctx *AssignContext) ApplyNewInferringState(s *InferringState) {
	ctx.inferring = s
}

func assignFromBeingInferred(to Type, p *Parameter, ctx AssignContext) (bool, *InferringState) {
	var ps, exists = ctx.inferring.currentInferredParameterState(p)
	if exists {
		if assignWithoutInferring(to, ps.currentInferred, ctx) {
			// 1. update condition
			if ctx.subtyping && ps.constraint == typeCanWiden {
				return true, ctx.inferring.withInferredTypeUpdate(p, ps, to)
			} else {
				return true, ctx.inferring
			}
		} else {
			return false, ctx.inferring
		}
	} else {
		var c inferredTypeConstraint
		if ctx.subtyping {
			// 2. constraint
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
				return true, ctx.inferring
			}
		} else {
			return false, ctx.inferring
		}
	} else {
		var c inferredTypeConstraint
		if ctx.subtyping {
			// 2. constraint
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

func Assign(to Type, from Type, ctx AssignContext) (bool, *InferringState) {
	if ctx.inferring != nil {
		var T, to_param = to.(ParameterType)
		var F, from_param = from.(ParameterType)
		var to_being_inferred = ctx.inferring.IsTargetType(T, to_param)
		var from_being_inferred = ctx.inferring.IsTargetType(F, from_param)
		if to_being_inferred && from_being_inferred {
			if T == F {
				return true, ctx.inferring
			} else {
				return false, nil
			}
		} else if !(to_being_inferred) && from_being_inferred {
			return assignFromBeingInferred(to, F.Parameter, ctx)
		} else if to_being_inferred && !(from_being_inferred) {
			return assignToBeingInferred(T.Parameter, from, ctx)
		}
	}
	switch T := to.(type) {
	// TODO
	}
	if ctx.subtyping {
		var _, to_top = to.(TopType)
		var _, from_bottom = from.(BottomType)
		if to_top || from_bottom {
			return true, ctx.inferring
		}
		// TODO
	}
}

