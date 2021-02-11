package lang

import (
	"kumachan/rpc/kmd"
	"kumachan/rpc"
)


type KmdApi interface {
	GetTypeFromId(id kmd.TypeId) *kmd.Type
	Serialize(v Value, t *kmd.Type) ([] byte, error)
	Deserialize(binary ([] byte), t *kmd.Type) (Value, error)
	rpc.KmdApi
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
	case kmd.EnumSchema:   return kmd.AlgebraicType(kmd.Enum, id)
	default:               panic("something went wrong")
	}
}

