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


#define WebUiHtmlUrl "qrc:/qtbinding/webui/webui.html"

class WebUiBridge final: public QObject {
    Q_OBJECT
signals:
    void LoadFinish();
    void EmitEvent(QString id, QString name, QVariantMap event);
    void CloseWindow();
    void EraseStyle(QString id, QString key);
    void ApplyStyle(QString id, QString key, QString value);
    void DetachEvent(QString id, QString event);
    void ModifyEvent(QString id, QString event, bool prevent, bool stop);
    void AttachEvent(QString id, QString event, bool prevent, bool stop, QString handler);
    void SetText(QString id, QString text);
    // void InsertNode(QString parent, QString ref, QString id, QString tag);
    void AppendNode(QString parent, QString id, QString tag);
    void RemoveNode(QString parent, QString id);
    void UpdateNode(QString old_id, QString new_id);
    void ReplaceNode(QString old_id, QString id, QString tag);
};

class WebUiWindow final: public QMainWindow {
    Q_OBJECT
private:
    QWebView* view;
    QWebPage* page;
    QWebFrame* frame;
    bool debug;
public:
    WebUiBridge* bridge;
    WebUiWindow(QString title): QMainWindow(nullptr), debug(true) {
        setWindowTitle(title);
        bridge = new WebUiBridge();
        connect(bridge, &WebUiBridge::EmitEvent, this, &WebUiWindow::emitEvent);
        connect(bridge, &WebUiBridge::LoadFinish, this, &WebUiWindow::loadFinished);
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
    void closeEvent(QCloseEvent* ev) override {
        ev->ignore();
        bridge->CloseWindow();
    };
    QString emittedEventNode;
    QString emittedEventName;
    QVariantMap emittedEventPayload;
    size_t detachedHandler;
public:
    QString getEmittedEventNode() const { return emittedEventNode; }
    QString getEmittedEventName() const { return emittedEventName; }
    QVariantMap getEmittedEventPayload() const { return emittedEventPayload; }
    size_t getDetachedHandler() const { return detachedHandler; }
signals:
    void loadFinished();
    void eventEmitted();
public slots:
    void emitEvent(QString id, QString name, QVariantMap event) {
        emittedEventNode = id;
        emittedEventName = name;
        emittedEventPayload = event;
        eventEmitted();
        if (debug) {
            qDebug() << "Event: " << id << " " << name;
        }
    };
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
};

#endif  // WEBUI_HPP
