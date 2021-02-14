package lang

import (
	"fmt"
	"kumachan/rx"
	"kumachan/rpc"
	"kumachan/rpc/kmd"
)


type RpcApi interface {
	GetKmdApi() KmdApi
	GetServiceInterface(rpc.ServiceIdentifier) (rpc.ServiceInterface, bool)
}
type RpcInfo struct {
	rpc.ServiceIndex
}
type ServiceInstance interface {
	Call(name string, arg Value) rx.Action
}
type ServerSideServiceInstance struct {
	data     Value
	methods  map[string] (func(data Value, arg Value) rx.Action)
}
type ClientSideServiceInstance struct {
	underlying  *rpc.ClientInstance
}
func (instance ServerSideServiceInstance) Call(name string, arg Value) rx.Action {
	var f, exists = instance.methods[name]
	if !(exists) { panic("something went wrong") }
	return f(instance.data, arg)
}
func (instance ClientSideServiceInstance) Call(name string, arg Value) rx.Action {
	return instance.underlying.Call(name, arg)
}
func CreateServiceMethodCaller(method_name string) NativeFunctionValue {
	return NativeFunctionValue(func(arg Value, h InteropContext) Value {
		var prod = arg.(ProductValue)
		var instance = prod.Elements[0].(ServiceInstance)
		var method_arg = prod.Elements[1]
		return instance.Call(method_name, method_arg)
	})
}
func CreateServiceInstance(data Value, impl ([] Value), names ([] string), h InteropContext) ServiceInstance {
	var methods = make(map[string] func(Value,Value)(rx.Action))
	for i, name := range names {
		var index = i
		methods[name] = func(data Value, arg Value) rx.Action {
			var pair = &ValProd { Elements: [] Value { data, arg } }
			var ret = h.Call(impl[index], pair)
			return ret.(rx.Action)
		}
	}
	return ServerSideServiceInstance {
		data:    data,
		methods: methods,
	}
}
func AdaptServiceInstance(instance *rpc.ClientInstance) ServiceInstance {
	return ClientSideServiceInstance {
		underlying: instance,
	}
}

type KmdApi interface {
	GetTypeFromId(id kmd.TypeId) *kmd.Type
	Serialize(v Value, t *kmd.Type) ([] byte, error)
	Deserialize(binary ([] byte), t *kmd.Type) (Value, error)
	rpc.KmdApi
}
type KmdInfo struct {
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
func CreateKmdApiFunction(id kmd.TransformerPartId) NativeFunctionValue {
	switch id := id.(type) {
	case kmd.SerializerId:
		return func(arg Value, h InteropContext) Value {
			var api = h.GetKmdApi()
			var t = api.GetTypeFromId(id.TypeId)
			var binary, err = api.Serialize(arg, t)
			if err != nil {
				var wrapped = fmt.Errorf("serialiation error: %w", err)
				panic(wrapped)
			}
			return binary
		}
	case kmd.DeserializerId:
		return func(arg Value, h InteropContext) Value {
			var api = h.GetKmdApi()
			var t = api.GetTypeFromId(id.TypeId)
			var obj, err = api.Deserialize(arg.([] byte), t)
			if err != nil { return Ng(err) }
			return Ok(obj)
		}
	default:
		panic("impossible branch")
	}
}

