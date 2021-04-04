package api

import (
	"errors"
	. "kumachan/lang"
	"kumachan/misc/rpc"
	"kumachan/runtime/lib/container"
)


var ErrorFunctions = map[string] interface{} {
	"make-error": func(msg String) error {
		return errors.New(GoStringFromString(msg))
	},
	"make-error-with-data": func(msg_ String, data_ Value) error {
		var msg = GoStringFromString(msg_)
		var data = make(map[string] string)
		var data_items = container.ArrayFrom(data_)
		for i := uint(0); i < data_items.Length; i += 1 {
			var p = data_items.GetItem(i).(ProductValue)
			var k = GoStringFromString(p.Elements[0].(String))
			var v = GoStringFromString(p.Elements[1].(String))
			data[k] = v
		}
		return &rpc.ErrorWithExtraData {
			Desc: msg,
			Data: data,
		}
	},
	"error-get-data": func(e error, opts ProductValue) String {
		var key = opts.Elements[0].(String)
		var fallback = opts.Elements[1].(String)
		var e_with_extra, with_extra = e.(*rpc.ErrorWithExtraData)
		if with_extra {
			var k = GoStringFromString(key)
			var v, exists = e_with_extra.Data[k]
			if exists {
				var value = StringFromGoString(v)
				return value
			} else {
				return fallback
			}
		} else {
			return fallback
		}
	},
}

