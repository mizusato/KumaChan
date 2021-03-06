package ast


type MaybePattern interface { Maybe(Pattern,MaybePattern) }
func (impl VariousPattern) Maybe(Pattern,MaybePattern) {}
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

type MaybePatternTuple interface { Maybe(PatternTuple,MaybePatternTuple) }
func (impl PatternTuple) Pattern() {}
func (impl PatternTuple) Maybe(PatternTuple,MaybePatternTuple) {}
type PatternTuple struct {
	Node                   `part:"pattern_tuple"`
	Names  [] Identifier   `list_more:"namelist" item:"name"`
}

func (impl PatternRecord) Pattern() {}
type PatternRecord struct {
	Node                     `part:"pattern_record"`
	FieldMaps  [] FieldMap   `list_more:"field_map_list" item:"field_map"`
}

type FieldMap struct {
	Node                    `part:"field_map"`
	ValueName  Identifier   `part:"name"`
	FieldName  Identifier   `part:"field_map_to.name" fallback:"name"`
}
