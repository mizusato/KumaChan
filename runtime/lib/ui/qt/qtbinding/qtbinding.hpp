#ifndef QTBINDING_HPP
#define QTBINDING_HPP
#include <QObject>
#include <QWidget>
#include <QEvent>
#include <QUiLoader>
#include <QWebView>
#include <cstdlib>


#define CALLBACK_BLOCKED "qtbindingCallbackBlocked"

typedef void (*callback_t)(size_t);
Q_DECLARE_METATYPE(callback_t);

class Bridge final: public QObject {
    Q_OBJECT
public:
    Bridge(QWidget *parent = nullptr): QObject(parent) {
        QObject::connect (
            this, &Bridge::QueueCallback,
            this, &Bridge::__InvokeCallback,
            Qt::QueuedConnection
        );
    };
    virtual ~Bridge() {};
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

class UiLoader: public QUiLoader {
public:
    virtual QWidget* createWidget(const QString &className, QWidget *parent = nullptr, const QString &name = QString()) override {
        if (className == "QWebView") {
            QWidget* w = new QWebView(parent);
            w->setObjectName(name);
            return w;
        } else {
            return QUiLoader::createWidget(className, parent, name);
        }
    }
};

#endif  // QTBINDING_HPP
