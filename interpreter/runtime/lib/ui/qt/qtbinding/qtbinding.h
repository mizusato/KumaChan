#ifndef QTBINDING_H
#define QTBINDING_H

#include <stdlib.h>
#include <stdint.h>


#ifdef _WIN32
	#ifdef QTBINDING_WIN32_DLL
		#define EXPORT __declspec(dllexport)
	#else
		#define EXPORT __declspec(dllimport)
	#endif
#else
	#define EXPORT
#endif

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


#ifdef __cplusplus
extern "C" {
#endif
	// Event Categories
	EXPORT extern const size_t QtEventMove;
	EXPORT extern const size_t QtEventResize;
	EXPORT extern const size_t QtEventClose;
	EXPORT extern const size_t QtEventDynamicPropertyChange;
	// API
    EXPORT void QtInit(QtBool debug);
    EXPORT int QtMain();
    EXPORT void QtCommitTask(void (*cb)(size_t), size_t payload);
    EXPORT void QtExit(int code);
    EXPORT void QtQuit();
    EXPORT void* QtLoadWidget(const char* definition, const char* directory);
    EXPORT void* QtWidgetFindChild(void* widget_ptr, const char* name);
    EXPORT void* QtWidgetFindChildAction(void* widget_ptr, const char* name);
    EXPORT void QtWidgetShow(void* widget_ptr);
    EXPORT void QtWidgetHide(void* widget_ptr);
    EXPORT void QtWidgetMoveToScreenCenter(void* widget_ptr);
    EXPORT QtBool QtObjectSetPropBool(void* obj_ptr, const char* prop, QtBool val);
    EXPORT QtBool QtObjectGetPropBool(void* obj_ptr, const char* prop);
    EXPORT QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val);
    EXPORT QtString QtObjectGetPropString(void* obj_ptr, const char* prop);
    EXPORT QtBool QtObjectSetPropInt(void* obj_ptr, const char* prop, int val);
    EXPORT int QtObjectGetPropInt(void* obj_ptr, const char* prop);
    EXPORT QtBool QtObjectSetPropPoint(void* obj_ptr, const char* prop, QtPoint val);
    EXPORT QtPoint QtObjectGetPropPoint(void* obj_ptr, const char* prop);
    EXPORT QtConnHandle QtConnect(void* obj_ptr, const char* signal, void (*cb)(size_t), size_t payload);
    EXPORT QtBool QtIsConnectionValid(QtConnHandle handle);
    EXPORT void QtDisconnect(QtConnHandle handle);
    EXPORT void QtBlockSignals(void* obj_ptr, QtBool block);
    EXPORT void QtBlockCallbacks(void* obj_ptr, QtBool block);
    EXPORT QtEventListener QtAddEventListener(void* obj_ptr, size_t kind, QtBool prevent, void (*cb)(size_t), size_t payload);
    EXPORT QtEvent QtGetCurrentEvent(QtEventListener listener);
    EXPORT void QtRemoveEventListener(void* obj_ptr, QtEventListener listener);
    EXPORT size_t QtResizeEventGetWidth(QtEvent ev);
    EXPORT size_t QtResizeEventGetHeight(QtEvent ev);
    EXPORT QtString QtDynamicPropertyChangeEventGetPropertyName(QtEvent ev);
    EXPORT QtPoint QtMakePoint(int x, int y);
    EXPORT int QtPointGetX(QtPoint p);
    EXPORT int QtPointGetY(QtPoint p);
    EXPORT QtString QtNewStringUTF8(const uint8_t* buf, size_t len);
    EXPORT QtString QtNewStringUTF32(const uint32_t* buf, size_t len);
    EXPORT void QtDeleteString(QtString str);
    EXPORT size_t QtStringListGetSize(QtStringList list);
    EXPORT QtString QtStringListGetItem(QtStringList list, size_t index);
    EXPORT void QtDeleteStringList(QtStringList list);
    EXPORT QtVariantList QtNewVariantList();
    EXPORT void QtVariantListAppendNumber(QtVariantList l, double n);
    EXPORT void QtVariantListAppendString(QtVariantList l, QtString str);
    EXPORT void QtDeleteVariantList(QtVariantList l);
    EXPORT QtString QtVariantMapGetString(QtVariantMap m, QtString key);
    EXPORT double QtVariantMapGetFloat(QtVariantMap m, QtString key);
    EXPORT QtBool QtVariantMapGetBool(QtVariantMap m, QtString key);
    EXPORT void QtDeleteVariantMap(QtVariantMap m);
    EXPORT size_t QtStringUTF16Length(QtString str);
    EXPORT size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf);
    EXPORT QtIcon QtNewIcon(QtPixmap pm);
    EXPORT QtIcon QtNewIconEmpty();
    EXPORT void QtDeleteIcon(QtIcon icon);
    EXPORT QtPixmap QtNewPixmapPNG(const uint8_t* buf, size_t len);
    EXPORT QtPixmap QtNewPixmapJPEG(const uint8_t* buf, size_t len);
    EXPORT void QtDeletePixmap(QtPixmap pm);
    EXPORT void QtListWidgetClear(void *widget_ptr);
    EXPORT void QtListWidgetAddItem(void* widget_ptr, QtString key_, QtString label_, QtBool as_current);
    EXPORT void QtListWidgetAddItemWithIcon(void* widget_ptr, QtString key_, QtIcon icon_, QtString label_, QtBool as_current);
    EXPORT QtBool QtListWidgetHasCurrentItem(void* widget_ptr);
    EXPORT QtString QtListWidgetGetCurrentItemKey(void *widget_ptr);
    EXPORT void QtWebViewDisableContextMenu(void* widget_ptr);
    EXPORT void QtWebViewEnableLinkDelegation(void *widget_ptr);
    EXPORT void QtWebViewSetHTML(void *widget_ptr, QtString html, QtString base_url);
    EXPORT void QtWebViewScrollToAnchor(void *widget_ptr, QtString anchor);
    EXPORT QtPoint QtWebViewGetScroll(void* widget_ptr);
    EXPORT void QtDialogExec(void *dialog_ptr);
    EXPORT void QtDialogAccept(void *dialog_ptr);
    EXPORT void QtDialogReject(void *dialog_ptr);
    EXPORT void QtWebViewSetScroll(void* widget_ptr, QtPoint pos);
    EXPORT void QtDialogShowModal(void* dialog_ptr);
    EXPORT void QtDialogSetParent(void* dialog_ptr, void* parent_ptr);
    EXPORT QtString QtFileDialogOpen(void* parent_ptr, QtString title, QtString cwd, QtString filter);
    EXPORT QtStringList QtFileDialogOpenMultiple(void* parent_ptr,  QtString title, QtString cwd, QtString filter);
    EXPORT QtString QtFileDialogSelectDirectory(void *parent_ptr, QtString title, QtString cwd);
    EXPORT QtString QtFileDialogSave(void* parent_ptr, QtString title, QtString cwd, QtString filter);
    // API (Web)
    EXPORT void WebViewLoadContent(void* view_ptr);
    EXPORT QtBool WebViewIsContentLoaded(void* view_ptr);
    EXPORT void WebViewRegisterAsset(void* view_ptr, QtString path, QtString mime, const uint8_t* buf, size_t len);
    EXPORT QtString WebViewInjectCSS(void* view_ptr, QtString path);
    EXPORT QtString WebViewInjectJS(void* view_ptr, QtString path);
    EXPORT QtString WebViewInjectTTF(void* view_ptr, QtString path, QtString family, QtString weight, QtString style);
    EXPORT void WebViewCallMethod(void* view_ptr, QtString id, QtString name, QtVariantList args);
    EXPORT QtString WebViewGetCurrentEventHandler(void* view_ptr);
    EXPORT QtVariantMap WebViewGetCurrentEventPayload(void* view_ptr);
    EXPORT void WebViewPatchActualDOM(void* view_ptr, QtString operations);
    EXPORT void* WebDialogCreate(void* parent_ptr, QtIcon icon, QtString title, int width, int height, QtBool closable);
    EXPORT void* WebDialogGetWebView(void* dialog_ptr);
    EXPORT void WebDialogDispose(void* dialog_ptr);
#ifdef __cplusplus
}
#endif

#endif

