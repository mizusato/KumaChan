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

#ifdef __cplusplus
extern "C" {
#endif
    void QtInit();
    int QtMainLoop();
    void QtCommitTask(void (*task)(void*), void* arg);
    void QtExit(int code);
    void QtQuit();
    void* QtLoadWidget(const char* definition);
    void* QtWidgetFindChild(void* widget_ptr, const char* name);
    void QtWidgetShow(void* widget_ptr);
    void QtWidgetHide(void* widget_ptr);
    QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val);
    QtString QtObjectGetPropString(void* obj_ptr, const char* prop);
    QtConnHandle QtConnect(void* widget_ptr, const char* signal, void (*callback)(void*,void*), void* payload);
    QtBool QtIsConnectionValid(QtConnHandle handle);
    QtBool QtDisconnect(QtConnHandle handle);
    void QtDeleteConnection(QtConnHandle handle);
    QtString QtNewStringUTF8(const char* buf, size_t len);
    QtString QtNewStringUTF32(const uint32_t* buf, size_t len);
    void QtDeleteString(QtString str);
#ifdef __cplusplus
}
#endif

#endif
