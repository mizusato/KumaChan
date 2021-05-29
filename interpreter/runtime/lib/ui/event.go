package ui

import (
	"sync"
	"strconv"
	"kumachan/interpreter/runtime/lib/ui/vdom"
)


const handlerIdToStringBase = 16
var handlersLock sync.Mutex
var handlerNextId uint64 = 0
var handlers = make(map[uint64] *vdom.EventHandler)
var handlersGetId = make(map[*vdom.EventHandler] uint64)

func lookupEventHandler(id_str string) (*vdom.EventHandler, bool) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	var id, err = strconv.ParseUint(id_str, handlerIdToStringBase, 64)
	if err != nil { panic("something went wrong") }
	var handler, exists = handlers[id]
	return handler, exists
}

func registerEventHandler(handler *vdom.EventHandler) string {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	var id = handlerNextId
	handlerNextId += 1
	handlers[id] = handler
	handlersGetId[handler] = id
	return strconv.FormatUint(id, handlerIdToStringBase)
}

func unregisterEventHandler(handler *vdom.EventHandler) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	var id, exists = handlersGetId[handler]
	if !(exists) { panic("something went wrong") }
	delete(handlers, id)
	delete(handlersGetId, handler)
}

func getEventHandlerId(handler *vdom.EventHandler) (string, bool) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	var id, exists = handlersGetId[handler]
	if exists {
		return strconv.FormatUint(id, handlerIdToStringBase), true
	} else {
		return "", false
	}
}

