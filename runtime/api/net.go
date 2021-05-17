package api

import (
	"net/url"
	"kumachan/misc/rx"
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
)


var NetFunctions = map[string] interface{} {
	"parse-url": func(str string) SumValue {
		var url, err = url.Parse(str)
		if err != nil {
			return None()
		} else {
			return Some(url)
		}
	},
	"url!": func(str string) *url.URL {
		var url, err = url.Parse(str)
		if err != nil { panic(err) }
		return url
	},
	"http-response-status-code": func(res rx.HttpResponse) uint {
		return res.StatusCode
	},
	"http-response-header": func(res rx.HttpResponse) Map {
		var h = res.Header
		var m = NewMapOfStringKey()
		for k, v := range h {
			m, _ = m.Inserted(k, v)
		}
		return m
	},
	"http-response-body": func(res rx.HttpResponse) ([] byte) {
		return res.Body
	},
	"http-get": func(url *url.URL) rx.Observable {
		return rx.HttpGet(url)
	},
}
