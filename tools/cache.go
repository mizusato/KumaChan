package tools

import "time"


type LinterCache struct {
	Data  map[LintRequest] LinterCacheItem
	Keep  time.Duration
}

type LinterCacheItem struct {
	Expire    time.Time
	Response  LintResponse
}

func MakeLinterCache(keep time.Duration) LinterCache {
	return LinterCache {
		Data: make(map[LintRequest] LinterCacheItem),
		Keep: keep,
	}
}

func (c LinterCache) SweepExpired() {
	var now = time.Now()
	var expired = make([] LintRequest, 0)
	for k, v := range c.Data {
		if now.Sub(v.Expire) > c.Keep {
			expired = append(expired, k)
		}
	}
	for _, k := range expired {
		delete(c.Data, k)
	}
}

func (c LinterCache) Put(req LintRequest, res LintResponse) {
	var now = time.Now()
	c.Data[req] = LinterCacheItem {
		Expire:   now.Add(c.Keep),
		Response: res,
	}
}

func (c LinterCache) Get(req LintRequest) (LintResponse, bool) {
	var now = time.Now()
	var item, exists = c.Data[req]
	if exists {
		if now.Sub(item.Expire) < c.Keep {
			return item.Response, true
		} else {
			return LintResponse{}, false
		}
	} else {
		return LintResponse{}, false
	}
}

