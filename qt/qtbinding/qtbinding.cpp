#include <QApplication>
#include <QUiLoader>
#include <QBuffer>
#include <QByteArray>
#include <QString>
#include <QVariant>
#include "qtbinding.hpp"
#include "qtbinding.h"


QtString QtWrapString(QString str);
QString QtUnwrapString(QtString str);

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

void* QtWidgetFindChild(void* widget_ptr, const char* name) {
    QWidget* widget = (QWidget*) widget_ptr;
    QWidget* child = widget->findChild<QWidget*>(QString(name));
    return (void*) child;
}

void QtWidgetShow(void* widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->show();
}

void QtWidgetHide(void* widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->hide();
}

QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val) {
    QObject* obj = (QObject*) obj_ptr;
    return obj->setProperty(prop, QtUnwrapString(val));
}

QtString QtObjectGetPropString(void* obj_ptr, const char* prop) {
    QObject* obj = (QObject*) obj_ptr;
    QVariant val = obj->property(prop);
    return QtWrapString(val.toString());
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
    return { (void*) handle };
}

QtBool QtIsConnectionValid(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    return bool(h->conn);
};

QtBool QtDisconnect(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    if (!(h->conn) || h->disconnected) {
        return false;
    }
    QObject::disconnect(h->conn);
    delete h->cb_obj;
    h->cb_obj = nullptr;
    h->disconnected = true;
    return true;
};

void QtDeleteConnection(QtConnHandle handle) {
    QtDisconnect(handle);
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    delete h;
};

QtString QtNewStringUTF8(const char* buf, size_t len) {
    QString* ptr = new QString;
    *ptr = QString::fromUtf8(buf, len);
    return { (void*) ptr };
}

QtString QtNewStringUTF32(const uint32_t* buf, size_t len) {
    QString* ptr = new QString;
    *ptr = QString::fromUcs4(buf, len);
    return { (void*) ptr };
}

QtString QtWrapString(QString str) {
    QString* ptr = new QString;
    *ptr = str;
    return { (void*) ptr };
}

QString QtUnwrapString(QtString str) {
    return QString(*(QString*)(str.ptr));
}

void QtDeleteString(QtString str) {
    delete (QString*)(str.ptr);
}
