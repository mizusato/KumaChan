package api

import (
	"time"
	"math/big"
	"kumachan/misc/rx"
	"kumachan/misc/rpc"
	"kumachan/runtime/lib/librpc"
	. "kumachan/lang"
	"kumachan/misc/util"
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
	var log = opts.Elements[0].(ProductValue)
	var log_enabled = FromBool(log.Elements[0].(SumValue))
	var limits = rpcAdaptLimitOptions(opts.Elements[1].(ProductValue))
	return librpc.CommonOptions {
		LogEnabled: log_enabled,
		Limits:     limits,
	}
}
func rpcAdaptLimitOptions(opts ProductValue) rpc.Limits {
	var ms = func(v Value) time.Duration {
		var n = v.(*big.Int)
		return (time.Millisecond * time.Duration(util.GetUintNumber(n)))
	}
	return rpc.Limits {
		SendTimeout:       ms(opts.Elements[0]),
		RecvTimeout:       ms(opts.Elements[1]),
		RecvInterval:      ms(opts.Elements[2]),
		RecvMaxObjectSize: util.GetUintNumber(opts.Elements[3].(*big.Int)),
	}
}

var RpcFunctions = map[string] interface{} {
	"rpc-server-cleartext-net": func(network string, addr string) librpc.ServerBackend {
		return librpc.ServerCleartextNet {
			Network: network,
			Address: addr,
		}
	},
	"rpc-client-cleartext-net": func(network string, addr string) librpc.ClientBackend {
		return librpc.ClientCleartextNet {
			Network: network,
			Address: addr,
		}
	},
	"rpc-connection-close": func(conn *rx.WrappedConnection) rx.Observable {
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
	) rx.Observable {
		var api = h.GetRpcApi()
		var opts = rpcAdaptServerOptions(raw_opts)
		var wrapped_ctor = func(arg Value, conn Value) rx.Observable {
			var pair = Tuple(arg, conn)
			return h.Call(ctor, pair).(rx.Observable)
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
	) rx.Observable {
		var api = h.GetRpcApi()
		var opts = rpcAdaptClientOptions(raw_opts)
		var wrapped_consumer = func(instance Value) rx.Observable {
			return h.Call(consumer, instance).(rx.Observable)
		}
		return librpc.Access(id, api, backend, opts, argument, wrapped_consumer)
	},
}

