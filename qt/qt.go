package qt

/*
#include <stdlib.h>
#include "qtbinding/qtbinding.h"
*/
// #cgo LDFLAGS: -L./build -lqtbinding
import "C"

import (
    "runtime"
    "unsafe"
    "sync"
    "kumachan/qt/cgohelper"
)


type Object interface {
    Object()
    ptr()  unsafe.Pointer
}
func (obj object) Object() {}
func (obj object) ptr() unsafe.Pointer { return obj.addr }
type object struct { addr unsafe.Pointer }

type Widget interface {
    Object
    Widget()
}
func (widget) Widget() {}
type widget struct { object }

type String C.QtString

type EventKind uint
func EventMove() EventKind { return EventKind(uint(C.QtEventMove)) }
func EventResize() EventKind { return EventKind(uint(C.QtEventResize)) }

type Event C.QtEvent

//func main() {
//    var ui_bytes, err = ioutil.ReadFile("example.ui")
//    if err != nil { panic(err) }
//    var ui_str = string(ui_bytes)
//    MakeSureInitialized()
//    CommitTask(func() {
//        window, ok := LoadWidget(ui_str)
//        if !ok { panic("failed to load widget") }
//        Listen(window, EventResize(), false, func(ev Event) {
//            var width = ev.ResizeEventGetWidth()
//            var height = ev.ResizeEventGetHeight()
//            fmt.Printf("Resize Event: (%d, %d)\n", width, height)
//        })
//        Show(window)
//        label, ok := FindChild(window, "label")
//        if !ok { panic("unable to find label") }
//        SetPropString(label, "text", "你好世界")
//        btn, ok := FindChild(window, "button")
//        if !ok { panic("unable to find button") }
//        Connect(btn, "clicked()", func() {
//            SetPropString(label, "text", fmt.Sprint(rand.Float64()))
//        })
//        input, ok := FindChild(window, "input")
//        Connect(input, "textEdited(const QString&)", func() {
//            var input_text = GetPropString(input, "text")
//            SetPropString(label, "text", input_text)
//            fmt.Printf("input_text: %s\n", input_text)
//        })
//    })
//    <- chan struct{} (nil)
//}

var initialized = false
var init_mutex sync.Mutex
func MakeSureInitialized() {
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

func CommitTask(operation func()) {
    var delete_callback (func() bool)
    var f = func() {
        operation()
        delete_callback()
    }
    callback, delete_callback := cgohelper.NewCallback(f)
    C.QtCommitTask(cgo_callback, C.size_t(callback))
}

func LoadWidget(def string, dir string) (Widget, bool) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtLoadWidget(new_str(def), new_str(dir))
    if ptr != nil {
        return widget{object{ptr}}, true
    } else {
        return widget{}, false
    }
}

func FindChild(w Widget, name string) (Widget, bool) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtWidgetFindChild(w.ptr(), new_str(name))
    if ptr != nil {
        return widget{object{ptr}}, true
    } else {
        return widget{}, false
    }
}

func Show(w Widget) {
    C.QtWidgetShow(w.ptr())
}

func Connect(obj Object, signal string, callback func()) func() {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var cb, del_cb = cgohelper.NewCallback(callback)
    var conn = C.QtConnect(obj.ptr(), new_str(signal), cgo_callback, C.size_t(cb))
    if int(C.QtIsConnectionValid(conn)) != 0 {
        return func() {
            C.QtDisconnect(conn)
            // use CommitTask() to prevent removing pending callbacks
            CommitTask(func() {
                del_cb()
            })
        }
    } else {
        panic("invalid connection")
    }
}

func Listen(obj Object, kind EventKind, prevent bool, callback func(Event)) func() {
    var l C.QtEventListener
    var cb, del_cb = cgohelper.NewCallback(func() {
        var ev = C.QtGetCurrentEvent(l)
        callback(Event(ev))
    })
    CommitTask(func() {
        var prevent_flag int
        if prevent { prevent_flag = 1 } else { prevent_flag = 0 }
        l = C.QtAddEventListener(obj.ptr(), C.size_t(kind), C.int(prevent_flag), cgo_callback, C.size_t(cb))
    })
    return func() {
        CommitTask(func() {
            C.QtRemoveEventListener(obj.ptr(), l)
            del_cb()
        })
    }
}

func GetPropQtString(obj Object, prop string) String {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    return String(C.QtObjectGetPropString(obj.ptr(), new_str(prop)))
}

func SetPropQtString(obj Object, prop string, val String) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropString(obj.ptr(), new_str(prop), C.QtString(val))
}

func GetPropRuneString(obj Object, prop string) ([] rune) {
    return StringToRunes(GetPropQtString(obj, prop))
}

func SetPropRuneString(obj Object, prop string, val ([] rune)) {
    var q_val, del_str = NewStringFromRunes(val)
    defer del_str()
    SetPropQtString(obj, prop, q_val)
}

func GetPropString(obj Object, prop string) string {
    return string(GetPropRuneString(obj, prop))
}

func SetPropString(obj Object, prop string, value string) {
    SetPropRuneString(obj, prop, ([] rune)(value))
}

func (ev Event) ResizeEventGetWidth() uint {
    return uint(C.QtResizeEventGetWidth(C.QtEvent(ev)))
}

func (ev Event) ResizeEventGetHeight() uint {
    return uint(C.QtResizeEventGetHeight(C.QtEvent(ev)))
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

