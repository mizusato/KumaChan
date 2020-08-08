#ifndef WEBUI_HPP
#define WEBUI_HPP

#include <QMainWindow>
#include <QWebView>
#include <QCloseEvent>
#include <QUrl>
#include "vdom.hpp"


class WebUiBridge final: public QObject, public DeltaNotifier {
    Q_OBJECT
    Q_INTERFACES(DeltaNotifier)
signals:
    void Close();
    void InsertNode(uintptr parent, uintptr ref, uintptr id, QString tag);
    void AppendNode(uintptr parent, uintptr id, QString tag);
    void RemoveNode(uintptr parent, uintptr id);
    void UpdateNode(uintptr old_id, uintptr new_id);
    void ReplaceNode(uintptr old_id, uintptr id, QString tag);
    void EraseStyle(uintptr id, QString key);
    void ApplyStyle(uintptr id, QString key, QString value);
    void DetachEvent(uintptr id, QString event);
    void AttachEvent(uintptr id, QString event, bool prevent, bool stop, size_t handler);
};

class WebUiWindow final: public QMainWindow {
    Q_OBJECT
private:
    QWebView* view;
    WebUiBridge* bridge;
    WebUiWindow(QString page, QString title) {
        bridge = new WebUiBridge();
        view = new QWebView(this);
        view->setUrl(QUrl::fromLocalFile(page));
        setWindowTitle(title);
    }
    ~WebUiWindow() {}
    void closeEvent(QCloseEvent* ev) override {
        ev->ignore();
        bridge->Close();
    };
    QVariantMap emittedEvent;
    size_t targetEventHandler;
    QVariantMap getEmittedEvent() const { return emittedEvent; }
    size_t getTargetEventHandler() const { return targetEventHandler; }
signals:
    void eventAttached();
    void eventEmitted();
    void eventDetached();
};

#endif  // WEBUI_HPP
