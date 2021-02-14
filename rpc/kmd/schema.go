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

func (table SchemaTable) GetTypeFromId(id TypeId) *Type {
	switch table[id].(type) {
	case RecordSchema: return AlgebraicType(Record, id)
	case TupleSchema:  return AlgebraicType(Tuple, id)
	case EnumSchema:   return AlgebraicType(Enum, id)
	default:           panic("given type id not found in schema table")
	}
}

