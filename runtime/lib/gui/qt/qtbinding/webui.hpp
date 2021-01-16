#ifndef WEBUI_HPP
#define WEBUI_HPP

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


#define WebUiHtmlUrl "qrc:/qtbinding/webui/webui.html"
#define BaseSize 18.0
#define BaseScreen 768.0

class WebUiBridge final: public QObject {
    Q_OBJECT
public:
    WebUiBridge(QObject* parent): QObject(parent) {};
signals:
    void LoadFinish();
    void EmitEvent(QString handler, QVariantMap event);
    void UpdateRootFontSize(double size);
    void InjectCSS(QString uuid, QString path);
    void InjectJS(QString uuid, QString path);
    void InjectTTF(QString uuid, QString path, QString family, QString weight, QString style);
    void CallMethod(QString id, QString name, QVariantList args);
    void EraseStyle(QString id, QString key);
    void ApplyStyle(QString id, QString key, QString value);
    void RemoveAttr(QString id, QString name);
    void SetAttr(QString id, QString name, QString value);
    void DetachEvent(QString id, QString event);
    void ModifyEvent(QString id, QString event, bool prevent, bool stop, bool capture);
    void AttachEvent(QString id, QString event, bool prevent, bool stop, bool capture, QString handler);
    void SetText(QString id, QString text);
    void AppendNode(QString parent, QString id, QString tag);
    void RemoveNode(QString parent, QString id);
    void UpdateNode(QString old_id, QString new_id);
    void ReplaceNode(QString parent, QString old_id, QString id, QString tag);
    void SwapNode(QString id, QString a, QString b);
    void PerformActualRendering();
};

struct WebUiAsset {
    bool        exists;
    QString     mime;
    QByteArray  content;
};

class WebUiAssetStore final: public QObject {
    Q_OBJECT
    QHash<QString,WebUiAsset> mapping;
public:
    WebUiAssetStore(QObject* parent): QObject(parent) {};
    void InsertItem(QString path, QString mime, QByteArray data) {
        // qDebug() << "register" << path << "(" << data.size() << ")";
        mapping[path] = { true, mime, data };
    };
    WebUiAsset LookupItem(QString path) {
        return mapping.value(path);
    };
};

class WebUiAssetReply final: public QNetworkReply {
    Q_OBJECT
    WebUiAsset asset;
    qint64 offset;
public:
    WebUiAssetReply(WebUiAssetStore* store, QString path): QNetworkReply() {
        offset = 0;
        asset = store->LookupItem(path);
        if (asset.exists) {
            open(ReadOnly | Unbuffered);
            setHeader(QNetworkRequest::ContentTypeHeader, QVariant(asset.mime));
            setHeader(QNetworkRequest::ContentLengthHeader, asset.content.size());
            metaDataChanged();
            QTimer::singleShot(0, this, &WebUiAssetReply::metaDataChanged);
            QTimer::singleShot(0, this, &WebUiAssetReply::readyRead);
            QTimer::singleShot(0, this, &WebUiAssetReply::finished);
        } else {
            qDebug() << "[WebUi] asset file not found:" << path;
            setError(ContentNotFoundError, "asset file not found");
            QTimer::singleShot(0, this, &WebUiAssetReply::error404);
            QTimer::singleShot(0, this, &WebUiAssetReply::finished);
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

class WebUiNetworkAccessManager final: public QNetworkAccessManager {
    Q_OBJECT
    QNetworkAccessManager* existing;
    WebUiAssetStore* store;
public:
    WebUiNetworkAccessManager(QNetworkAccessManager* existing, WebUiAssetStore* store, QObject *parent) : QNetworkAccessManager(parent), existing(existing), store(store) {
        setCache(existing->cache());
        setCookieJar(existing->cookieJar());
        setProxy(existing->proxy());
        setProxyFactory(existing->proxyFactory());
    }
    QNetworkReply* createRequest(Operation operation, const QNetworkRequest &request, QIODevice *device) override {
        QUrl url = request.url();
        if (url.scheme() == "asset") {
            QString path = url.path();
            // qDebug() << "[asset] request =" << path;
            return new WebUiAssetReply(store, path);
        } else {
            // qDebug() << "[other] request =" << url.toString();
            return QNetworkAccessManager::createRequest(operation, request, device);
        }
    }
};

class WebUiWindow final: public QMainWindow {
    Q_OBJECT
private:
    QWebView* view;
    QWebPage* page;
    QWebFrame* frame;
    bool debug;
public:
    WebUiAssetStore* store;
    WebUiBridge* bridge;
    WebUiWindow(QString title): QMainWindow(nullptr), view(nullptr), debug(true) {
        setWindowTitle(title);
        store = new WebUiAssetStore(this);
        bridge = new WebUiBridge(this);
        connect(bridge, &WebUiBridge::EmitEvent, this, &WebUiWindow::emitEvent);
        connect(bridge, &WebUiBridge::LoadFinish, this, &WebUiWindow::loadFinished);
        QScreen *screen = QGuiApplication::primaryScreen();
        connect(screen, &QScreen::geometryChanged, this, &WebUiWindow::updateRootFontSize);
        connect(bridge, &WebUiBridge::LoadFinish, this, &WebUiWindow::updateRootFontSize);
    }
    void loadView() {
        if (view != nullptr) { return; }
        view = new QWebView(this);
        page = view->page();
        replaceNetworkManager();
        view->setContextMenuPolicy(Qt::NoContextMenu);
        page->settings()->setAttribute(QWebSettings::DeveloperExtrasEnabled, true);
        frame = page->mainFrame();
        connect(frame, &QWebFrame::javaScriptWindowObjectCleared, [this] () -> void {
            frame->addToJavaScriptWindowObject("WebUI", bridge);
        });
        view->setUrl(QUrl(WebUiHtmlUrl));
        setCentralWidget(view);
        show();
        if (debug) {
            #ifndef _WIN32
            // inspector crashes on windows
            openInspector();
            #endif
        }
    }
    ~WebUiWindow() {}
private:
    void replaceNetworkManager() {
        auto old_nm = page->networkAccessManager();
        auto new_nm = new WebUiNetworkAccessManager(old_nm, store, this);
        page->setNetworkAccessManager(new_nm);
    };
    void closeEvent(QCloseEvent* ev) override {
        ev->ignore();
        closeButtonClicked();
    };
    QString emittedEventHandler;
    QVariantMap emittedEventPayload;
    size_t detachedHandler;
public:
    QString getEmittedEventHandler() const { return emittedEventHandler; }
    QVariantMap getEmittedEventPayload() const { return emittedEventPayload; }
    size_t getDetachedHandler() const { return detachedHandler; }
signals:
    void loadFinished();
    void eventEmitted();
    void closeButtonClicked();
public slots:
    void emitEvent(QString handler, QVariantMap event) {
        emittedEventHandler = handler;
        emittedEventPayload = event;
        eventEmitted();
        // if (debug) {
        //     qDebug() << "Event: " << handler;
        // }
    };
    void updateRootFontSize() {
        QScreen *screen = QGuiApplication::primaryScreen();
        QRect screenGeometry = screen->geometry();
        int screenHeight = screenGeometry.height();
        int screenWidth = screenGeometry.width();
        int minEdgeLength = std::min(screenHeight, screenWidth);
        double fontSize = round(BaseSize * (((double) minEdgeLength) / BaseScreen));
        bridge->UpdateRootFontSize(fontSize);
    }
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

