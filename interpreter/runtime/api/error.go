package api

import (
	"errors"
	. "kumachan/interpreter/base"
	"kumachan/standalone/rpc"
	"kumachan/interpreter/runtime/lib/container"
)


var ErrorFunctions = map[string] interface{} {
	"make-error": func(msg string) error {
		return errors.New(msg)
	},
	"make-error-with-data": func(msg string, data_ Value) error {
		var data = make(map[string] string)
		var data_items = container.ListFrom(data_)
		data_items.ForEach(func(i uint, p_ Value) {
			var p = p_.(TupleValue)
			var k = p.Elements[0].(string)
			var v = p.Elements[1].(string)
			data[k] = v
		})
		return &rpc.ErrorWithExtraData {
			Desc: msg,
			Data: data,
		}
	},
	"error-get-data": func(e error, opts TupleValue) string {
		var key = opts.Elements[0].(string)
		var fallback = opts.Elements[1].(string)
		var e_with_extra, with_extra = e.(*rpc.ErrorWithExtraData)
		if with_extra {
			var value, exists = e_with_extra.Data[key]
			if exists {
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
			var wrapped_str = h.Call(f, desc).(string)
			return wrapped_str
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

