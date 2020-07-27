#ifndef HELLO_H
#define HELLO_H
#include <QObject>
#include <QWidget>
#include <QEvent>
#include <cstdlib>


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
    CallbackObject(callback_t cb, size_t payload): QObject(nullptr) {
        this->cb = cb;
        this->payload = payload;
    };
    virtual ~CallbackObject() {};
public slots:
    void slot() {
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

#endif
