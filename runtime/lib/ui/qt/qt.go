package qt

/*
#include <stdlib.h>
#include "qtbinding/qtbinding.h"
*/
// #cgo LDFLAGS: -L./build -lqtbinding -Wl,-rpath=\$ORIGIN/
import "C"

import (
    "fmt"
    "unsafe"
    "kumachan/runtime/lib/ui/qt/cgohelper"
    "reflect"
)


type Object interface {
    ptr() unsafe.Pointer
    QtObject()
}
func (obj object) QtObject()           {}
func (obj object) ptr() unsafe.Pointer { return obj.addr }
type object struct { addr unsafe.Pointer }
func ObjectNullablePointer(obj Object) unsafe.Pointer {
    if obj == nil {
        return nil
    } else {
        return obj.ptr()
    }
}

type Widget interface {
    Object
    QtWidget()
}
func (widget) QtWidget() {}
type widget struct { object }
func Show(w Widget) {
    C.QtWidgetShow(w.ptr())
}
func MoveToScreenCenter(w Widget) {
    C.QtWidgetMoveToScreenCenter(w.ptr())
}
func DialogExec(w Widget) {
    C.QtDialogExec(w.ptr())
}
func DialogAccept(w Widget) {
    C.QtDialogAccept(w.ptr())
}
func DialogReject(w Widget) {
    C.QtDialogReject(w.ptr())
}

type Action interface {
    Object
    QtAction()
}
func (action) QtAction() {}
type action struct { object }

type Ucs4String = [] rune
type String C.QtString
type Bool C.int
type VariantMap C.QtVariantMap
type Icon C.QtIcon
type Pixmap C.QtPixmap

type ListWidgetItem struct {
    Key    Ucs4String
    Label  Ucs4String
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

type Event C.QtEvent
type EventKind uint
func EventMove() EventKind { return EventKind(uint(C.QtEventMove)) }
func EventResize() EventKind { return EventKind(uint(C.QtEventResize)) }
func EventClose() EventKind { return EventKind(uint(C.QtEventClose)) }
func (ev Event) ResizeEventGetWidth() uint {
    return uint(C.QtResizeEventGetWidth(C.QtEvent(ev)))
}
func (ev Event) ResizeEventGetHeight() uint {
    return uint(C.QtResizeEventGetHeight(C.QtEvent(ev)))
}

var mock = false
var initializing = make(chan struct{}, 1)
var initialized = make(chan struct{})
var initRequestSignal = make(chan func())
func InitRequestSignal() (chan func()) { return initRequestSignal }
func Mock() {
    mock = true
}
func MakeSureInitialized() {
    if mock {
        return
    }
    select {
    case initializing <- struct{}{}:
        var wait = make(chan struct{})
        initRequestSignal <- func() {
            C.QtInit()
            wait <- struct{}{}
            C.QtMain()
        }
        <- wait
        close(initialized)
    default:
        <- initialized
    }
}

// Invokes the operation callback in the main thread of Qt.
func CommitTask(operation func()) {
    if mock {
        go operation()
        return
    }
    var delete_callback (func() bool)
    var f = func() {
        operation()
        delete_callback()
    }
    callback, delete_callback := cgohelper.NewCallback(f)
    C.QtCommitTask(cgo_callback, C.size_t(callback))
}

func LoadWidget(def string, dir string) (Widget, bool) {
    if mock {
        return widget{}, true
    }
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
    if mock {
        return widget{}, true
    }
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtWidgetFindChild(w.ptr(), new_str(name))
    if ptr != nil {
        return widget{object{ptr}}, true
    } else {
        return widget{}, false
    }
}

func FindChildAction(w Widget, name string) (Action, bool) {
    if mock {
        return action{}, true
    }
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var ptr = C.QtWidgetFindChildAction(w.ptr(), new_str(name))
    if ptr != nil {
        return action{object{ptr}}, true
    } else {
        return action{}, false
    }
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
        var prevent_ = MakeBool(prevent)
        l = C.QtAddEventListener(obj.ptr(), C.size_t(kind), C.int(prevent_), cgo_callback, C.size_t(cb))
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

func GetPropBool(obj Object, prop string) bool {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var val = C.QtObjectGetPropBool(obj.ptr(), new_str(prop))
    return (val != 0)
}
func SetPropBool(obj Object, prop string, val bool) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var int_val int
    if val {
        int_val = 1
    } else {
        int_val = 0
    }
    C.QtObjectSetPropBool(obj.ptr(), new_str(prop), C.int(int_val))
}
func GetPropQtString(obj Object, prop string) String {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var value = C.QtObjectGetPropString(obj.ptr(), new_str(prop))
    return String(value)
}
func SetPropQtString(obj Object, prop string, val String) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropString(obj.ptr(), new_str(prop), C.QtString(val))
}
func GetPropRuneString(obj Object, prop string) ([] rune) {
    var value = GetPropQtString(obj, prop)
    var value_runes = StringToRunes(value)
    return value_runes
}
func SetPropRuneString(obj Object, prop string, val ([] rune)) {
    var q_val, del_str = NewString(val)
    defer del_str()
    SetPropQtString(obj, prop, q_val)
}
func GetPropString(obj Object, prop string) string {
    return string(GetPropRuneString(obj, prop))
}
func SetPropString(obj Object, prop string, value string) {
    SetPropRuneString(obj, prop, ([] rune)(value))
}
func GetPropInt(obj Object, prop string) int {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    return int(C.QtObjectGetPropInt(obj.ptr(), new_str(prop)))
}
func SetPropInt(obj Object, prop string, val int) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropInt(obj.ptr(), new_str(prop), C.int(val))
}

func MakeBool(p bool) Bool {
    if p { return Bool(C.int(int(1))) } else { return Bool(C.int(int(0))) }
}
func NewString(runes Ucs4String) (String, func()) {
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
func NewStringFromUtf8Binary(buf ([] byte)) (String, func()) {
    var str C.QtString
    if len(buf) > 0 {
        var ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
        var size = (C.size_t)(len(buf))
        str = C.QtNewStringUTF8(ptr, size)
    } else {
        str = C.QtNewStringUTF8(nil, 0)
    }
    return String(str), func() {
        C.QtDeleteString(str)
    }
}
func NewStringFromGoString(go_str string) (String, func()) {
    var hdr = *(*reflect.StringHeader)(unsafe.Pointer(&go_str))
    var bin = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader {
        Data: hdr.Data,
        Len:  hdr.Len,
        Cap:  hdr.Len,
    }))
    return NewStringFromUtf8Binary(bin)
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
    C.QtDeleteString(q_str)
    return buf
}
func DeleteString(str String) {
    C.QtDeleteString((C.QtString)(str))
}

func VariantMapGetRunes(m VariantMap, key String) ([] rune) {
    var val = C.QtVariantMapGetString(C.QtVariantMap(m), C.QtString(key))
    var val_runes = StringToRunes(String(val))
    return val_runes
}
func VariantMapGetFloat(m VariantMap, key String) float64 {
    var val = C.QtVariantMapGetFloat(C.QtVariantMap(m), C.QtString(key))
    return float64(val)
}
func VariantMapGetBool(m VariantMap, key String) bool {
    var val = C.QtVariantMapGetBool(C.QtVariantMap(m), C.QtString(key))
    return (val != 0)
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
        var key, del_key = NewString(item.Key)
        var label, del_label = NewString(item.Label)
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
    var key = StringToRunes(String(raw_key))
    return key
}

func WebViewDisableContextMenu(w Widget) {
    C.QtWebViewDisableContextMenu(w.ptr())
}
func WebViewEnableLinkDelegation(w Widget) {
    C.QtWebViewEnableLinkDelegation(w.ptr())
}
func WebViewRecordClickedLink(w Widget) {
    C.QtWebViewRecordClickedLink(w.ptr())
}
func WebViewSetHTML(w Widget, html String) {
    C.QtWebViewSetHTML(w.ptr(), C.QtString(html))
}
func WebViewScrollToAnchor(w Widget, anchor String) {
    C.QtWebViewScrollToAnchor(w.ptr(), C.QtString(anchor))
}
func WebViewScrollToTop(w Widget) {
    C.QtWebViewScrollToTop(w.ptr())
}

type FileDialogOptions struct {
    Title   [] rune
    Cwd     [] rune
    Filter  [] rune
}
func fileDialogAdaptOptions(opts FileDialogOptions) (String, String, String, func()) {
    var title, del_title = NewString(opts.Title)
    var cwd, del_cwd = NewString(opts.Cwd)
    var filter, del_filter = NewString(opts.Filter)
    return title, cwd, filter, func() {
        del_title()
        del_cwd()
        del_filter()
    }
}
func FileDialogOpen(parent Widget, opts FileDialogOptions) ([] rune) {
    var parent_ptr = ObjectNullablePointer(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogOpen(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    return StringToRunes(String(raw_path))
}
func FileDialogOpenMultiple(parent Widget, opts FileDialogOptions) ([][] rune) {
    var parent_ptr = ObjectNullablePointer(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path_list = C.QtFileDialogOpenMultiple(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    var path_list = make([][] rune, 0)
    var L = uint(C.QtStringListGetSize(raw_path_list))
    for i := uint(0); i < L; i += 1 {
        var raw_item = C.QtStringListGetItem(raw_path_list, C.size_t(i))
        var item = StringToRunes(String(raw_item))
        path_list = append(path_list, item)
    }
    C.QtDeleteStringList(raw_path_list)
    return path_list
}
func FileDialogSelectDirectory(parent Widget, opts FileDialogOptions) ([] rune) {
    var parent_ptr = ObjectNullablePointer(parent)
    var title, cwd, _, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogSelectDirectory(parent_ptr,
        C.QtString(title), C.QtString(cwd))
    return StringToRunes(String(raw_path))
}
func FileDialogSave(parent Widget, opts FileDialogOptions) ([] rune) {
    var parent_ptr = ObjectNullablePointer(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogSave(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    return StringToRunes(String(raw_path))
}

var webui_initialized = false
func WebUiInit(title String, debug bool) {
    MakeSureInitialized()
    C.WebUiInit(C.QtString(title), C.int(MakeBool(debug)))
    webui_initialized = true
}
func WebUiLoadView() {
    if !webui_initialized { panic("webui not initialized") }
    C.WebUiLoadView()
}
func WebUiGetWindow() Widget {
    return widget { object { C.WebUiGetWindow() } }
}
func WebUiRegisterAsset(path String, mime String, data ([] byte))  {
    var buf = (*C.uint8_t)(unsafe.Pointer(&data[0]))
    var length = C.size_t(uint(len(data)))
    C.WebUiRegisterAsset(C.QtString(path), C.QtString(mime), buf, length)
}
func WebUiInjectCSS(path String) String {
    return String(C.WebUiInjectCSS(C.QtString(path)))
}
func WebUiInjectJS(path String) String {
    return String(C.WebUiInjectJS(C.QtString(path)))
}
func WebUiInjectTTF(path String, family String, weight String, style String) String {
    return String(C.WebUiInjectTTF(C.QtString(path), C.QtString(family), C.QtString(weight), C.QtString(style)))
}
func WebUiPatchActualDOM(patch_data ([] byte)) {
    var str, del = NewStringFromUtf8Binary(patch_data)
    defer del()
    C.WebUiPatchActualDOM(C.QtString(str))
}

type WebUiEventPayload struct {
    Data  VariantMap
}
func WebUiGetCurrentEventHandler() Ucs4String {
    var raw_id = C.WebUiGetCurrentEventHandler()
    var id_str = StringToRunes(String(raw_id))
    return id_str
}
func WebUiGetCurrentEventPayload() *WebUiEventPayload {
    return &WebUiEventPayload { VariantMap(C.WebUiGetCurrentEventPayload()) }
}
func WebUiConsumeEventPayload(ev *WebUiEventPayload, f func(*WebUiEventPayload) interface{}) interface{} {
    defer func() {
        C.QtDeleteVariantMap(C.QtVariantMap(ev.Data))
    } ()
    return f(ev)
}
func WebUiEventPayloadGetRunes(ev *WebUiEventPayload, key ([] rune)) ([] rune) {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetRunes(ev.Data, key_str)
}
func WebUiEventPayloadGetFloat(ev *WebUiEventPayload, key ([] rune)) float64 {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetFloat(ev.Data, key_str)
}
func WebUiEventPayloadGetBool(ev *WebUiEventPayload, key ([] rune)) bool {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetBool(ev.Data, key_str)
}

