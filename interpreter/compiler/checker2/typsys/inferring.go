package typsys


type InferringSuccessResult struct {
	mapping  map[*Parameter] Type
}

type InferringFailureResult struct {
	missing  [] *Parameter
}

type InferringState struct {
	targets  map[*Parameter] struct{}
	mapping  map[*Parameter] beingInferredParameterState
}
func (s *InferringState) WithMappingCloned() *InferringState {
	var cloned_mapping = make(map[*Parameter] beingInferredParameterState)
	for k, v := range s.mapping {
		cloned_mapping[k] = v
	}
	return &InferringState {
		targets: s.targets,
		mapping: cloned_mapping,
	}
}

type beingInferredParameterState struct {
	status           activeInferredTypeStatus
	currentInferred  Type
}
// TODO: maybe should not be called "status" (not changing)
type activeInferredTypeStatus int
const (
	typeFixed activeInferredTypeStatus = iota
	typeCanWiden
	typeCanNarrow
)
func activeInferredTypeStatusFromBound(kind BoundKind) activeInferredTypeStatus {
	switch kind {
	case SupBound:
		return typeCanNarrow
	case InfBound:
		return typeCanWiden
	default:
		panic("invalid argument")
	}
}
func (c activeInferredTypeStatus) OperatorString() string {
	switch c {
	case typeFixed:     return "="
	case typeCanWiden:  return ">"
	case typeCanNarrow: return "<"
	default:            panic("impossible branch")
	}
}

func StartInferring(targets ([] *Parameter)) *InferringState {
	var target_set = make(map[*Parameter] struct{})
	var mapping = make(map[*Parameter] beingInferredParameterState)
	for _, p := range targets {
		target_set[p] = struct{}{}
		if p.Bound.Kind != NullBound {
			var bound_kind = p.Bound.Kind
			var bound_t = p.Bound.Value
			mapping[p] = beingInferredParameterState {
				status:          activeInferredTypeStatusFromBound(bound_kind),
				currentInferred: bound_t,
			}
		}
	}
	return &InferringState{
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
		return &InferringSuccessResult{mapping }, nil
	} else {
		return nil, &InferringFailureResult{missing }
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


