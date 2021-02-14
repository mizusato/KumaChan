package librpc

import (
	"fmt"
	"net"
	"kumachan/rx"
	"kumachan/rpc"
	"kumachan/rpc/kmd"
	. "kumachan/lang"
)


func Access (
	id        rpc.ServiceIdentifier,
	api       RpcApi,
	backend   ClientBackend,
	options   ClientOptions,
	argument  Value,
	consumer  func(Value)(rx.Action),
) rx.Action {
	return rx.NewSync(func() (rx.Object, bool) {
		var conn, err = backend.Access()
		if err != nil { return err, false }
		return conn, true
	}).Then(func(conn_ rx.Object) rx.Action {
		var conn = conn_.(net.Conn)
		var service, ok = api.GetServiceInterface(id)
		if !(ok) { panic(fmt.Sprintf("service %s does not exist", id)) }
		var wrapped_consumer = func(instance *rpc.ClientInstance) rx.Action {
			return consumer(AdaptServiceInstance(instance))
		}
		var full_options = &rpc.ClientOptions {
			Connection:          conn,
			DebugOutput:         options.GetDebugOutput(),
			ConstructorArgument: argument,
			InstanceConsumer:    wrapped_consumer,
			Limits:              options.Limits,
			KmdApi:              api.GetKmdApi(),
		}
		return rpc.Client(service, full_options)
	})
}

func Serve (
	id           rpc.ServiceIdentifier,
	api          RpcApi,
	backend      ServerBackend,
	options      ServerOptions,
	constructor  func(Value)(rx.Action),
) rx.Action {
	return rx.NewSync(func() (rx.Object, bool) {
		var l, err = backend.Serve()
		if err != nil { return err, false }
		return l, true
	}).Then(func(l_ rx.Object) rx.Action {
		var l = l_.(net.Listener)
		var service, ok = api.GetServiceInterface(id)
		if !(ok) { panic(fmt.Sprintf("service %s does not exist", id)) }
		var service_impl = implementService(service, constructor)
		var full_options = &rpc.ServerOptions {
			Listener:    l,
			DebugOutput: options.GetDebugOutput(),
			Limits:      options.Limits,
			KmdApi:      api.GetKmdApi(),
		}
		return rpc.Server(service_impl, full_options)
	})
}

func implementService(i rpc.ServiceInterface, ctor (func(Value) rx.Action)) rpc.Service {
	var methods = make(map[string] rpc.ServiceMethod)
	for name, method_info := range i.Methods {
		methods[name] = rpc.ServiceMethod {
			ServiceMethodInterface: method_info,
			GetAction: func(instance kmd.Object, arg kmd.Object) rx.Action {
				return instance.(ServerSideServiceInstance).Call(name, arg)
			},
		}
	}
	var service = rpc.Service {
		ServiceIdentifier: i.ServiceIdentifier,
		Constructor: rpc.ServiceConstructor {
			ServiceConstructorInterface: i.Constructor,
			GetAction: ctor,
		},
		Methods: methods,
	}
	return service
}

