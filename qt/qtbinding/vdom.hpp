#ifndef VDOM_HPP
#define VDOM_HPP

#include <QObject>
#include <QString>
#include <QList>
#include <QMap>
#include <cassert>


using uintptr = size_t;

class EventOptions;
class DeltaNotifier;

class Node final: public QObject {
    Q_OBJECT
public:
    QString tagName;
    QMap<QString,QString> style;
    QMap<QString,EventOptions*> events;
    QList<Node*> children;
    QString key;
    static void diff(DeltaNotifier* ctx, Node* parent, Node* old, Node* _new);
};

class EventOptions: public QObject {
    Q_OBJECT
public:
    bool prevent;
    bool stop;
    size_t handler;
};

class DeltaNotifier {
public:
    virtual ~DeltaNotifier() {}
    virtual void InsertNode(uintptr parent, uintptr ref, uintptr id, QString tag) = 0;
    virtual void AppendNode(uintptr parent, uintptr id, QString tag) = 0;
    virtual void RemoveNode(uintptr parent, uintptr id) = 0;
    virtual void UpdateNode(uintptr old_id, uintptr new_id) = 0;
    virtual void ReplaceNode(uintptr old_id, uintptr id, QString tag) = 0;
    virtual void EraseStyle(uintptr id, QString key) = 0;
    virtual void ApplyStyle(uintptr id, QString key, QString value) = 0;
    virtual void DetachEvent(uintptr id, QString event, bool dispose_handler) = 0;
    virtual void AttachEvent(uintptr id, QString event, bool prevent, bool stop, size_t handler) = 0;
};

Q_DECLARE_INTERFACE(DeltaNotifier, "DeltaNotifier");


#endif  // VDOM_HPP
