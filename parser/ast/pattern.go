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
	Node                     `part:"pattern_bundle"`
	FieldMaps  [] FieldMap   `list_more:"field_map_list" item:"field_map"`
}

type FieldMap struct {
	Node                    `part:"field_map"`
	ValueName  Identifier   `part:"name"`
	FieldName  Identifier   `part:"field_map_to.name" fallback:"name"`
}
