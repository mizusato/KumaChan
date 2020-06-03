package cgohelper

import (
	"sync"
	"fmt"
)


var __GlobalCallbackRegistry = CreateCallbackRegistry()

func NewCallback(cb func()) (uint, (func() bool)) {
	var id = __GlobalCallbackRegistry.Register(cb)
	return id, func() bool {
		return __GlobalCallbackRegistry.Unregister(id)
	}
}

func GetCallback(id uint) func() {
	return __GlobalCallbackRegistry.Get(id)
}

type CallbackRegistry struct {
	mutex      sync.Mutex
	callbacks  map[uint] func()
	next       uint
	available  [] uint
}

func CreateCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry {
		callbacks: make(map[uint]func()),
		next:      0,
		available: make([] uint, 0),
	}
}

func (reg *CallbackRegistry) Register(callback func()) uint {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	var id = reg.next
	reg.next = (1 + id)
	reg.callbacks[id] = callback
	return id
}

func (reg *CallbackRegistry) Unregister(id uint) bool {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	var _, exists = reg.callbacks[id]
	if exists {
		delete(reg.callbacks, id)
		reg.available = append(reg.available, id)
		return true
	} else {
		return false
	}
}

func (reg *CallbackRegistry) Get(id uint) func()  {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	var callback, exists = reg.callbacks[id]
	if exists {
		return callback
	} else {
		panic(fmt.Sprintf("cannot find callback of id %d", id))
	}
}

