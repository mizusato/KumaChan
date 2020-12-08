#ifndef LIB_H
#define LIB_H

#include <stdlib.h>
#include <stdint.h>

struct _QtConnHandle {
    void* ptr;
};
typedef struct _QtConnHandle QtConnHandle;

typedef int QtBool;

struct _QtString {
    void* ptr;
};
typedef struct _QtString QtString;

struct _QtStringList {
    void* ptr;
};
typedef struct _QtStringList QtStringList;

struct _QtVariantMap {
    void* ptr;
};
typedef struct _QtVariantMap QtVariantMap;

struct _QtVariantList {
    void* ptr;
};
typedef struct _QtVariantList QtVariantList;

struct _QtIcon {
    void* ptr;
};
typedef struct _QtIcon QtIcon;

struct _QtPixmap {
    void* ptr;
};
typedef struct _QtPixmap QtPixmap;

struct _QtEvent {
    void* ptr;
};
typedef struct _QtEvent QtEvent;

struct _QtEventListener {
    void* ptr;
};
typedef struct _QtEventListener QtEventListener;

extern const size_t QtEventMove;
extern const size_t QtEventResize;
extern const size_t QtEventClose;

#ifdef __cplusplus
extern "C" {
#endif
    void QtInit();
    int QtMain();
    void QtCommitTask(void (*cb)(size_t), size_t payload);
    void QtExit(int code);
    void QtQuit();
    void* QtLoadWidget(const char* definition, const char* directory);
    void* QtWidgetFindChild(void* widget_ptr, const char* name);
    void QtWidgetShow(void* widget_ptr);
    void QtWidgetHide(void* widget_ptr);
    QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val);
    QtString QtObjectGetPropString(void* obj_ptr, const char* prop);
    QtConnHandle QtConnect(void* obj_ptr, const char* signal, void (*cb)(size_t), size_t payload);
    QtBool QtIsConnectionValid(QtConnHandle handle);
    void QtDisconnect(QtConnHandle handle);
    void QtBlockSignals(void* obj_ptr, QtBool block);
    QtEventListener QtAddEventListener(void* obj_ptr, size_t kind, QtBool prevent, void (*cb)(size_t), size_t payload);
    QtEvent QtGetCurrentEvent(QtEventListener listener);
    void QtRemoveEventListener(void* obj_ptr, QtEventListener listener);
    size_t QtResizeEventGetWidth(QtEvent ev);
    size_t QtResizeEventGetHeight(QtEvent ev);
    QtString QtNewStringUTF8(const char* buf, size_t len);
    QtString QtNewStringUTF32(const uint32_t* buf, size_t len);
    void QtDeleteString(QtString str);
    size_t QtStringListGetSize(QtStringList list);
    QtString QtStringListGetItem(QtStringList list, size_t index);
    void QtDeleteStringList(QtStringList list);
    QtVariantList QtNewVariantList();
    void QtVariantListAppendNumber(QtVariantList l, double n);
    void QtVariantListAppendString(QtVariantList l, QtString str);
    void QtDeleteVariantList(QtVariantList l);
    QtString QtVariantMapGetString(QtVariantMap m, QtString key);
    double QtVariantMapGetNumber(QtVariantMap m, QtString key);
    QtBool QtVariantMapGetBool(QtVariantMap m, QtString key);
    void QtDeleteVariantMap(QtVariantMap m);
    size_t QtStringUTF16Length(QtString str);
    size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf);
    QtIcon QtNewIcon(QtPixmap pm);
    void QtDeleteIcon(QtIcon icon);
    QtPixmap QtNewPixmapPNG(const uint8_t* buf, size_t len);
    QtPixmap QtNewPixmapJPEG(const uint8_t* buf, size_t len);
    void QtDeletePixmap(QtPixmap pm);
    void QtListWidgetClear(void *widget_ptr);
    void QtListWidgetAddItem(void* widget_ptr, QtString key_, QtString label_, QtBool as_current);
    void QtListWidgetAddItemWithIcon(void* widget_ptr, QtString key_, QtIcon icon_, QtString label_, QtBool as_current);
    QtBool QtListWidgetHasCurrentItem(void* widget_ptr);
    QtString QtListWidgetGetCurrentItemKey(void *widget_ptr);
    QtString QtFileDialogOpen(void* parent_ptr, QtString title, QtString cwd, QtString filter);
    QtStringList QtFileDialogOpenMultiple(void* parent_ptr,  QtString title, QtString cwd, QtString filter);
    QtString QtFileDialogSelectDirectory(void *parent_ptr, QtString title, QtString cwd);
    QtString QtFileDialogSave(void *parent_ptr, QtString title, QtString cwd, QtString filter);
    void WebUiInit(QtString title, QtString css);
    void WebUiLoadView();
    void* WebUiGetWindow();
    QtString WebUiGetEventHandler();
    QtVariantMap WebUiGetEventPayload();
    void WebUiCallMethod(QtString id, QtString name, QtVariantList args);
    void WebUiEraseStyle(QtString id, QtString key);
    void WebUiApplyStyle(QtString id, QtString key, QtString value);
    void WebUiRemoveAttr(QtString id, QtString name);
    void WebUiSetAttr(QtString id, QtString name, QtString value);
    void WebUiDetachEvent(QtString id, QtString event);
    void WebUiModifyEvent(QtString id, QtString event, QtBool prevent, QtBool stop);
    void WebUiAttachEvent(QtString id, QtString event, QtBool prevent, QtBool stop, QtString handler);
    void WebUiSetText(QtString id, QtString text);
    void WebUiInsertNode(QtString parent, QtString ref, QtString id, QtString tag);
    void WebUiAppendNode(QtString parent, QtString id, QtString tag);
    void WebUiRemoveNode(QtString parent, QtString id);
    void WebUiUpdateNode(QtString old_id, QtString new_id);
    void WebUiReplaceNode(QtString parent, QtString old_id, QtString id, QtString tag);
#ifdef __cplusplus
}
#endif

#endif
