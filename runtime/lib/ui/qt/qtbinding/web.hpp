#ifndef WEB_HPP
#define WEB_HPP

#include <QGuiApplication>
#include <QScreen>
#include <QMainWindow>
#include <QCloseEvent>
#include <QUrl>
#include <QWebView>
#include <QWebPage>
#include <QWebFrame>
#include <QWebInspector>
#include <QNetworkAccessManager>
#include <QNetworkProxy>
#include <QNetworkReply>
#include <QTimer>
#include <QDialog>
#include <QVBoxLayout>
#include <cmath>
#include "util.hpp"


#define WebUiHtmlUrl "qrc:/qtbinding/web/content.html"

class WebBridge final: public QObject {
    Q_OBJECT
public:
    WebBridge(QWebView* parent): QObject(parent) {
        QWebFrame* frame = parent->page()->mainFrame();
        connect(frame, &QWebFrame::javaScriptWindowObjectCleared, [frame,this] () -> void {
            frame->addToJavaScriptWindowObject("WebBridge", this);
        });
    };
signals:
    void LoadFinish();
    void EmitEvent(QString handler, QVariantMap event);
    void UpdateRootFontSize(double size);
    void InjectCSS(QString uuid, QString path);
    void InjectJS(QString uuid, QString path);
    void InjectTTF(QString uuid, QString path, QString family, QString weight, QString style);
    void CallMethod(QString id, QString name, QVariantList args);
    void PatchActualDOM(QString operations);
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

class WebAssetReply final: public QNetworkReply {
    Q_OBJECT
    WebAsset asset;
    qint64 offset;
public:
    WebAssetReply(WebAssetStore* store, QString path): QNetworkReply() {
        offset = 0;
        asset = store->LookupItem(path);
        if (asset.exists) {
            open(ReadOnly | Unbuffered);
            setHeader(QNetworkRequest::ContentTypeHeader, QVariant(asset.mime));
            setHeader(QNetworkRequest::ContentLengthHeader, asset.content.size());
            metaDataChanged();
            QTimer::singleShot(0, this, &WebAssetReply::metaDataChanged);
            QTimer::singleShot(0, this, &WebAssetReply::readyRead);
            QTimer::singleShot(0, this, &WebAssetReply::finished);
        } else {
            // qDebug() << "[WebUi] asset file not found:" << path;
            setError(ContentNotFoundError, "asset file not found");
            QTimer::singleShot(0, this, &WebAssetReply::error404);
            QTimer::singleShot(0, this, &WebAssetReply::finished);
        }
    };
    void abort() override {};
    qint64 readData(char *data, qint64 maxSize) override {
        if (offset < asset.content.size()) {
            qint64 number = qMin(maxSize, asset.content.size() - offset);
            memcpy(data, asset.content.constData() + offset, number);
            offset += number;
            return number;
        } else {
            return -1;
        }
    };
    qint64 bytesAvailable() const override {
        qint64 bc = QIODevice::bytesAvailable() + asset.content.size() - offset;
        return bc;
    }
public slots:
    void error404() {
        errorOccurred(ContentNotFoundError);
    };
};

class WebNetworkAccessManager final: public QNetworkAccessManager {
    Q_OBJECT
    QNetworkAccessManager* existing;
    WebAssetStore* store;
public:
    WebNetworkAccessManager(QNetworkAccessManager* existing, WebAssetStore* store, QObject *parent) : QNetworkAccessManager(parent), existing(existing), store(store) {
        setCache(existing->cache());
        setCookieJar(existing->cookieJar());
        setProxy(existing->proxy());
        setProxyFactory(existing->proxyFactory());
    }
    QNetworkReply* createRequest(Operation operation, const QNetworkRequest &request, QIODevice *device) override {
        QUrl url = request.url();
        if (url.scheme() == "asset") {
            QString path_base64 = url.path();
            QString path = DecodeBase64(path_base64);
            // qDebug() << "[asset] request =" << path;
            return new WebAssetReply(store, path);
        } else {
            // qDebug() << "[other] request =" << url.toString();
            return QNetworkAccessManager::createRequest(operation, request, device);
        }
    }
};

class WebView final: public QWebView {
    Q_OBJECT
public:
    WebView(QWidget* parent): QWebView(parent) {
        initAssetStore();
        initBridge();
        setContextMenuPolicy(Qt::NoContextMenu);
        setUrl(QUrl(WebUiHtmlUrl));
        if (DebugEnabled()) {
            #ifndef _WIN32
            // inspector crashes on windows
            openInspector();
            #endif
        }
    }
    ~WebView() {}
    WebBridge* getBridge() {
        return bridge;
    }
    WebAssetStore* getStore() {
        return store;
    }
signals:
    void loadFinished();
    void eventEmitted();
private slots:
    void bridgeLoaded() {
        syncRootFontSizeWithScreenSize();
        connect(bridge, &WebBridge::EmitEvent, this, &WebView::emitEvent);
        emit loadFinished();
    }
    void emitEvent(QString handler, QVariantMap event) {
        emittedEventHandler = handler;
        emittedEventPayload = event;
        emit eventEmitted();
        emittedEventHandler = "";
        emittedEventPayload = QVariantMap();
        // if (debug) {
        //     qDebug() << "Event: " << handler;
        // }
    }
private:
    QString emittedEventHandler;
    QVariantMap emittedEventPayload;
public:
    QString getEmittedEventHandler() const { return emittedEventHandler; }
    QVariantMap getEmittedEventPayload() const { return emittedEventPayload; }
private:
    WebAssetStore* store = nullptr;
    void initAssetStore() {
        store = new WebAssetStore(this);
        auto old_nm = page()->networkAccessManager();
        auto new_nm = new WebNetworkAccessManager(old_nm, store, this);
        page()->setNetworkAccessManager(new_nm);
    };
private:
    WebBridge* bridge;
    void initBridge() {
        bridge = new WebBridge(this);
        connect(bridge, &WebBridge::LoadFinish, this, &WebView::bridgeLoaded);
    }
    void syncRootFontSizeWithScreenSize() {
        QScreen *screen = QGuiApplication::primaryScreen();
        connect(screen, &QScreen::geometryChanged, [this] ()->void {
            int fontSize = Get1remPixels();
            bridge->UpdateRootFontSize(double(fontSize));
        });
    }
private:
    QWebInspector* inspector = nullptr;
    void openInspector() {
        if (inspector != nullptr) {
            return;
        }
        inspector = new QWebInspector(this);
        page()->settings()->setAttribute(QWebSettings::DeveloperExtrasEnabled, true);
        inspector->setPage(page());
        QDialog* inspector_dialog = new QDialog(this);
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

#endif  // WEB_HPP

