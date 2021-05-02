#ifndef WEB_HPP
#define WEB_HPP

#include <QGuiApplication>
#include <QScreen>
#include <QMainWindow>
#include <QCloseEvent>
#include <QUrl>
#include <QDialog>
#include <QVBoxLayout>
#include <QBuffer>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QWebEngineView>
#include <QWebEngineProfile>
#include <QWebEngineUrlSchemeHandler>
#include <QWebEngineUrlRequestJob>
#include <cmath>
#include "util.hpp"


#define WebViewContent "qrc:/qtbinding/web/content.html"
#define WebAssetScheme "asset"

class WebBridge final: public QObject {
    Q_OBJECT
private:
    QWebEngineView* view;
public:
    WebBridge(QWebEngineView* parent): QObject(parent), view(parent) {};
private:
    void RunJS(QString method, QJsonObject args) {
        QString msg = QString::fromUtf8(QJsonDocument(args).toJson());
        QString script = QString("WebBridge.%1(%2);").arg(method, msg);
        view->page()->runJavaScript(script);
    }
public:
    void UpdateRootFontSize(double size) {
        QJsonObject args;
        args["size"] = qreal(size);
        RunJS("UpdateRootFontSize", args);
    }
    void InjectCSS(QString uuid, QString path) {
        QJsonObject args;
        args["uuid"] = uuid;
        args["path"] = path;
        RunJS("InjectCSS", args);
    }
    void InjectJS(QString uuid, QString path) {
        QJsonObject args;
        args["uuid"] = uuid;
        args["path"] = path;
        RunJS("InjectJS", args);
    }
    void InjectTTF(QString uuid, QString path, QString family, QString weight, QString style) {
        QJsonObject args;
        args["uuid"] = uuid;
        args["path"] = path;
        args["family"] = family;
        args["weight"] = weight;
        args["style"] = style;
        RunJS("InjectTTF", args);
    }
    void CallMethod(QString id, QString method, QVariantList args) {
        QJsonObject bridge_args;
        bridge_args["id"] = id;
        bridge_args["method"] = method;
        bridge_args["args"] = QJsonArray::fromVariantList(args);
        RunJS("CallMethod", bridge_args);
    }
    void PatchActualDOM(QString data) {
        QJsonObject args;
        args["data"] = data;
        RunJS("PatchActualDOM", args);
    }
};

struct WebAsset {
    bool        exists;
    QString     mime;
    QByteArray  content;
};

class WebAssetStore final: public QObject {
    Q_OBJECT
    QHash<QString,WebAsset> mapping;
public:
    WebAssetStore(QObject* parent): QObject(parent) {};
    void InsertItem(QString path, QString mime, QByteArray data) {
        // qDebug() << "register" << path << "(" << data.size() << ")";
        mapping[path] = { true, mime, data };
    };
    WebAsset LookupItem(QString path) {
        return mapping.value(path);
    };
};

class WebAssetSchemeHandler: public QWebEngineUrlSchemeHandler {
    Q_OBJECT
private:
    WebAssetStore* store;
public:
    WebAssetSchemeHandler(QObject* parent, WebAssetStore* store):
        QWebEngineUrlSchemeHandler(parent), store(store) {};
    void requestStarted(QWebEngineUrlRequestJob* req) {
        QUrl url = req->requestUrl();
        QString raw_path = url.path();
        QString path = DecodeBase64(raw_path);
        WebAsset asset = store->LookupItem(path);
        if (asset.exists) {
            QByteArray* data = new QByteArray(asset.content);
            QBuffer* buf = new QBuffer(data, nullptr);
            QByteArray* contentType = new QByteArray(asset.mime.toUtf8());
            req->reply(*contentType, buf);
            connect(req, &QObject::destroyed, [contentType,buf,data] ()->void {
                delete contentType;
                delete buf;
                delete data;
            });
        } else {
            req->fail(QWebEngineUrlRequestJob::UrlNotFound);
        }
    }
};

class WebViewInterface {
public:
    virtual void emitEvent(QString handler, QVariantMap payload) = 0;
    virtual void finishLoad() = 0;
};

class WebViewPage final: public QWebEnginePage {
    Q_OBJECT
private:
    WebViewInterface* view;
public:
    WebViewPage(QWebEngineProfile* profile, QObject* parent, WebViewInterface* view):
        QWebEnginePage(profile, parent), view(view) {}
    void javaScriptAlert(const QUrl &securityOrigin, const QString &msg) override {
        // We abuse the blocking nature of the alert() function to
        // achieve synchronous IPC between Qt and Web Contents.
        static_cast<void>(securityOrigin);
        QString ipc_prefix = "IPC:";
        if (msg.startsWith(ipc_prefix)) {
            QString ipc_payload = msg.mid(ipc_prefix.length());
            QJsonDocument doc = QJsonDocument::fromJson(ipc_payload.toUtf8());
            if (doc.isNull()) { return; }
            auto obj = doc.object();
            auto method = obj["method"].toString();
            if (method == "finishLoad") {
                view->finishLoad();
            } else if (method == "emitEvent") {
                auto args = obj["args"].toObject();
                QString handler = args["handler"].toString();
                QVariantMap payload = args["payload"].toObject().toVariantMap();
                view->emitEvent(handler, payload);
            }
        }
    }
    bool javaScriptConfirm(const QUrl &securityOrigin, const QString &msg) override {
        static_cast<void>(securityOrigin);
        static_cast<void>(msg);
        // reserved
        return false;
    }
    bool javaScriptPrompt(const QUrl &securityOrigin, const QString &msg, const QString &defaultValue, QString *result) override {
        static_cast<void>(securityOrigin);
        static_cast<void>(msg);
        static_cast<void>(defaultValue);
        static_cast<void>(result);
        // reserved
        return false;
    }
    void javaScriptConsoleMessage(JavaScriptConsoleMessageLevel level, const QString &message, int lineNumber, const QString &sourceID) override {
        QWebEnginePage::javaScriptConsoleMessage(level, message, lineNumber, sourceID);
        // reserved
    }
};

class WebView final: public QWebEngineView, public WebViewInterface {
    Q_OBJECT
public:
    WebView(QWidget* parent): QWebEngineView(parent) {
        initPage();
        initBridge();
        setContextMenuPolicy(Qt::NoContextMenu);
    }
    ~WebView() {}
    WebBridge* getBridge() {
        return bridge;
    }
    WebAssetStore* getStore() {
        return store;
    }
private:
    bool contentLoadStarted = false;
    bool contentLoaded = false;
public:
    void LoadContent() {
        if (contentLoadStarted) {
            return;
        } else {
            contentLoadStarted = true;
        }
        setUrl(QUrl(WebViewContent));
        if (DebugEnabled()) {
            openInspector();
        }
    }
    bool IsContentLoaded() {
        return contentLoaded;
    }
signals:
    void loadFinished();
    void eventEmitted();
public:
    void emitEvent(QString handler, QVariantMap payload) override {
        emittedEventHandler = handler;
        emittedEventPayload = payload;
        emit eventEmitted();
        emittedEventHandler = "";
        emittedEventPayload = QVariantMap();
        // if (debug) {
        //     qDebug() << "Event: " << handler;
        // }
    }
    void finishLoad() override {
        contentLoaded = true;
        syncRootFontSizeWithScreenSize();
        emit loadFinished();
    }
private:
    QString emittedEventHandler;
    QVariantMap emittedEventPayload;
public:
    QString getEmittedEventHandler() const { return emittedEventHandler; }
    QVariantMap getEmittedEventPayload() const { return emittedEventPayload; }
private:
    WebAssetStore* store = nullptr;
    void initPage() {
        store = new WebAssetStore(this);
        WebAssetSchemeHandler* handler = new WebAssetSchemeHandler(this, store);
        QWebEngineProfile* profile = new QWebEngineProfile(this);
        profile->installUrlSchemeHandler(WebAssetScheme, handler);
        WebViewPage* page = new WebViewPage(profile, this, this);
        setPage(page);
    };
private:
    WebBridge* bridge;
    void initBridge() {
        bridge = new WebBridge(this);
    }
    void syncRootFontSizeWithScreenSize() {
        updateRootFontSize();
        QScreen *screen = QGuiApplication::primaryScreen();
        connect(screen, &QScreen::geometryChanged, this, &WebView::updateRootFontSize);
    }
    void updateRootFontSize() {
        int fontSize = Get1remPixels();
        bridge->UpdateRootFontSize(double(fontSize));
    }
private:
    QDialog* inspector_dialog = nullptr;
    void openInspector() {
        if (inspector_dialog != nullptr) {
            return;
        }
        QWebEngineView* inspector = new QWebEngineView(this);
        inspector->page()->setInspectedPage(this->page());
        inspector_dialog = new QDialog(this);
        inspector_dialog->setLayout(new QVBoxLayout(inspector_dialog));
        inspector_dialog->layout()->addWidget(inspector);
        inspector_dialog->setModal(false);
        inspector_dialog->resize(800, 360);
        inspector_dialog->layout()->setContentsMargins(0, 0, 0, 0);
        inspector_dialog->setWindowTitle(tr("WebEngine Inspector"));
        inspector_dialog->show();
        inspector_dialog->raise();
        MoveToScreenCenter(inspector_dialog);
    }
};

class WebDialog final: public QDialog {
    Q_OBJECT
private:
    bool closable;
    WebView* view;
public:
    WebDialog(QWidget* parent, QIcon icon, QString title, QSize size_rem, bool closable): QDialog(parent), closable(closable) {
        view = new WebView(this);
        setLayout(new QVBoxLayout(this));
        layout()->addWidget(view);
        layout()->setContentsMargins(0, 0, 0, 0);
        setWindowIcon(icon);
        setWindowTitle(title);
        resize(GetSizeFromRelative(size_rem));
        if (!(closable)) {
            setWindowFlag(Qt::CustomizeWindowHint);
            setWindowFlag(Qt::WindowTitleHint);
        }
    }
    ~WebDialog() {}
    WebView* getWebView() {
        return view;
    }
public slots:
    virtual void reject() override {
        if (!(closable)) {
            // it is not clear whether custom window hint works on all platforms
            return;
        }
        QDialog::reject();
    }
};

#endif  // WEB_HPP

