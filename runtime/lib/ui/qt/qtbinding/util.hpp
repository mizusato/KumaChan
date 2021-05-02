#ifndef UTIL_HPP
#define UTIL_HPP

#include <QObject>
#include <QMetaMethod>
#include <QWidget>
#include <QEvent>
#include <QWebEnginePage>
#include <cstdlib>
#include <cmath>
#include "qtbinding.h"


#define BaseSize 18.0
#define BaseScreen 768.0

QtString WrapString(QString str);
QString UnwrapString(QtString str);

QString EncodeBase64(QString str);
QString DecodeBase64(QString str);

int Get1remPixels();
QSize GetSizeFromRelative(QSize size_rem);
void MoveToScreenCenter(QWidget* widget);

bool DebugEnabled();
void EnableDebug();


#define CALLBACK_BLOCKED "qtbindingCallbackBlocked"

typedef void (*callback_t)(size_t);
Q_DECLARE_METATYPE(callback_t);

class CallbackObject;
struct ConnectionHandle {
    QMetaObject::Connection
        conn;
    CallbackObject*
        cb_obj;
};

QMetaObject::Connection QtDynamicConnect (
        QObject* emitter , const QString& signalName,
        QObject* receiver, const QString& slotName
);

class CallbackExecutor final: public QObject {
    Q_OBJECT
public:
    CallbackExecutor(QWidget *parent = nullptr): QObject(parent) {
        QObject::connect (
            this, &CallbackExecutor::QueueCallback,
            this, &CallbackExecutor::__InvokeCallback,
            Qt::QueuedConnection
        );
    };
    virtual ~CallbackExecutor() {};
signals:
    void QueueCallback(callback_t cb, size_t payload);
private slots:
    void __InvokeCallback(callback_t cb, size_t payload) {
        cb(payload);
    };
};

class CallbackObject final: public QObject {
    Q_OBJECT
public:
    callback_t cb;
    size_t payload;
    CallbackObject(QObject* parent, callback_t cb, size_t payload): QObject(parent) {
        this->cb = cb;
        this->payload = payload;
    };
    virtual ~CallbackObject() {};
public slots:
    void slot() {
        QVariant v = parent()->property(CALLBACK_BLOCKED);
        if (v.isValid() && v.toBool() == true) {
            return;
        };
        cb(payload);
    };
};

class EventListener final: public QObject {
    Q_OBJECT
public:
    QEvent::Type accept_type;
    bool prevent_default;
    callback_t cb;
    size_t payload;
    QObject* current_object;
    QEvent* current_event;
    EventListener(QEvent::Type t, bool prevent, callback_t cb, size_t payload): QObject(nullptr) {
        this->accept_type = t;
        this->prevent_default = prevent;
        this->cb = cb;
        this->payload = payload;
    }
    bool eventFilter(QObject* obj, QEvent *event) {
        if (event->type() == accept_type) {
            current_object = obj;
            current_event = event;
            cb(payload);
            current_event = nullptr;
            current_object = nullptr;
            if (prevent_default) {
                return true;
            } else {
                return false;
            }
        } else {
            return false;
        }
    }
};

class LinkDelegatedPage: public QWebEnginePage {
public:
    LinkDelegatedPage(QObject* parent): QWebEnginePage(parent) {}
    bool acceptNavigationRequest(const QUrl &url, NavigationType type, bool isMainFrame) override {
        if (type == NavigationTypeLinkClicked) {
            parent()->setProperty("qtbindingClickedLinkUrl", url.toString());
            emit loadFinished(false);
            return false;
        } else {
            return QWebEnginePage::acceptNavigationRequest(url, type, isMainFrame);
        }
    }
};

#endif  // UTIL_HPP

