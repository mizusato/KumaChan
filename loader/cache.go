package loader

import "time"


type Cache struct {
	Data  map[string] CacheItem
	Keep  time.Duration
}

type CacheItem struct {
	Expire  time.Time
	Result  EntryResult
}

type EntryResult struct {
	Module  *Module
	Index   Index
	Error   *Error
}

func MakeCache(keep time.Duration) Cache {
	return Cache {
		Data: make(map[string] CacheItem),
		Keep: keep,
	}
}

func (c Cache) SweepExpired() {
	var now = time.Now()
	var expired = make([] string, 0)
	for k, v := range c.Data {
		if now.Sub(v.Expire) > c.Keep {
			expired = append(expired, k)
		}
	}
	for _, k := range expired {
		delete(c.Data, k)
	}
}

func (c Cache) Put(path string, result EntryResult) {
	var now = time.Now()
	c.Data[path] = CacheItem {
		Expire: now.Add(c.Keep),
		Result: result,
	}
}

func (c Cache) Get(path string) (EntryResult, bool) {
	var now = time.Now()
	var item, exists = c.Data[path]
	if exists {
		if now.Sub(item.Expire) > c.Keep {
			return item.Result, true
		} else {
			return EntryResult{}, false
		}
	} else {
		return EntryResult{}, false
	}
}

func LoadEntryWithCache(path string, cache Cache) (*Module, Index, *Error) {
	var cached, is_cached = cache.Get(path)
	if is_cached {
		return cached.Module, cached.Index, cached.Error
	} else {
		var mod, idx, err = LoadEntry(path)
		cache.Put(path, EntryResult {
			Module: mod,
			Index:  idx,
			Error:  err,
		})
		return mod, idx, err
	}
}

