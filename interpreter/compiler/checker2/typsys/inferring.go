package typsys


type InferringSuccessResult struct {
	mapping  map[*Parameter] Type
}

type InferringFailureResult struct {
	missing  [] *Parameter
}

type InferringState struct {
	targets  map[*Parameter] struct{}
	mapping  map[*Parameter] parameterInferringState
}
func (s *InferringState) assertInferringTarget(p *Parameter) {
	if !(s.IsTarget(p)) {
		panic("something went wrong")
	}
}
func (s *InferringState) assertCertainType(t Type) {
	if s != nil {
		TypeOpMap(t, func(t Type) (Type, bool) {
			var T, is_param = t.(ParameterType)
			if s.IsTargetType(T, is_param) {
				panic("something went wrong")
			}
			return nil, false
		})
	}
}
func (s *InferringState) withMappingCloned() *InferringState {
	var cloned_mapping = make(map[*Parameter]parameterInferringState)
	for k, v := range s.mapping {
		cloned_mapping[k] = v
	}
	return &InferringState {
		targets: s.targets,
		mapping: cloned_mapping,
	}
}
func (s *InferringState) currentInferredParameterState(p *Parameter) (parameterInferringState, bool) {
	s.assertInferringTarget(p)
	var ps, exists = s.mapping[p]
	return ps, exists
}
func (s *InferringState) withNewParameterState(p *Parameter, c inferredTypeConstraint, t Type) *InferringState {
	s.assertCertainType(t)
	{
		var s = s.withMappingCloned()
		s.mapping[p] = parameterInferringState{
			constraint:      c,
			currentInferred: t,
		}
		return s
	}
}
func (s *InferringState) withInferredTypeUpdate(p *Parameter, ps parameterInferringState, t Type) *InferringState {
	s.assertCertainType(t)
	{
		var s = s.withMappingCloned()
		s.mapping[p] = parameterInferringState{
			constraint:      ps.constraint,
			currentInferred: t,
		}
		return s
	}
}
func (s *InferringState) IsTarget(p *Parameter) bool {
	var _, is_being_inferred = s.targets[p]
	return is_being_inferred
}
func (s *InferringState) IsTargetType(T ParameterType, ok bool) bool {
	if ok {
		return s.IsTarget(T.Parameter)
	} else {
		return false
	}
}

type parameterInferringState struct {
	constraint       inferredTypeConstraint
	currentInferred  Type
}
type inferredTypeConstraint int
const (
	typeFixed inferredTypeConstraint = iota
	typeCanWiden
	typeCanNarrow
)
func activeInferredTypeStatusFromBound(kind BoundKind) inferredTypeConstraint {
	switch kind {
	case SupBound:
		return typeCanNarrow
	case InfBound:
		return typeCanWiden
	default:
		panic("invalid argument")
	}
}
func (c inferredTypeConstraint) OperatorString() string {
	switch c {
	case typeFixed:     return "="
	case typeCanWiden:  return ">"
	case typeCanNarrow: return "<"
	default:            panic("impossible branch")
	}
}

func StartInferring(targets ([] *Parameter)) *InferringState {
	var target_set = make(map[*Parameter] struct{})
	var mapping = make(map[*Parameter]parameterInferringState)
	for _, p := range targets {
		target_set[p] = struct{}{}
		if p.Bound.Kind != NullBound {
			var bound_kind = p.Bound.Kind
			var bound_t = p.Bound.Value
			mapping[p] = parameterInferringState {
				constraint:      activeInferredTypeStatusFromBound(bound_kind),
				currentInferred: bound_t,
			}
		}
	}
	return &InferringState {
		targets: target_set,
		mapping: mapping,
	}
}

func GetInferringResult(s *InferringState) (*InferringSuccessResult, *InferringFailureResult) {
	var missing = make([] *Parameter, 0)
	var success = true
	for p := range s.targets {
		var _, exists = s.mapping[p]
		if !(exists) {
			success = false
			missing = append(missing, p)
		}
	}
	if success {
		var mapping = make(map[*Parameter] Type)
		for p, ps := range s.mapping {
			mapping[p] = ps.currentInferred
		}
		return &InferringSuccessResult { mapping }, nil
	} else {
		return nil, &InferringFailureResult { missing }
	}
}

func ApplyInferringResult(t Type, result *InferringSuccessResult) Type {
	if result == nil {
		panic("invalid argument")
	}
	return TypeOpMap(t, func(t Type) (Type, bool) {
		var param_t, is_param = t.(ParameterType)
		if !(is_param) { return nil, false }
		var result_t, is_target = result.mapping[param_t.Parameter]
		if !(is_target) { return nil, false }
		return result_t, true
	})
}


