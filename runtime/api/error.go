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
		var data_items = container.ListFrom(data_)
		data_items.ForEach(func(i uint, p_ Value) {
			var p = p_.(ProductValue)
			var k = GoStringFromString(p.Elements[0].(String))
			var v = GoStringFromString(p.Elements[1].(String))
			data[k] = v
		})
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
	"error-wrap": func(e error, f Value, h InteropContext) error {
		var wrap_desc = func(desc string) string {
			var desc_str = StringFromGoString(desc)
			var wrapped_str = h.Call(f, desc_str).(String)
			return GoStringFromString(wrapped_str)
		}
		var e_with_extra, with_extra = e.(*rpc.ErrorWithExtraData)
		if with_extra {
			return &rpc.ErrorWithExtraData {
				Desc: wrap_desc(e_with_extra.Desc),
				Data: e_with_extra.Data,
			}
		} else {
			return errors.New(wrap_desc(e.Error()))
		}
	},
}

