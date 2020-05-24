#ifndef HELLO_H
#define HELLO_H
#include <QObject>
#include <QWidget>

typedef void (*Callback1)(void*);

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
    void QueueCallback(Callback1 callback, void* payload);
private slots:
    void __InvokeCallback(Callback1 callback, void* payload) {
        callback(payload);
    };
};

class CallbackObject: public QObject {
Q_OBJECT
public:
    void (*callback)(void*,void*);
    void* target;
    void* payload;
    CallbackObject(void (*callback)(void*,void*), void* target, void* payload): QObject(nullptr) {
        this->callback = callback;
        this->target = target;
        this->payload = payload;
    };
    virtual ~CallbackObject() {};
public slots:
    void Callback() {
        this->callback(this->target, this->payload);
    };
};

#endif
