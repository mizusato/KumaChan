package api

import (
	"time"
	"kumachan/rx"
	"kumachan/rpc"
	"kumachan/runtime/lib/librpc"
	. "kumachan/lang"
)


func rpcAdaptServerOptions(opts ProductValue) librpc.ServerOptions {
	return librpc.ServerOptions {
		CommonOptions: rpcAdaptCommonOptions(opts.Elements[0].(ProductValue)),
	}
}
func rpcAdaptClientOptions(opts ProductValue) librpc.ClientOptions {
	return librpc.ClientOptions {
		CommonOptions: rpcAdaptCommonOptions(opts.Elements[0].(ProductValue)),
	}
}
func rpcAdaptCommonOptions(opts ProductValue) librpc.CommonOptions {
	var debug = FromBool(opts.Elements[0].(SumValue))
	var limits = rpcAdaptLimitOptions(opts.Elements[1].(ProductValue))
	return librpc.CommonOptions {
		Debug:  debug,
		Limits: limits,
	}
}
func rpcAdaptLimitOptions(opts ProductValue) rpc.Limits {
	var ms = func(v Value) time.Duration {
		return (time.Millisecond * time.Duration(v.(uint)))
	}
	return rpc.Limits {
		SendTimeout:       ms(opts.Elements[0]),
		RecvTimeout:       ms(opts.Elements[1]),
		RecvInterval:      ms(opts.Elements[2]),
		RecvMaxObjectSize: opts.Elements[3].(uint),
	}
}

var RpcFunctions = map[string] interface{} {
	"rpc-server-cleartext-net": func(network String, addr String) librpc.ServerBackend {
		return librpc.ServerCleartextNet {
			Network: GoStringFromString(network),
			Address: GoStringFromString(addr),
		}
	},
	"rpc-client-cleartext-net": func(network String, addr String) librpc.ClientBackend {
		return librpc.ClientCleartextNet {
			Network: GoStringFromString(network),
			Address: GoStringFromString(addr),
		}
	},
	"rpc-connection-wait-closed": func(conn *rx.WrappedConnection) rx.Action {
		return conn.WaitClosed()
	},
	"rpc-connection-close": func(conn *rx.WrappedConnection) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			_ = conn.Close()
			return nil, true
		})
	},
	"rpc-serve": func (
		id        rpc.ServiceIdentifier,
		backend   librpc.ServerBackend,
		raw_opts  ProductValue,
		ctor      Value,
		h         InteropContext,
	) rx.Action {
		var api = h.GetRpcApi()
		var opts = rpcAdaptServerOptions(raw_opts)
		var wrapped_ctor = func(arg Value, conn Value) rx.Action {
			var pair = &ValProd { Elements: [] Value { arg, conn } }
			return h.Call(ctor, pair).(rx.Action)
		}
		return librpc.Serve(id, api, backend, opts, wrapped_ctor)
	},
	"rpc-access": func (
		id        rpc.ServiceIdentifier,
		backend   librpc.ClientBackend,
		raw_opts  ProductValue,
		argument  Value,
		consumer  Value,
		h         InteropContext,
	) rx.Action {
		var api = h.GetRpcApi()
		var opts = rpcAdaptClientOptions(raw_opts)
		var wrapped_consumer = func(instance Value) rx.Action {
			return h.Call(consumer, instance).(rx.Action)
		}
		return librpc.Access(id, api, backend, opts, argument, wrapped_consumer)
	},
}

