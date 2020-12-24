package loader

import (
	"time"
	"os"
)


type Cache struct {
	Data  map[string] CacheItem
}

type CacheItem struct {
	ModTime  time.Time
	Result   EntryResult
}

type EntryResult struct {
	Module  *Module
	Index   Index
	Error   *Error
}

func MakeCache() Cache {
	return Cache {
		Data: make(map[string] CacheItem),
	}
}

func (c Cache) Put(path string, result EntryResult) {
	if result.Error != nil {
		c.Data[path] = CacheItem {
			ModTime: time.Unix(0, 1),
			Result:  result,
		}
	} else {
		c.Data[path] = CacheItem {
			ModTime: result.Module.FileInfo.ModTime(),
			Result:  result,
		}
	}
}

func (c Cache) Get(path string) (EntryResult, bool) {
	var item, exists = c.Data[path]
	if exists {
		fd, err := os.Open(path)
		if err != nil { return EntryResult{}, false }
		fd_info, err := fd.Stat()
		if err != nil { return EntryResult{}, false }
		var mod_time = fd_info.ModTime()
		_ = fd.Close()
		if mod_time.Equal(item.ModTime) && !(fd_info.IsDir()) {
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
		var mod, idx, _, err = LoadEntry(path)
		cache.Put(path, EntryResult {
			Module: mod,
			Index:  idx,
			Error:  err,
		})
		return mod, idx, err
	}
}

