package rx

import (
	"net/url"
	"net/http"
	"io/ioutil"
)


type HttpResponse struct {
	StatusCode  uint
	Header      http.Header
	Body        [] byte
}

func HttpGet(url *url.URL) Effect {
	return CreateEffect(func(sender Sender) {
		if sender.Context().AlreadyCancelled() {
			return
		}
		res, err := http.Get(url.String())
		if err != nil {
			sender.Error(err)
			return
		}
		if sender.Context().AlreadyCancelled() {
			_ = res.Body.Close()
			return
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			sender.Error(err)
			return
		}
		_ = res.Body.Close()
		sender.Next(HttpResponse {
			Body:       body,
			Header:     res.Header,
			StatusCode: uint(res.StatusCode),
		})
		sender.Complete()
	})
}

