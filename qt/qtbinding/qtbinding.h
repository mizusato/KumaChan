#ifndef LIB_H
#define LIB_H

struct _QtConnHandle {
    void* raw;
};
typedef struct _QtConnHandle QtConnHandle;

#ifdef __cplusplus
extern "C" {
#endif
    void QtInit();
    int QtMainLoop();
    void QtCommitTask(void (*task)(void*), void* arg);
    void QtExit(int code);
    void QtQuit();
    void* QtLoadWidget(const char* definition);
    void QtWidgetShow(void *widget_ptr);
    void QtWidgetHide(void *widget_ptr);
    QtConnHandle QtConnect(void* widget_ptr, const char* signal, void (*callback)(void*,void*), void* payload);
//    TODO: bool in C
//    bool QtIsConnectionValid(QtConnHandle handle);
//    bool QtDisconnect(QtConnHandle handle);
    void QtFreeConnection(QtConnHandle handle);
#ifdef __cplusplus
}
#endif

#endif
