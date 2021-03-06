package def

import (
	"fmt"
	"kumachan/standalone/rx"
	"kumachan/standalone/rpc"
	"kumachan/standalone/rpc/kmd"
)


type RpcApi interface {
	GetKmdApi() KmdApi
	GetServiceInterface(rpc.ServiceIdentifier) (rpc.ServiceInterface, bool)
}
type RpcInfo struct {
	rpc.ServiceIndex
}
type ServiceInstance interface {
	Call(name string, arg Value) rx.Observable
}
type ServerSideServiceInstance struct {
	data     Value
	dtor     func(data Value) rx.Observable
	methods  map[string] (func(data Value, arg Value) rx.Observable)
}
type ClientSideServiceInstance struct {
	underlying  *rpc.ClientInstance
}
func (instance ServerSideServiceInstance) Delete() rx.Observable {
	return instance.dtor(instance.data)
}
func (instance ServerSideServiceInstance) Call(name string, arg Value) rx.Observable {
	var f, exists = instance.methods[name]
	if !(exists) { panic("something went wrong") }
	return f(instance.data, arg)
}
func (instance ClientSideServiceInstance) Call(name string, arg Value) rx.Observable {
	return instance.underlying.Call(name, arg)
}
func CreateServiceMethodCaller(method_name string) NativeFuncValue {
	return ValNativeFunc(func(arg Value, h InteropContext) Value {
		var prod = arg.(TupleValue)
		var instance = prod.Elements[0].(ServiceInstance)
		var method_arg = prod.Elements[1]
		return instance.Call(method_name, method_arg)
	})
}
func CreateServiceInstance (
	data          Value,
	dtor_impl     Value,
	methods_impl  [] Value,
	method_names  [] string,
	h             InteropContext,
) ServiceInstance {
	var methods = make(map[string] func(Value,Value)(rx.Observable))
	for i, name := range method_names {
		var index = i
		methods[name] = func(data Value, arg Value) rx.Observable {
			var pair = Tuple(data, arg)
			var ret = h.Call(methods_impl[index], pair)
			return ret.(rx.Observable)
		}
	}
	var dtor = func(data Value) rx.Observable {
		return h.Call(dtor_impl, data).(rx.Observable)
	}
	return ServerSideServiceInstance {
		data:    data,
		dtor:    dtor,
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
	Symbol  Symbol
}
type KmdValidatorTable  map[kmd.ValidatorId] KmdValidatorInfo
type KmdValidatorInfo   struct {
	Symbol  Symbol
}
func CreateKmdApiFunction(id kmd.TransformerPartId) NativeFuncValue {
	switch id := id.(type) {
	case kmd.SerializerId:
		return ValNativeFunc(func(arg Value, h InteropContext) Value {
			var api = h.GetKmdApi()
			var t = api.GetTypeFromId(id.TypeId)
			var binary, err = api.Serialize(arg, t)
			if err != nil {
				var wrapped = fmt.Errorf("serialiation error: %w", err)
				panic(wrapped)
			}
			return binary
		})
	case kmd.DeserializerId:
		return ValNativeFunc(func(arg Value, h InteropContext) Value {
			var api = h.GetKmdApi()
			var t = api.GetTypeFromId(id.TypeId)
			var obj, err = api.Deserialize(arg.([] byte), t)
			if err != nil { return Ng(err) }
			return Ok(obj)
		})
	default:
		panic("impossible branch")
	}
}

