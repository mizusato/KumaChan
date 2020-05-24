#include <QApplication>
#include <QUiLoader>
#include <QBuffer>
#include <QByteArray>
#include "qtbinding.hpp"
#include "qtbinding.h"


struct __QtConnHandle {
    QMetaObject::Connection
        conn;
    CallbackObject*
        cb_obj;
    bool
        disconnected;
};

static QApplication*
    app = nullptr;
static Bridge*
    bridge = nullptr;
static QUiLoader*
    loader = nullptr;
static bool
    initialized = false;

void QtInit() {
    static int fake_argc = 1;
    static char fake_arg[] = {'Q','t','A','p','p','\0'};
    static char* fake_argv[] = { fake_arg };
    if (!(initialized)) {
        app = new QApplication(fake_argc, fake_argv);
        bridge = new Bridge();
        loader = new QUiLoader();
        initialized = true;
    }
}

int QtMainLoop() {
    return app->exec();
}

void QtCommitTask(void (*task)(void*), void* arg) {
    bridge->QueueCallback(task, arg);
}

void QtExit(int code) {
    app->exit(code);
}

void QtQuit() {
    app->quit();
}

void* QtLoadWidget(const char* definition) {
    QByteArray bytes(definition);
    QBuffer buf(&bytes);
    QWidget* widget = loader->load(&buf, nullptr);
    return (void*) widget;
};

void QtWidgetShow(void *widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->show();
}

void QtWidgetHide(void *widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->hide();
}

QtConnHandle QtConnect (
        void* widget_ptr,
        const char* signal,
        void (*callback)(void*,void*),
        void* payload
) {
    QWidget* widget = (QWidget*) widget_ptr;
    CallbackObject* cb_obj = new CallbackObject(callback, widget_ptr, payload);
    __QtConnHandle* handle = new __QtConnHandle;
    handle->conn = QObject::connect(widget, signal, cb_obj, "Callback");
    handle->cb_obj = cb_obj;
    QtConnHandle c_handle;
    c_handle.raw = (void*) handle;
    return c_handle;
}

bool QtIsConnectionValid(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.raw;
    return bool(h->conn);
};

bool QtDisconnect(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.raw;
    if (!(h->conn) || h->disconnected) {
        return false;
    }
    QObject::disconnect(h->conn);
    delete h->cb_obj;
    h->cb_obj = nullptr;
    h->disconnected = true;
    return true;
};

void QtFreeConnection(QtConnHandle handle) {
    QtDisconnect(handle);
    __QtConnHandle* h = (__QtConnHandle*) handle.raw;
    delete h;
};
