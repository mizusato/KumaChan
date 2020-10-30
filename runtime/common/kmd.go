package common

import "kumachan/kmd"


type KmdApi interface {
	KmdGetTypeFromId(id kmd.TypeId) *kmd.Type
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

type KmdValidatorTable  map[kmd.ValidatorId] KmdValidatorInfo
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

func (conf KmdConfig) GetTypeFromId(id kmd.TypeId) *kmd.Type {
	switch conf.KmdSchemaTable[id].(type) {
	case KmdRecordSchema: return kmd.AlgebraicType(kmd.Record, id)
	case KmdTupleSchema:  return kmd.AlgebraicType(kmd.Tuple, id)
	case KmdUnionSchema:  return kmd.AlgebraicType(kmd.Union, id)
	default:              panic("something went wrong")
	}
}