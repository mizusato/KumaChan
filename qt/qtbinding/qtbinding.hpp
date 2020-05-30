#ifndef HELLO_H
#define HELLO_H
#include <QObject>
#include <QWidget>
#include <cstdlib>


typedef void (*callback_t)(size_t);
Q_DECLARE_METATYPE(callback_t);

class Bridge: public QObject {
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

class CallbackObject: public QObject {
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
        this->cb(this->payload);
    };
};

#endif
