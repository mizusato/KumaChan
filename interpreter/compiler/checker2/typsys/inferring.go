package typsys


type InferringState struct {
	Mapping  map[uintptr] InferredTypeState
}

type InferredTypeState struct {
	CurrentValue  Type
	Constraint    InferringConstraint
}
type InferringConstraint int
const (
	AcceptExact InferringConstraint = iota
	AcceptExactOrBigger
	AcceptExactOrSmaller
)


