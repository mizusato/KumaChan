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
    "fmt"
    "math/rand"
    "kumachan/qt/cgohelper"
    "sync"
)


type Object struct {
    addr  unsafe.Pointer
}

type Widget struct {
    Object
}

type String C.QtString

func main() {
    var ui_bytes, err = ioutil.ReadFile("example.ui")
    if err != nil { panic(err) }
    var ui_str = string(ui_bytes)
    Init()
    Schedule(func() {
        window, ok := LoadWidget(ui_str)
        if !ok { panic("failed to load widget") }
        window.Show()
        label, ok := window.FindChild("label")
        if !ok { panic("unable to find label") }
        label.SetPropString("text", "你好世界")
        btn, ok := window.FindChild("button")
        if !ok { panic("unable to find button") }
        btn.Connect("clicked()", func() {
            label.SetPropString("text", fmt.Sprint(rand.Float64()))
        })
        input, ok := window.FindChild("input")
        input.Connect("textEdited(const QString&)", func() {
            var input_text = input.GetPropString("text")
            label.SetPropString("text", input_text)
            fmt.Printf("input_text: %s\n", input_text)
        })
    })
    <- chan struct{} (nil)
}

var initialized = false
var init_mutex sync.Mutex
func Init() {
    init_mutex.Lock()
    if !(initialized) {
        initialized = true
        init_mutex.Unlock()
        var wait = make(chan struct{}, 1)
        go (func() {
           runtime.LockOSThread()
           C.QtInit()
           wait <- struct{}{}
           C.QtMain()
        })()
        <- wait
    } else {
        init_mutex.Unlock()
    }
}

func Schedule(operation func()) {
    var f_id uint
    var f = func() {
        operation()
        cgohelper.UnregisterCallback(f_id)
    }
    f_id = cgohelper.RegisterCallback(f)
    C.QtSchedule(cgo_callback, C.size_t(f_id))
}

func LoadWidget(def string) (Widget, bool) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtLoadWidget(new_str(def))
    if ptr != nil {
        return Widget{Object{ptr}}, true
    } else {
        return Widget{}, false
    }
}

func (w Widget) FindChild(name string) (Widget, bool) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtWidgetFindChild(w.addr, new_str(name))
    if ptr != nil {
        return Widget{Object{ptr}}, true
    } else {
        return Widget{}, false
    }
}

func (w Widget) Show() {
    C.QtWidgetShow(w.addr)
}

func (obj Object) Connect(signal string, callback func()) func() {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var cb = cgohelper.RegisterCallback(callback)
    var conn = C.QtConnect(obj.addr, new_str(signal), cgo_callback, C.size_t(cb))
    if int(C.QtIsConnectionValid(conn)) != 0 {
        return func() {
            C.QtDisconnect(conn)
            // use Schedule() to prevent removing pending callbacks
            Schedule(func() {
                cgohelper.UnregisterCallback(cb)
            })
        }
    } else {
        panic("invalid connection")
    }
}

func (obj Object) GetPropQtString(prop string) String {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    return String(C.QtObjectGetPropString(obj.addr, new_str(prop)))
}

func (obj Object) SetPropQtString(prop string, val String) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropString(obj.addr, new_str(prop), C.QtString(val))
}

func (obj Object) GetPropRuneString(prop string) ([] rune) {
    return StringToRunes(obj.GetPropQtString(prop))
}

func (obj Object) SetPropRuneString(prop string, val ([] rune)) {
    var q_val, del_str = NewStringFromRunes(val)
    defer del_str()
    obj.SetPropQtString(prop, q_val)
}

func (obj Object) GetPropString(prop string) string {
    return string(obj.GetPropRuneString(prop))
}

func (obj Object) SetPropString(prop string, value string) {
    obj.SetPropRuneString(prop, ([] rune)(value))
}

func NewStringFromRunes(runes ([] rune)) (String, func()) {
    var str C.QtString
    if len(runes) > 0 {
        var ptr = (*C.uint32_t)(unsafe.Pointer(&runes[0]))
        var size = (C.size_t)(len(runes))
        str = C.QtNewStringUTF32(ptr, size)
    } else {
        str = C.QtNewStringUTF32(nil, 0)
    }
    return String(str), func() {
        C.QtDeleteString(str)
    }
}

func StringToRunes(str String) ([] rune) {
    var q_str = (C.QtString)(str)
    var size16 = uint(C.QtStringUTF16Length(q_str))
    var buf = make([] rune, size16)
    if size16 > 0 {
        var size32 = uint(C.QtStringWriteToUTF32Buffer(q_str,
            (*C.uint32_t)(unsafe.Pointer(&buf[0]))))
        buf = buf[:size32]
    }
    return buf
}

