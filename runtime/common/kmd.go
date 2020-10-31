package common

import "kumachan/kmd"


type KmdApi interface {
	KmdGetTypeFromId(id kmd.TypeId) *kmd.Type
	KmdSerialize(v Value, t *kmd.Type) ([] byte, error)
	KmdDeserialize(binary ([] byte), t *kmd.Type) (Value, error)
}

type KmdConfig struct {
	kmd.SchemaTable
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

func (conf KmdConfig) GetTypeFromId(id kmd.TypeId) *kmd.Type {
	switch conf.SchemaTable[id].(type) {
	case kmd.RecordSchema: return kmd.AlgebraicType(kmd.Record, id)
	case kmd.TupleSchema:  return kmd.AlgebraicType(kmd.Tuple, id)
	case kmd.UnionSchema:  return kmd.AlgebraicType(kmd.Union, id)
	default:               panic("something went wrong")
	}
}