package qt

/*
#include <stdlib.h>
typedef void (*CgoCallback)(size_t);
void CgoInvokeCallback(size_t id);
*/
import "C"
import "kumachan/runtime/lib/ui/qt/cgohelper"

func str_alloc() (func(string) *C.char, func()) {
	var alloc, dealloc = cgohelper.CreateStringAllocator()
	return func(str string) *C.char {
		return (*C.char)(alloc(str))
	}, dealloc
}

//export invoke_callback
func invoke_callback(id C.size_t) {
	var f = cgohelper.GetCallback(uint(id))
	f()
}

var cgo_callback = C.CgoCallback(C.CgoInvokeCallback)
