package librpc

import (
	"fmt"
	"net"
	"kumachan/standalone/rx"
	"kumachan/standalone/rpc"
	"kumachan/standalone/rpc/kmd"
	. "kumachan/interpreter/def"
)


func Access (
	id        rpc.ServiceIdentifier,
	api       RpcApi,
	backend   ClientBackend,
	options   ClientOptions,
	argument  Value,
	consumer  func(Value)(rx.Observable),
) rx.Observable {
	return rx.NewSync(func() (rx.Object, bool) {
		var conn, err = backend.Access()
		if err != nil { return err, false }
		return conn, true
	}).Then(func(conn_ rx.Object) rx.Observable {
		var conn = conn_.(net.Conn)
		var service, ok = api.GetServiceInterface(id)
		if !(ok) { panic(fmt.Sprintf("service %s does not exist", id)) }
		var wrapped_consumer = func(instance *rpc.ClientInstance) rx.Observable {
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
	constructor  func(arg Value, conn Value)(rx.Observable),
) rx.Observable {
	return rx.NewSync(func() (rx.Object, bool) {
		var l, err = backend.Serve()
		if err != nil { return err, false }
		return l, true
	}).Then(func(l_ rx.Object) rx.Observable {
		var l = l_.(net.Listener)
		var service, ok = api.GetServiceInterface(id)
		if !(ok) { panic(fmt.Sprintf("service %s does not exist", id)) }
		var destructor = func(instance Value) rx.Observable {
			return instance.(ServerSideServiceInstance).Delete()
		}
		var service_impl = implementService(service, constructor, destructor)
		var full_options = &rpc.ServerOptions {
			Listener:    l,
			DebugOutput: options.GetDebugOutput(),
			Limits:      options.Limits,
			KmdApi:      api.GetKmdApi(),
		}
		return rpc.Server(service_impl, full_options)
	})
}

func implementService(i rpc.ServiceInterface, ctor (func(Value,Value) rx.Observable), dtor (func(Value) rx.Observable)) rpc.Service {
	var methods = make(map[string] rpc.ServiceMethod)
	for name, method_info := range i.Methods {
		var name = name  // spent 3 hours to find out this problem
		methods[name] = rpc.ServiceMethod {
			ServiceMethodInterface: method_info,
			GetAction: func(instance kmd.Object, arg kmd.Object) rx.Observable {
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
		Destructor: rpc.ServiceDestructor {
			GetAction: dtor,
		},
		Methods: methods,
	}
	return service
}

