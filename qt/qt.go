package qt

/*
#include <stdlib.h>
#include "qtbinding/qtbinding.h"
*/
// #cgo LDFLAGS: -L./build -lqtbinding -Wl,-rpath=\$ORIGIN/
import "C"

import (
    "unsafe"
    "sync"
    "kumachan/qt/cgohelper"
    "fmt"
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
type Icon C.QtIcon
type Pixmap C.QtPixmap

type ListWidgetItem struct {
    Key    [] rune
    Label  [] rune
    Icon   *ImageData
}
type ImageData struct {
    Data    [] byte
    Format  ImageDataFormat
}
type ImageDataFormat int
const (
    PNG  ImageDataFormat  =  iota
    JPEG
)

type EventKind uint
func EventMove() EventKind { return EventKind(uint(C.QtEventMove)) }
func EventResize() EventKind { return EventKind(uint(C.QtEventResize)) }
func EventClose() EventKind { return EventKind(uint(C.QtEventClose)) }

type Event C.QtEvent


var initialized = false
var init_mutex sync.Mutex
var InitRequestSignal = make(chan func())
func MakeSureInitialized() {
    init_mutex.Lock()
    if !(initialized) {
        initialized = true
        init_mutex.Unlock()
        var wait = make(chan struct{})
        InitRequestSignal <- func() {
           C.QtInit()
           wait <- struct{}{}
           C.QtMain()
        }
        <- wait
    } else {
        init_mutex.Unlock()
    }
}

/// Invokes the operation callback in the Qt main thread.
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
    var channel = make(chan C.QtConnHandle)
    // Note: Although connect() is documented as "thread-safe",
    //       it is not clear what will happen if the goroutine
    //       is preempted and moved to another thread while
    //       calling connect(), thus CommitTask() is used here.
    CommitTask(func() {
        var conn = C.QtConnect(obj.ptr(), new_str(signal), cgo_callback, C.size_t(cb))
        channel <- conn
    })
    var conn = <- channel
    if int(C.QtIsConnectionValid(conn)) != 0 {
        return func() {
            CommitTask(func() {
                C.QtDisconnect(conn)
                // Note: Use CommitTask() to prevent pending callbacks
                //       from being removed.
                CommitTask(func() {
                    del_cb()
                })
            })
        }
    } else {
        panic("invalid connection")
    }
}

func BlockSignals(obj Object) (error, func()) {
    C.QtBlockSignals(obj.ptr(), 1)
    return nil, func() {
        C.QtBlockSignals(obj.ptr(), 0)
    }
}

func Listen(obj Object, kind EventKind, prevent bool, callback func(Event)) func() {
    var l C.QtEventListener
    var cb, del_cb = cgohelper.NewCallback(func() {
        var ev = C.QtGetCurrentEvent(l)
        callback(Event(ev))
    })
    var ok = make(chan struct{})
    CommitTask(func() {
        var prevent_flag int
        if prevent { prevent_flag = 1 } else { prevent_flag = 0 }
        l = C.QtAddEventListener(obj.ptr(), C.size_t(kind), C.int(prevent_flag), cgo_callback, C.size_t(cb))
        ok <- struct{} {}
    })
    <- ok
    return func() {
        CommitTask(func() {
            C.QtRemoveEventListener(obj.ptr(), l)
            del_cb()
        })
    }
}

func GetPropQtString(obj Object, prop string) (String, func()) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var value = C.QtObjectGetPropString(obj.ptr(), new_str(prop))
    return String(value), func() {
        C.QtDeleteString(value)
    }
}

func SetPropQtString(obj Object, prop string, val String) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropString(obj.ptr(), new_str(prop), C.QtString(val))
}

func GetPropRuneString(obj Object, prop string) ([] rune) {
    var value, del = GetPropQtString(obj, prop)
    defer del()
    return StringToRunes(value)
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

func NewPixmap(data ([] byte), format ImageDataFormat) (Pixmap, func()) {
    var buf = (*C.uint8_t)(unsafe.Pointer(&data[0]))
    var length = C.size_t(uint(len(data)))
    if format == PNG {
        var pm = C.QtNewPixmapPNG(buf, length)
        return Pixmap(pm), func() { C.QtDeletePixmap(pm) }
    } else if format == JPEG {
        var pm = C.QtNewPixmapJPEG(buf, length)
        return Pixmap(pm), func() { C.QtDeletePixmap(pm) }
    } else {
        panic("qt pixmap: unsupported image format")
    }
}

func NewIcon(pm Pixmap) (Icon, func()) {
    var icon = C.QtNewIcon(C.QtPixmap(pm))
    return Icon(icon), func() {
        C.QtDeleteIcon(icon)
    }
}

func ListWidgetClear(w Widget) {
    C.QtListWidgetClear(w.ptr())
}

func ListWidgetSetItems(w Widget, get_item (func(uint) ListWidgetItem), length uint, current ([] rune)) {
    var _, unblock = BlockSignals(w)
    ListWidgetClear(w)
    var icon_pool = make(map[*ImageData] struct { icon Icon; del func() })
    var occurred_keys = make(map[string] bool)
    for i := uint(0); i < length; i += 1 {
        var item = get_item(i)
        var key_str = string(item.Key)
        if occurred_keys[key_str] {
            panic(fmt.Sprintf("qt listwidget: duplicate item key %s", key_str))
        }
        occurred_keys[key_str] = true
        var key, del_key = NewStringFromRunes(item.Key)
        var label, del_label = NewStringFromRunes(item.Label)
        var is_current = (current != nil && key_str == string(current))
        var current_flag int
        if is_current { current_flag = 1 } else { current_flag = 0 }
        if item.Icon != nil {
            var icon Icon
            var cached, is_cached = icon_pool[item.Icon]
            if is_cached {
                icon = cached.icon
            } else {
                var pm, del_pm = NewPixmap(item.Icon.Data, item.Icon.Format)
                var new_icon, del_icon = NewIcon(pm)
                del_pm()
                icon = new_icon
                icon_pool[item.Icon] = struct { icon Icon; del func() } {
                    icon: new_icon,
                    del:  del_icon,
                }
            }
            C.QtListWidgetAddItemWithIcon(
                w.ptr(), C.QtString(key), C.QtIcon(icon), C.QtString(label), C.int(current_flag))
        } else {
            C.QtListWidgetAddItem(
                w.ptr(), C.QtString(key), C.QtString(label), C.int(current_flag))
        }
        del_label()
        del_key()
    }
    for _, cached := range icon_pool {
        cached.del()
    }
    unblock()
}

func ListWidgetHasCurrentItem(w Widget) bool {
    return (C.QtListWidgetHasCurrentItem(w.ptr()) != 0)
}

func ListWidgetGetCurrentItemKey(w Widget) ([] rune) {
    var raw_key = C.QtListWidgetGetCurrentItemKey(w.ptr())
    defer C.QtDeleteString(raw_key)
    return StringToRunes(String(raw_key))
}

func (ev Event) ResizeEventGetWidth() uint {
    return uint(C.QtResizeEventGetWidth(C.QtEvent(ev)))
}

func (ev Event) ResizeEventGetHeight() uint {
    return uint(C.QtResizeEventGetHeight(C.QtEvent(ev)))
}

func WebUiDebug(title String) {
    MakeSureInitialized()
    C.QtWebUiInit(C.QtString(title))
}
