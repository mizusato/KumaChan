package main

/*
#include <stdlib.h>
#include "qtbinding/qtbinding.h"
*/
// #cgo LDFLAGS: -L./build -lqtbinding
import "C"

import (
    "io/ioutil"
    "runtime"
    "unsafe"
)


func main() {
    runtime.LockOSThread()
    C.QtInit()
    var ui_bytes, err = ioutil.ReadFile("example.ui")
    if err != nil { panic(err) }
    var ui_str = string(ui_bytes)
    var ui_c_str *C.char = C.CString(ui_str)
    var w = C.QtLoadWidget(ui_c_str)
    C.free(unsafe.Pointer(ui_c_str))
    if w == nil { panic("failed to load widget") }
    C.QtWidgetShow(w)
    C.QtMainLoop()
}

