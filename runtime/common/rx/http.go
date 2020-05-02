package rx

import (
	"net/url"
	"net/http"
	"io/ioutil"
)


type HttpResponse struct {
	Data        [] byte
	Header      http.Header
	StatusCode  uint
}

func HttpGet(url url.URL) Effect {
	return CreateEffect(func(sender Sender) {
		var cancel, cancellable = sender.CancelSignal()
		res, err := http.Get(url.String())
		if err != nil {
			sender.Error(err)
			return
		}
		if cancellable {
			select {
			case <- cancel:
				_ = res.Body.Close()
				return
			default:
			}
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			sender.Error(err)
			return
		}
		_ = res.Body.Close()
		sender.Next(HttpResponse {
			Data:       data,
			Header:     res.Header,
			StatusCode: uint(res.StatusCode),
		})
		sender.Complete()
	})
}
