#ifndef VDOM_HPP
#define VDOM_HPP

#include <QObject>
#include <QString>
#include <QList>
#include <QMap>
#include <cassert>


class EventOptions;
class DeltaNotifier;

class Node final: public QObject {
    Q_OBJECT
public:
    Node(): QObject(nullptr) {};
    ~Node() {
        for (Node* child: children) {
            delete child;
        }
    }
    QString tagName;
    QMap<QString,QString> style;
    QMap<QString,EventOptions*> events;
    bool is_text;
    QString text;
    QList<Node*> children;
    static void diff(DeltaNotifier* ctx, Node* parent, Node* old, Node* _new);
};

class EventOptions: public QObject {
    Q_OBJECT
public:
    EventOptions(Node* parent): QObject(parent) {};
    bool prevent;
    bool stop;
    QString handler;
};

class DeltaNotifier {
public:
    virtual ~DeltaNotifier() {}
    virtual void SetText(QString id, QString text) = 0;
    virtual void InsertNode(QString parent, QString ref, QString id, QString tag) = 0;
    virtual void AppendNode(QString parent, QString id, QString tag) = 0;
    virtual void RemoveNode(QString parent, QString id) = 0;
    virtual void UpdateNode(QString old_id, QString new_id) = 0;
    virtual void ReplaceNode(QString old_id, QString id, QString tag) = 0;
    virtual void EraseStyle(QString id, QString key) = 0;
    virtual void ApplyStyle(QString id, QString key, QString value) = 0;
    virtual void DetachEvent(QString id, QString event, QString handler) = 0;
    virtual void ModifyEvent(QString id, QString event, bool prevent, bool stop) = 0;
    virtual void AttachEvent(QString id, QString event, bool prevent, bool stop, QString handler) = 0;
};

Q_DECLARE_INTERFACE(DeltaNotifier, "DeltaNotifier");


#endif  // VDOM_HPP
