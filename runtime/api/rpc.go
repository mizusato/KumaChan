package api

import (
	"kumachan/rpc"
	"kumachan/runtime/lib/librpc"
	. "kumachan/lang"
)

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
	"rpc-serve": func (
		id           rpc.ServiceIdentifier,
		backend      librpc.ServerBackend,
		options      ProductValue,
		constructor  Value,
		h            InteropContext,
	) {
		panic("not implemented")
	},
	"rpc-access": func (
		id        rpc.ServiceIdentifier,
		backend   librpc.ClientBackend,
		options   ProductValue,
		argument  Value,
		consumer  Value,
	) {
		panic("not implemented")
	},
}

