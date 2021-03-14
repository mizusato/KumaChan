#ifndef QTBINDING_H
#define QTBINDING_H

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

struct _QtPoint {
    int x;
    int y;
};
typedef struct _QtPoint QtPoint;

extern const size_t QtEventMove;
extern const size_t QtEventResize;
extern const size_t QtEventClose;

#ifdef __cplusplus
extern "C" {
#endif
    void QtInit(QtBool debug);
    int QtMain();
    void QtCommitTask(void (*cb)(size_t), size_t payload);
    void QtExit(int code);
    void QtQuit();
    void* QtLoadWidget(const char* definition, const char* directory);
    void* QtWidgetFindChild(void* widget_ptr, const char* name);
    void* QtWidgetFindChildAction(void* widget_ptr, const char* name);
    void QtWidgetShow(void* widget_ptr);
    void QtWidgetHide(void* widget_ptr);
    void QtWidgetMoveToScreenCenter(void* widget_ptr);
    QtBool QtObjectSetPropBool(void* obj_ptr, const char* prop, QtBool val);
    QtBool QtObjectGetPropBool(void* obj_ptr, const char* prop);
    QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val);
    QtString QtObjectGetPropString(void* obj_ptr, const char* prop);
    QtBool QtObjectSetPropInt(void* obj_ptr, const char* prop, int val);
    int QtObjectGetPropInt(void* obj_ptr, const char* prop);
    QtBool QtObjectSetPropPoint(void* obj_ptr, const char* prop, QtPoint val);
    QtPoint QtObjectGetPropPoint(void* obj_ptr, const char* prop);
    QtConnHandle QtConnect(void* obj_ptr, const char* signal, void (*cb)(size_t), size_t payload);
    QtBool QtIsConnectionValid(QtConnHandle handle);
    void QtDisconnect(QtConnHandle handle);
    void QtBlockSignals(void* obj_ptr, QtBool block);
    void QtBlockCallbacks(void* obj_ptr, QtBool block);
    QtEventListener QtAddEventListener(void* obj_ptr, size_t kind, QtBool prevent, void (*cb)(size_t), size_t payload);
    QtEvent QtGetCurrentEvent(QtEventListener listener);
    void QtRemoveEventListener(void* obj_ptr, QtEventListener listener);
    size_t QtResizeEventGetWidth(QtEvent ev);
    size_t QtResizeEventGetHeight(QtEvent ev);
    QtPoint QtMakePoint(int x, int y);
    int QtPointGetX(QtPoint p);
    int QtPointGetY(QtPoint p);
    QtString QtNewStringUTF8(const uint8_t* buf, size_t len);
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
    double QtVariantMapGetFloat(QtVariantMap m, QtString key);
    QtBool QtVariantMapGetBool(QtVariantMap m, QtString key);
    void QtDeleteVariantMap(QtVariantMap m);
    size_t QtStringUTF16Length(QtString str);
    size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf);
    QtIcon QtNewIcon(QtPixmap pm);
    QtIcon QtNewIconEmpty();
    void QtDeleteIcon(QtIcon icon);
    QtPixmap QtNewPixmapPNG(const uint8_t* buf, size_t len);
    QtPixmap QtNewPixmapJPEG(const uint8_t* buf, size_t len);
    void QtDeletePixmap(QtPixmap pm);
    void QtListWidgetClear(void *widget_ptr);
    void QtListWidgetAddItem(void* widget_ptr, QtString key_, QtString label_, QtBool as_current);
    void QtListWidgetAddItemWithIcon(void* widget_ptr, QtString key_, QtIcon icon_, QtString label_, QtBool as_current);
    QtBool QtListWidgetHasCurrentItem(void* widget_ptr);
    QtString QtListWidgetGetCurrentItemKey(void *widget_ptr);
    void QtWebViewDisableContextMenu(void* widget_ptr);
    void QtWebViewEnableLinkDelegation(void *widget_ptr);
    void QtWebViewRecordClickedLink(void* widget_ptr);
    void QtWebViewSetHTML(void *widget_ptr, QtString html);
    void QtWebViewScrollToAnchor(void *widget_ptr, QtString anchor);
    QtPoint QtWebViewGetScroll(void* widget_ptr);
    void QtDialogExec(void *dialog_ptr);
    void QtDialogAccept(void *dialog_ptr);
    void QtDialogReject(void *dialog_ptr);
    void QtWebViewSetScroll(void* widget_ptr, QtPoint pos);
    void QtDialogShowModal(void* dialog_ptr);
    void QtDialogSetParent(void* dialog_ptr, void* parent_ptr);
    QtString QtFileDialogOpen(void* parent_ptr, QtString title, QtString cwd, QtString filter);
    QtStringList QtFileDialogOpenMultiple(void* parent_ptr,  QtString title, QtString cwd, QtString filter);
    QtString QtFileDialogSelectDirectory(void *parent_ptr, QtString title, QtString cwd);
    QtString QtFileDialogSave(void* parent_ptr, QtString title, QtString cwd, QtString filter);
    // Web
    void WebViewLoadContent(void* view_ptr);
    void WebViewRegisterAsset(void* view_ptr, QtString path, QtString mime, const uint8_t* buf, size_t len);
    QtString WebViewInjectCSS(void* view_ptr, QtString path);
    QtString WebViewInjectJS(void* view_ptr, QtString path);
    QtString WebViewInjectTTF(void* view_ptr, QtString path, QtString family, QtString weight, QtString style);
    void WebViewCallMethod(void* view_ptr, QtString id, QtString name, QtVariantList args);
    QtString WebViewGetCurrentEventHandler(void* view_ptr);
    QtVariantMap WebViewGetCurrentEventPayload(void* view_ptr);
    void WebViewPatchActualDOM(void* view_ptr, QtString operations);
    void* WebDialogCreate(void* parent_ptr, QtIcon icon, QtString title, int width, int height, QtBool closable);
    void* WebDialogGetWebView(void* dialog_ptr);
    void WebDialogDispose(void* dialog_ptr);
#ifdef __cplusplus
}
#endif

#endif

