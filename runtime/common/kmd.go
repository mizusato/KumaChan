package common

import "kumachan/kmd"


type KmdApi interface {
	KmdSerialize(v Value, t *kmd.Type) ([] byte, error)
	KmdDeserialize(binary ([] byte), t *kmd.Type) (Value, error)
}

type KmdConfig struct {
	KmdSchemaTable
	KmdAdapterTable
	KmdValidatorTable
}

type KmdAdapterTable  map[kmd.AdapterId] KmdAdapterInfo
type KmdAdapterInfo   struct {
	Index  uint
}

type KmdValidatorTable  map[kmd.TypeId] KmdValidatorInfo
type KmdValidatorInfo   struct {
	Index  uint
}

type KmdSchemaTable  map[kmd.TypeId] KmdSchema
type KmdSchema       interface { KmdSchema() }
func (KmdRecordSchema) KmdSchema() {}
type KmdRecordSchema struct {
	Fields  map[string] KmdRecordField
}
type KmdRecordField struct {
	Type   *kmd.Type
	Index  uint
}
func (KmdTupleSchema) KmdSchema() {}
type KmdTupleSchema struct {
	Elements  [] *kmd.Type
}
func (KmdUnionSchema) KmdSchema() {}
type KmdUnionSchema struct {
	CaseIndexMap  map[kmd.TypeId] uint
}
