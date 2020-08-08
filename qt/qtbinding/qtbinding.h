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

struct _QtVariantMap {
    void* ptr;
};
typedef struct _QtVariantMap QtVariantMap;

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

struct _QtWebUiNode {
    void* ptr;
};
typedef struct _QtWebUiNode QtWebUiNode;

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
    size_t QtStringUTF16Length(QtString str);
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
    size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf);
    void QtWebUiInit(QtString title);
    void* QtWebUiGetWindow();
    void QtWebUiUpdateVDOM(QtWebUiNode root);
    QtWebUiNode QtWebUiNewNode(QtString tagName);
    void QtWebUiNodeAddStyle(QtWebUiNode node, QtString key, QtString value);
    void QtWebUiNodeAddEvent(QtWebUiNode node, QtString name, QtBool prevent, QtBool stop, size_t handler);
    void QtWebUiNodeSetText(QtWebUiNode node, QtString text);
    void QtWebUiNodeAppendChild(QtWebUiNode node, QtWebUiNode child);
#ifdef __cplusplus
}
#endif

#endif
