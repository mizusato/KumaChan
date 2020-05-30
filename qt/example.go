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
    "reflect"
    "unsafe"
    "fmt"
    "math/rand"
    "kumachan/qt/cgohelper"
)


func main() {
    runtime.LockOSThread()
    C.QtInit()
    (func() {
        var new_str, del_all_str = str_alloc()
        defer del_all_str()
        var ui_bytes, err = ioutil.ReadFile("example.ui")
        if err != nil { panic(err) }
        var ui_str = string(ui_bytes)
        var w = C.QtLoadWidget(new_str(ui_str))
        if w == nil { panic("failed to load widget") }
        C.QtWidgetShow(w)
        var label = C.QtWidgetFindChild(w, new_str("label"))
        if label == nil { panic("unable to find label") }
        var text, del_text = NewStringFromRunes([]rune("你好世界"))
        defer del_text()
        C.QtObjectSetPropString(label, new_str("text"), text)
        var cb = cgohelper.RegisterCallback(func() {
            var new_str, del_all_str = str_alloc()
            defer del_all_str()
            var x, del_x = NewStringFromRunes([]rune(fmt.Sprint(rand.Float64())))
            defer del_x()
            C.QtObjectSetPropString(label, new_str("text"), x)
        })
        var btn = C.QtWidgetFindChild(w, new_str("button"))
        if btn == nil { panic("unable to find button") }
        C.QtConnect(btn, new_str("clicked()"), cgo_callback, C.size_t(cb))
    })()
    C.QtMainLoop()
}

func NewStringFromBuf(buf ([] uint32)) (C.QtString, func()) {
    var slice = *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
    var ptr = (*C.uint32_t)(unsafe.Pointer(slice.Data))
    var size = (C.size_t)(slice.Len)
    var str = C.QtNewStringUTF32(ptr, size)
    return str, func() {
        C.QtDeleteString(str)
    }
}

func NewStringFromRunes(runes ([] rune)) (C.QtString, func()) {
    var slice = *(*reflect.SliceHeader)(unsafe.Pointer(&runes))
    var ptr = (*C.uint32_t)(unsafe.Pointer(slice.Data))
    var size = (C.size_t)(slice.Len)
    var str = C.QtNewStringUTF32(ptr, size)
    return str, func() {
        C.QtDeleteString(str)
    }
}

