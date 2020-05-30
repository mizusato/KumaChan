package cgohelper

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"


func CreateStringAllocator() (func(string) unsafe.Pointer, func()) {
	var allocated = make([] *C.char, 0)
	var allocate = func(str string) unsafe.Pointer {
		var ptr = C.CString(str)
		allocated = append(allocated, ptr)
		return unsafe.Pointer(ptr)
	}
	var deallocate = func() {
		for _, ptr := range allocated {
			C.free(unsafe.Pointer(ptr))
		}
	}
	return allocate, deallocate
}

