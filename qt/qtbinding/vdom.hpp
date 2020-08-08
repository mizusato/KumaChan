#ifndef VDOM_HPP
#define VDOM_HPP

#include <QObject>
#include <QString>
#include <QList>
#include <QMap>
#include <cassert>


using NodeId = size_t;

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
    virtual void InsertNode(NodeId parent, NodeId ref, NodeId id, QString tag) = 0;
    virtual void AppendNode(NodeId parent, NodeId id, QString tag) = 0;
    virtual void RemoveNode(NodeId parent, NodeId id) = 0;
    virtual void UpdateNode(NodeId old_id, NodeId new_id) = 0;
    virtual void ReplaceNode(NodeId old_id, NodeId id, QString tag) = 0;
    virtual void EraseStyle(NodeId id, QString key) = 0;
    virtual void ApplyStyle(NodeId id, QString key, QString value) = 0;
    virtual void DetachEvent(NodeId id, QString event, size_t handler) = 0;
    virtual void ModifyEvent(NodeId id, QString event, bool prevent, bool stop) = 0;
    virtual void AttachEvent(NodeId id, QString event, bool prevent, bool stop, size_t handler) = 0;
};

Q_DECLARE_INTERFACE(DeltaNotifier, "DeltaNotifier");


#endif  // VDOM_HPP
