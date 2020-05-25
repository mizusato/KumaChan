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
    (func() {
        var ui_bytes, err = ioutil.ReadFile("example.ui")
        if err != nil { panic(err) }
        var ui_str = string(ui_bytes)
        var ui_c_str *C.char = C.CString(ui_str)
        defer C.free(unsafe.Pointer(ui_c_str))
        var w = C.QtLoadWidget(ui_c_str)
        if w == nil { panic("failed to load widget") }
        C.QtWidgetShow(w)
        var label_name *C.char = C.CString("label")
        defer C.free(unsafe.Pointer(label_name))
        var label = C.QtWidgetFindChild(w, label_name)
        if label == nil { panic("unable to find label") }
        var prop_text *C.char = C.CString("text")
        defer C.free(unsafe.Pointer(prop_text))
        var buf [100]rune
        var text_content = []rune("你好世界")
        copy(buf[:], text_content)
        var text = C.QtNewStringUTF32((*C.uint32_t)(unsafe.Pointer(&buf)), (C.size_t)(uint(len(text_content))))
        defer C.QtDeleteString(text)
        C.QtObjectSetPropString(label, prop_text, text)
    })()
    C.QtMainLoop()
}

