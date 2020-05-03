package lib

import (
	"net/url"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
	"kumachan/runtime/common/rx"
)


var NetFunctions = map[string] interface{} {
	"parse-url": func(str String) SumValue {
		var url, err = url.Parse(string(str))
		if err != nil {
			return Na()
		} else {
			return Just(url)
		}
	},
	"url": func(str String) *url.URL {
		var url, err = url.Parse(string(str))
		if err != nil { panic(err) }
		return url
	},
	"http-response-status-code": func(res rx.HttpResponse) uint {
		return res.StatusCode
	},
	"http-response-header": func(res rx.HttpResponse) Map {
		var h = res.Header
		var m = NewStrMap()
		for k, v := range h {
			m, _ = m.Inserted(k, v)
		}
		return m
	},
	"http-response-body": func(res rx.HttpResponse) ([] byte) {
		return res.Body
	},
	"http-get": func(url *url.URL) rx.Effect {
		return rx.HttpGet(url)
	},
}
