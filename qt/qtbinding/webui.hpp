#ifndef WEBUI_HPP
#define WEBUI_HPP

#include <QMainWindow>
#include <QCloseEvent>
#include <QUrl>
#include <QWebView>
#include <QWebPage>
#include <QWebFrame>
#include <QWebInspector>
#include <QDialog>
#include <QVBoxLayout>
#include "vdom.hpp"


#define WebUiHtmlUrl "qrc:/qtbinding/webui/webui.html"

class WebUiBridge final: public QObject, public DeltaNotifier {
    Q_OBJECT
    Q_INTERFACES(DeltaNotifier)
signals:
    void CloseWindow();
    void EmitEvent(NodeId id, QString name, QVariantMap event);
    void InsertNode(NodeId parent, NodeId ref, NodeId id, QString tag);
    void AppendNode(NodeId parent, NodeId id, QString tag);
    void RemoveNode(NodeId parent, NodeId id);
    void UpdateNode(NodeId old_id, NodeId new_id);
    void ReplaceNode(NodeId old_id, NodeId id, QString tag);
    void EraseStyle(NodeId id, QString key);
    void ApplyStyle(NodeId id, QString key, QString value);
    void DetachEvent(NodeId id, QString event, size_t handler);
    void ModifyEvent(NodeId id, QString event, bool prevent, bool stop);
    void AttachEvent(NodeId id, QString event, bool prevent, bool stop, size_t handler);
};

class WebUiWindow final: public QMainWindow {
    Q_OBJECT
private:
    QWebView* view;
    QWebPage* page;
    QWebFrame* frame;
    WebUiBridge* bridge;
    Node *vdom;
    bool debug;
public:
    WebUiWindow(QString title): QMainWindow(nullptr), vdom(nullptr), debug(true) {
        setWindowTitle(title);
        bridge = new WebUiBridge();
        connect(bridge, &WebUiBridge::AttachEvent, this, [this]
                (NodeId, QString, bool, bool, size_t handler) -> void {
            targetEventHandler = handler;
            eventAttached();
        });
        connect(bridge, &WebUiBridge::DetachEvent, this, [this]
                (NodeId, QString, size_t handler) -> void {
            targetEventHandler = handler;
            eventDetached();
        });
        connect(bridge, &WebUiBridge::EmitEvent, this, &WebUiWindow::emitEvent);
        view = new QWebView(this);
        view->setUrl(QUrl(WebUiHtmlUrl));
        view->setContextMenuPolicy(Qt::NoContextMenu);
        page = view->page();
        page->settings()->setAttribute(QWebSettings::DeveloperExtrasEnabled, true);
        frame = page->mainFrame();
        frame->addToJavaScriptWindowObject("WebUI", bridge);
        setCentralWidget(view);
        show();
        if (debug) {
            openInspector();
        }
    }
    ~WebUiWindow() {}
private:
    void openInspector() {
        QWebInspector* inspector = new QWebInspector(this);
        inspector->setPage(page);
        QDialog* inspector_dialog = new QDialog();
        inspector_dialog->setLayout(new QVBoxLayout());
        inspector_dialog->layout()->addWidget(inspector);
        inspector_dialog->setModal(false);
        inspector_dialog->resize(800, 360);
        inspector_dialog->layout()->setContentsMargins(0, 0, 0, 0);
        inspector_dialog->setWindowTitle(tr("Webkit Inspector"));
        inspector_dialog->show();
        inspector_dialog->raise();
    }
    void closeEvent(QCloseEvent* ev) override {
        ev->ignore();
        bridge->CloseWindow();
    };
    NodeId emittedEventNode;
    QString emittedEventName;
    QVariantMap emittedEventPayload;
    size_t targetEventHandler;
public:
    NodeId getEmittedEventNode() const { return emittedEventNode; }
    QString getEmittedEventName() const { return emittedEventName; }
    QVariantMap getEmittedEventPayload() const { return emittedEventPayload; }
    size_t getTargetEventHandler() const { return targetEventHandler; }
signals:
    void eventAttached();
    void eventEmitted();
    void eventDetached();
public slots:
    void emitEvent(NodeId id, QString name, QVariantMap event) {
        emittedEventNode = id;
        emittedEventName = name;
        emittedEventPayload = event;
        eventEmitted();
        if (debug) {
            qDebug() << "Event: " << id << " " << name;
        }
    };
};

#endif  // WEBUI_HPP
