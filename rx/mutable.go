package rx


type Cell struct {
	value   Object
}
func CreateCell(init_val Object) *Cell {
	return &Cell { init_val }
}
func (c *Cell) Get() Action {
	return NewSync(func()(Object, bool) {
		return c.value, true
	})
}
func (c *Cell) Set(new_val Object) Action {
	return NewSync(func()(Object, bool) {
		c.value = new_val
		return nil, true
	})
}
func (c *Cell) Swap(f func(Object)Object) Action {
	return NewSync(func()(Object, bool) {
		c.value = f(c.value)
		return nil, true
	})
}

type StringHashMap  map[string] Object
func CreateStringHashMap(m map[string] Object) StringHashMap {
	return StringHashMap(m)
}
func (m StringHashMap) Has(key string) Action {
	return NewSync(func()(Object, bool) {
		var _, exists = m[key]
		return exists, true
	})
}
func (m StringHashMap) Get(key string) Action {
	return NewSync(func()(Object, bool) {
		var val, exists = m[key]
		if exists {
			return Optional { true, val }, true
		} else {
			return Optional {}, true
		}
	})
}
func (m StringHashMap) Set(key string, val Object) Action {
	return NewSync(func()(Object, bool) {
		m[key] = val
		return nil, true
	})
}
func (m StringHashMap) Delete(key string) Action {
	return NewSync(func()(Object, bool) {
		var deleted, exists = m[key]
		if exists {
			delete(m, key)
			return Optional { true, deleted }, true
		} else {
			return Optional {}, true
		}
	})
}

type NumberHashMap  map[uint] Object
func CreateNumberHashMap(m map[uint] Object) NumberHashMap {
	return NumberHashMap(m)
}
func (m NumberHashMap) Has(key uint) Action {
	return NewSync(func()(Object, bool) {
		var _, exists = m[key]
		return exists, true
	})
}
func (m NumberHashMap) Get(key uint) Action {
	return NewSync(func()(Object, bool) {
		var val, exists = m[key]
		if exists {
			return Optional { true, val }, true
		} else {
			return Optional {}, true
		}
	})
}
func (m NumberHashMap) Set(key uint, val Object) Action {
	return NewSync(func()(Object, bool) {
		m[key] = val
		return nil, true
	})
}
func (m NumberHashMap) Delete(key uint) Action {
	return NewSync(func()(Object, bool) {
		var deleted, exists = m[key]
		if exists {
			delete(m, key)
			return Optional { true, deleted }, true
		} else {
			return Optional {}, true
		}
	})
}

type Buffer struct {
	data  *([] byte)
}
func CreateBuffer(capacity uint) Buffer {
	var data = make([] byte, 0, capacity)
	return Buffer { &data }
}
func (buf Buffer) Write(bytes ([] byte)) Action {
	return NewSync(func() (Object, bool) {
		*buf.data = append(*buf.data, bytes...)
		return nil, true
	})
}
func (buf Buffer) Dump() Action {
	return NewGoroutine(func(sender Sender) {
		var dumped = make([] byte, len(*buf.data))
		copy(dumped, *buf.data)
		sender.Next(dumped)
		sender.Complete()
	})
}

