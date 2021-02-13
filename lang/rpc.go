package lang

import (
	"kumachan/rx"
	"kumachan/rpc"
	"kumachan/rpc/kmd"
)


type ServiceInstance struct {
	data     Value
	methods  map[string] (func(data Value, arg Value) rx.Action)
}
func (instance ServiceInstance) Call(name string, arg Value) rx.Action {
	var method, exists = instance.methods[name]
	if !(exists) { panic("something went wrong") }
	return method(instance.data, arg)
}

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

