package kmd


type SchemaTable map[TypeId]Schema
type Schema interface{ KmdSchema() }

func (RecordSchema) KmdSchema() {}

type RecordSchema struct {
	Fields  map[string] RecordField
}
type RecordField struct {
	Type   *Type
	Index  uint
}

func (TupleSchema) KmdSchema() {}
type TupleSchema struct {
	Elements  [] *Type
}

func (EnumSchema) KmdSchema() {}
type EnumSchema struct {
	CaseIndexMap  map[TypeId] uint
}

