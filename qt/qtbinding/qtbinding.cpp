#include <QApplication>
#include <QMetaMethod>
#include <QDir>
#include <QUiLoader>
#include <QBuffer>
#include <QByteArray>
#include <QString>
#include <QVector>
#include <QVariant>
#include <QResizeEvent>
#include "qtbinding.hpp"
#include "qtbinding.h"


const size_t QtEventMove = QEvent::Move;
const size_t QtEventResize = QEvent::Resize;

QtString QtWrapString(QString str);
QString QtUnwrapString(QtString str);
QMetaObject::Connection QtDynamicConnect (
        QObject* emitter , const QString& signalName,
        QObject* receiver, const QString& slotName
);

struct __QtConnHandle {
    QMetaObject::Connection
        conn;
    CallbackObject*
        cb_obj;
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
        qRegisterMetaType<callback_t>();
        initialized = true;
    }
}

int QtMain() {
    return app->exec();
}

void QtCommitTask(callback_t cb, size_t payload) {
    bridge->QueueCallback(cb, payload);
}

void QtExit(int code) {
    app->exit(code);
}

void QtQuit() {
    app->quit();
}

void* QtLoadWidget(const char* definition, const char* directory) {
    QByteArray bytes(definition);
    QBuffer buf(&bytes);
    QDir dir(directory);
    loader->setWorkingDirectory(dir);
    QWidget* widget = loader->load(&buf, nullptr);
    return (void*) widget;
}

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
        void* obj_ptr,
        const char* signal,
        callback_t cb,
        size_t payload
) {
    QObject* target_obj = (QObject*) obj_ptr;
    CallbackObject* cb_obj = new CallbackObject(cb, payload);
    __QtConnHandle* handle = new __QtConnHandle;
    handle->conn = QtDynamicConnect(target_obj, signal, cb_obj, "slot()");
    handle->cb_obj = cb_obj;
    return { (void*) handle };
}

QtBool QtIsConnectionValid(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    return bool(h->conn);
};

void QtDisconnect(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    if (!(h->conn)) {
        QObject::disconnect(h->conn);
    }
    delete h->cb_obj;
    h->cb_obj = nullptr;
    delete h;
};

QtEventListener QtAddEventListener (
        void*       obj_ptr,
        size_t      kind,
        QtBool      prevent,
        callback_t  cb,
        size_t      payload
) {
    QObject* obj = (QObject*) obj_ptr;
    QEvent::Type q_kind = (QEvent::Type) kind;
    EventListener* l = new EventListener(q_kind, prevent, cb, payload);
    obj->installEventFilter(l);
    QtEventListener wrapped;
    wrapped.ptr = l;
    return wrapped;
}

QtEvent QtGetCurrentEvent(QtEventListener listener) {
    EventListener* l = (EventListener*) listener.ptr;
    QtEvent wrapped;
    wrapped.ptr = l->current_event;
    return wrapped;
}

void QtRemoveEventListener(void* obj_ptr, QtEventListener listener) {
    QObject* obj = (QObject*) obj_ptr;
    EventListener* l = (EventListener*) listener.ptr;
    obj->removeEventFilter(l);
    delete l;
}

size_t QtResizeEventGetWidth(QtEvent ev) {
    QResizeEvent* resize = (QResizeEvent*) ev.ptr;
    return resize->size().width();
}

size_t QtResizeEventGetHeight(QtEvent ev) {
    QResizeEvent* resize = (QResizeEvent*) ev.ptr;
    return resize->size().height();
}

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

size_t QtStringUTF16Length(QtString str) {
    return QtUnwrapString(str).length();
}

size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf) {
    QVector<uint> vec = QtUnwrapString(str).toUcs4();
    size_t len = 0;
    for(auto rune: vec) {
        *buf = rune;
        buf += 1;
        len += 1;
    }
    return len;
}

QMetaObject::Connection QtDynamicConnect (
        QObject* emitter , const QString& signalName,
        QObject* receiver, const QString& slotName
) {
    /* ref: https://stackoverflow.com/questions/26208851/qt-connecting-signals-and-slots-from-text */
    int index = emitter->metaObject()
                ->indexOfSignal(QMetaObject::normalizedSignature(qPrintable(signalName)));
    if (index == -1) {
        qWarning("Wrong signal name: %s", qPrintable(signalName));
        return QMetaObject::Connection();
    }
    QMetaMethod signal = emitter->metaObject()->method(index);
    index = receiver->metaObject()
            ->indexOfSlot(QMetaObject::normalizedSignature(qPrintable(slotName)));
    if (index == -1) {
        qWarning("Wrong slot name: %s", qPrintable(slotName));
        return QMetaObject::Connection();
    }
    QMetaMethod slot = receiver->metaObject()->method(index);
    return QObject::connect(emitter, signal, receiver, slot);
}
