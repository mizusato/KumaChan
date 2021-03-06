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
    "kumachan/standalone/qt/cgohelper"
    "reflect"
)


type Object interface {
    ptr() unsafe.Pointer
    QtObject()
}
func (obj object) QtObject() {}
func (obj object) ptr() unsafe.Pointer { return obj.addr }
type object struct { addr unsafe.Pointer }

// TODO: subclasses
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
func ParentNullable(widget Widget) unsafe.Pointer {
    if widget == nil {
        return nil
    } else {
        return widget.ptr()
    }
}

type Action interface {
    Object
    QtAction()
}
func (action) QtAction() {}
type action struct { object }

type String C.QtString
type Bool C.int
type VariantMap C.QtVariantMap
type Icon C.QtIcon
type Pixmap C.QtPixmap
type Point struct {
    X  int
    Y  int
}
func getBool(number C.int) bool { return (number != 0) }

type ListWidgetItem struct {
    Key    string
    Label  string
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
func EventDynamicPropertyChange() EventKind { return EventKind(uint(C.QtEventDynamicPropertyChange)) }
func (ev Event) ResizeEventGetWidth() uint {
    return uint(C.QtResizeEventGetWidth(C.QtEvent(ev)))
}
func (ev Event) ResizeEventGetHeight() uint {
    return uint(C.QtResizeEventGetHeight(C.QtEvent(ev)))
}
func (ev Event) DynamicPropertyChangeEventGetPropertyName() string {
    return ConsumeString(String(C.QtDynamicPropertyChangeEventGetPropertyName(C.QtEvent(ev))))
}

var debugEnabled = false
func EnableDebug() {
    debugEnabled = true
}
var initializing = make(chan struct{}, 1)
var initialized = make(chan struct{})
var initRequestSignal = make(chan func())
// Calling this function will notify that Qt is not used in the entire program
// so that Main() can return immediately
func NotifyNotUsed() {
    close(initRequestSignal)
}
// NOTE: should be called on main thread to make QtWebkit work normally
func Main() {
    var main, use_qt = <- initRequestSignal
    if use_qt {
        main()
    }
}
func Quit(after func()) {
    select {
    case <- initialized:
        var wait = make(chan struct{})
        CommitTask(func() {
            C.QtQuit()
            after()
            wait <- struct{}{}
        })
        <- wait
    default:
        after()
    }
}
// TODO: rename this function
func MakeSureInitialized() {
    select {
    case initializing <- struct{}{}:
        var wait = make(chan struct{})
        initRequestSignal <- func() {
            C.QtInit(MakeBool(debugEnabled))
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

func FindChildAction(w Widget, name string) (Action, bool) {
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
    if getBool(C.QtIsConnectionValid(conn)) {
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

func BlockCallbacks(obj Object) (error, func()) {
    C.QtBlockCallbacks(obj.ptr(), 1)
    return nil, func() {
        C.QtBlockCallbacks(obj.ptr(), 0)
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
        var prevent_flag = MakeBool(prevent)
        l = C.QtAddEventListener(obj.ptr(), C.size_t(kind), prevent_flag, cgo_callback, C.size_t(cb))
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
    return getBool(val)
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
func getPropQtString(obj Object, prop string) String {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    var value = C.QtObjectGetPropString(obj.ptr(), new_str(prop))
    return String(value)
}
func setPropQtString(obj Object, prop string, val String) {
    var new_str, del_all_str = str_alloc()
    defer del_all_str()
    C.QtObjectSetPropString(obj.ptr(), new_str(prop), C.QtString(val))
}
func GetPropString(obj Object, prop string) string {
    var value = getPropQtString(obj, prop)
    var value_runes = ConsumeString(value)
    return value_runes
}
func SetPropString(obj Object, prop string, val string) {
    var q_val, del_str = NewString(val)
    defer del_str()
    setPropQtString(obj, prop, q_val)
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

func MakeBool(p bool) C.int {
    if p { return C.int(int(1)) } else { return C.int(int(0)) }
}
func NewString(data string) (String, func()) {
    var str C.QtString
    if len(data) > 0 {
        var hdr = (*reflect.StringHeader)(unsafe.Pointer(&data))
        var ptr = (*C.uint8_t)(unsafe.Pointer(hdr.Data))
        var size = (C.size_t)(len(data))
        str = C.QtNewStringUTF8(ptr, size)
    } else {
        str = C.QtNewStringUTF8(nil, 0)
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
func ConsumeString(str String) string {
    var q_str = (C.QtString)(str)
    var size16 = uint(C.QtStringUTF16Length(q_str))
    var buf = make([] rune, size16)
    if size16 > 0 {
        var size32 = uint(C.QtStringWriteToUTF32Buffer(q_str,
            (*C.uint32_t)(unsafe.Pointer(&buf[0]))))
        buf = buf[:size32]
    }
    C.QtDeleteString(q_str)
    return string(buf)
}
func DeleteString(str String) {
    C.QtDeleteString((C.QtString)(str))
}

func makeQtPoint(p Point) C.QtPoint {
    return C.QtMakePoint(C.int(p.X), C.int(p.Y))
}
func makePoint(p C.QtPoint) Point {
    return Point { X: int(C.QtPointGetX(p)), Y: int(C.QtPointGetY(p)) }
}

func VariantMapGetString(m VariantMap, key String) string {
    var val = C.QtVariantMapGetString(C.QtVariantMap(m), C.QtString(key))
    var val_runes = ConsumeString(String(val))
    return val_runes
}
func VariantMapGetFloat(m VariantMap, key String) float64 {
    var val = C.QtVariantMapGetFloat(C.QtVariantMap(m), C.QtString(key))
    return float64(val)
}
func VariantMapGetBool(m VariantMap, key String) bool {
    var val = C.QtVariantMapGetBool(C.QtVariantMap(m), C.QtString(key))
    return getBool(val)
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
func NewIconEmpty() (Icon, func()) {
    var icon = C.QtNewIconEmpty()
    return Icon(icon), func() {
        C.QtDeleteIcon(icon)
    }
}

func ListWidgetClear(w Widget) {
    C.QtListWidgetClear(w.ptr())
}
func ListWidgetSetItems(w Widget, get_item (func(uint) ListWidgetItem), length uint, current ([] rune)) {
    // note: block signals instead of callbacks because
    //       this function is also used from the interpreter side,
    //       e.g. api browser
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
    return getBool(C.QtListWidgetHasCurrentItem(w.ptr()))
}
func ListWidgetGetCurrentItemKey(w Widget) string {
    var raw_key = C.QtListWidgetGetCurrentItemKey(w.ptr())
    var key = ConsumeString(String(raw_key))
    return key
}

func BaseWebViewDisableContextMenu(w Widget) {
    C.QtWebViewDisableContextMenu(w.ptr())
}
func BaseWebViewEnableLinkDelegation(w Widget) {
    C.QtWebViewEnableLinkDelegation(w.ptr())
}
func BaseWebViewSetHTML(w Widget, html String, base_url String) {
    C.QtWebViewSetHTML(w.ptr(), C.QtString(html), C.QtString(base_url))
}
func BaseWebViewScrollToAnchor(w Widget, anchor String) {
    C.QtWebViewScrollToAnchor(w.ptr(), C.QtString(anchor))
}
func BaseWebViewGetScroll(w Widget) Point {
    return makePoint(C.QtWebViewGetScroll(w.ptr()))
}
func BaseWebViewSetScroll(w Widget, pos Point) {
    C.QtWebViewSetScroll(w.ptr(), makeQtPoint(pos));
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
func DialogShowModal(w Widget) {
    C.QtDialogShowModal(w.ptr())
}

type FileDialogOptions struct {
    Title   string
    Cwd     string
    Filter  string
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
func FileDialogOpen(parent Widget, opts FileDialogOptions) string {
    var parent_ptr = ParentNullable(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogOpen(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    return ConsumeString(String(raw_path))
}
func FileDialogOpenMultiple(parent Widget, opts FileDialogOptions) ([] string) {
    var parent_ptr = ParentNullable(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path_list = C.QtFileDialogOpenMultiple(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    var path_list = make([] string, 0)
    var L = uint(C.QtStringListGetSize(raw_path_list))
    for i := uint(0); i < L; i += 1 {
        var raw_item = C.QtStringListGetItem(raw_path_list, C.size_t(i))
        var item = ConsumeString(String(raw_item))
        path_list = append(path_list, item)
    }
    C.QtDeleteStringList(raw_path_list)
    return path_list
}
func FileDialogSelectDirectory(parent Widget, opts FileDialogOptions) string {
    var parent_ptr = ParentNullable(parent)
    var title, cwd, _, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogSelectDirectory(parent_ptr,
        C.QtString(title), C.QtString(cwd))
    return ConsumeString(String(raw_path))
}
func FileDialogSave(parent Widget, opts FileDialogOptions) string {
    var parent_ptr = ParentNullable(parent)
    var title, cwd, filter, del = fileDialogAdaptOptions(opts)
    defer del()
    var raw_path = C.QtFileDialogSave(parent_ptr,
        C.QtString(title), C.QtString(cwd), C.QtString(filter))
    return ConsumeString(String(raw_path))
}

func WebViewLoadContent(view Widget) {
    C.WebViewLoadContent(view.ptr())
}
func WebViewIsContentLoaded(view Widget) bool {
    return getBool(C.WebViewIsContentLoaded(view.ptr()))
}
func WebViewRegisterAsset(view Widget, path String, mime String, data ([] byte))  {
    var buf = (*C.uint8_t)(unsafe.Pointer(&data[0]))
    var length = C.size_t(uint(len(data)))
    C.WebViewRegisterAsset(view.ptr(), C.QtString(path), C.QtString(mime), buf, length)
}
func WebViewInjectCSS(view Widget, path String) String {
    return String(C.WebViewInjectCSS(view.ptr(), C.QtString(path)))
}
func WebViewInjectJS(view Widget, path String) String {
    return String(C.WebViewInjectJS(view.ptr(), C.QtString(path)))
}
func WebViewInjectTTF(view Widget, path String, family String, weight String, style String) String {
    return String(C.WebViewInjectTTF(view.ptr(), C.QtString(path), C.QtString(family), C.QtString(weight), C.QtString(style)))
}
func WebViewPatchActualDOM(view Widget, patch_data ([] byte)) {
    var str, del = NewStringFromUtf8Binary(patch_data)
    defer del()
    C.WebViewPatchActualDOM(view.ptr(), C.QtString(str))
}

type WebViewEventPayload struct {
    Data  VariantMap
}
func WebViewGetCurrentEventHandler(view Widget) string {
    var raw_id = C.WebViewGetCurrentEventHandler(view.ptr())
    var id_str = ConsumeString(String(raw_id))
    return id_str
}
func WebViewGetCurrentEventPayload(view Widget) *WebViewEventPayload {
    return &WebViewEventPayload { VariantMap(C.WebViewGetCurrentEventPayload(view.ptr())) }
}
func WebViewConsumeEventPayload(ev *WebViewEventPayload, f func(*WebViewEventPayload) interface{}) interface{} {
    defer func() {
        C.QtDeleteVariantMap(C.QtVariantMap(ev.Data))
    } ()
    return f(ev)
}
func WebViewEventPayloadGetString(ev *WebViewEventPayload, key string) string {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetString(ev.Data, key_str)
}
func WebViewEventPayloadGetFloat(ev *WebViewEventPayload, key string) float64 {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetFloat(ev.Data, key_str)
}
func WebViewEventPayloadGetBool(ev *WebViewEventPayload, key string) bool {
    var key_str, del = NewString(key)
    defer del()
    return VariantMapGetBool(ev.Data, key_str)
}

