package ast


type MaybePattern interface { MaybePattern() }
func (impl VariousPattern) MaybePattern() {}
type VariousPattern struct {
	Node               `part:"pattern"`
	Pattern  Pattern   `use:"first"`
}
type Pattern interface { Pattern() }

func (impl PatternTrivial) Pattern() {}
type PatternTrivial struct {
	Node                `part:"pattern_trivial"`
	Name   Identifier   `part:"name"`
}

type MaybePatternTuple interface { MaybePatternTuple() }
func (impl PatternTuple) Pattern() {}
func (impl PatternTuple) MaybePatternTuple() {}
type PatternTuple struct {
	Node                   `part:"pattern_tuple"`
	Names  [] Identifier   `list_more:"namelist" item:"name"`
}

func (impl PatternBundle) Pattern() {}
type PatternBundle struct {
	Node                   `part:"pattern_bundle"`
	Names  [] Identifier   `list_more:"namelist" item:"name"`
}
