#include <QApplication>
#include <QScreen>
#include <QMetaMethod>
#include <QDir>
#include <QFileDialog>
#include <QUiLoader>
#include <QBuffer>
#include <QByteArray>
#include <QString>
#include <QVector>
#include <QVariant>
#include <QResizeEvent>
#include <QAction>
#include <QIcon>
#include <QPixmap>
#include <QListWidget>
#include <QListWidgetItem>
#include <QVariantMap>
#include <QVariantList>
#include <QWebView>
#include <QWebPage>
#include <QWebFrame>
#include "adapt.hpp"
#include "qtbinding.hpp"
#include "qtbinding.h"


const size_t QtEventMove = QEvent::Move;
const size_t QtEventResize = QEvent::Resize;
const size_t QtEventClose = QEvent::Close;

QMetaObject::Connection QtDynamicConnect (
        QObject* emitter , const QString& signalName,
        QObject* receiver, const QString& slotName
);

struct __QtConnHandle {
    QMetaObject::Connection
        conn;
    CallbackObject*
        cb_obj;
};

static QApplication*
    app = nullptr;
static Bridge*
    bridge = nullptr;
static QUiLoader*
    loader = nullptr;
static bool
    initialized = false;

void QtInit() {
    static int fake_argc = 1;
    static char fake_arg[] = {'Q','t','A','p','p','\0'};
    static char* fake_argv[] = { fake_arg };
    if (!(initialized)) {
        QCoreApplication::setAttribute(Qt::AA_ShareOpenGLContexts);
        app = new QApplication(fake_argc, fake_argv);
        app->setQuitOnLastWindowClosed(false);
        bridge = new Bridge();
        loader = new QUiLoader();
        qRegisterMetaType<callback_t>();
        initialized = true;
    }
}

int QtMain() {
    return app->exec();
}

void QtCommitTask(callback_t cb, size_t payload) {
    bridge->QueueCallback(cb, payload);
}

void QtExit(int code) {
    app->exit(code);
}

void QtQuit() {
    app->quit();
}

void* QtLoadWidget(const char* definition, const char* directory) {
    QByteArray bytes(definition);
    QBuffer buf(&bytes);
    QDir dir(directory);
    loader->setWorkingDirectory(dir);
    QWidget* widget = loader->load(&buf, nullptr);
    return (void*) widget;
}

void* QtWidgetFindChild(void* widget_ptr, const char* name) {
    QWidget* widget = (QWidget*) widget_ptr;
    QWidget* child = widget->findChild<QWidget*>(QString(name));
    return (void*) child;
}

void* QtWidgetFindChildAction(void* widget_ptr, const char* name) {
    QWidget* widget = (QWidget*) widget_ptr;
    QAction* child = widget->findChild<QAction*>(QString(name));
    return (void*) child;
}

void QtWidgetShow(void* widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->show();
}

void QtWidgetHide(void* widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    widget->hide();
}

void QtWidgetMoveToScreenCenter(void* widget_ptr) {
    QWidget* widget = (QWidget*) widget_ptr;
    QScreen* screen = QGuiApplication::primaryScreen();
    widget->move(widget->pos() + (screen->geometry().center() - widget->geometry().center()));
}

void QtDialogExec(void *dialog_ptr) {
    QDialog* dialog = (QDialog*) dialog_ptr;
    dialog->exec();
}

void QtDialogAccept(void *dialog_ptr) {
    QDialog* dialog = (QDialog*) dialog_ptr;
    dialog->accept();
}

void QtDialogReject(void *dialog_ptr) {
    QDialog* dialog = (QDialog*) dialog_ptr;
    dialog->reject();
}

QtBool QtObjectSetPropBool(void* obj_ptr, const char* prop, QtBool val) {
    QObject* obj = (QObject*) obj_ptr;
    return obj->setProperty(prop, (val != 0));
}

QtBool QtObjectGetPropBool(void* obj_ptr, const char* prop) {
    QObject* obj = (QObject*) obj_ptr;
    QVariant val = obj->property(prop);
    return val.toBool();
}

QtBool QtObjectSetPropString(void* obj_ptr, const char* prop, QtString val) {
    QObject* obj = (QObject*) obj_ptr;
    return obj->setProperty(prop, QtUnwrapString(val));
}

QtString QtObjectGetPropString(void* obj_ptr, const char* prop) {
    QObject* obj = (QObject*) obj_ptr;
    QVariant val = obj->property(prop);
    return QtWrapString(val.toString());
}

QtBool QtObjectSetPropInt(void* obj_ptr, const char* prop, int val) {
    QObject* obj = (QObject*) obj_ptr;
    return obj->setProperty(prop, val);
}

int QtObjectGetPropInt(void* obj_ptr, const char* prop) {
    QObject* obj = (QObject*) obj_ptr;
    QVariant val = obj->property(prop);
    return val.toInt();
}

QtBool QtObjectSetPropPoint(void* obj_ptr, const char* prop, QtPoint val) {
    QObject* obj = (QObject*) obj_ptr;
    QPoint p = QPoint(val.x, val.y);
    return obj->setProperty(prop, p);
}

QtPoint QtObjectGetPropPoint(void* obj_ptr, const char* prop) {
    QObject* obj = (QObject*) obj_ptr;
    QVariant val = obj->property(prop);
    QPoint point = val.toPoint();
    return { point.x(), point.y() };
}

QtConnHandle QtConnect (
        void* obj_ptr,
        const char* signal,
        callback_t cb,
        size_t payload
) {
    QObject* target_obj = (QObject*) obj_ptr;
    CallbackObject* cb_obj = new CallbackObject(cb, payload);
    __QtConnHandle* handle = new __QtConnHandle;
    handle->conn = QtDynamicConnect(target_obj, signal, cb_obj, "slot()");
    handle->cb_obj = cb_obj;
    return { (void*) handle };
}

QtBool QtIsConnectionValid(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    return bool(h->conn);
};

void QtDisconnect(QtConnHandle handle) {
    __QtConnHandle* h = (__QtConnHandle*) handle.ptr;
    if (!(h->conn)) {
        QObject::disconnect(h->conn);
    }
    delete h->cb_obj;
    h->cb_obj = nullptr;
    delete h;
};

void QtBlockSignals(void* obj_ptr, QtBool block) {
    QObject* obj = (QObject*) obj_ptr;
    obj->blockSignals(bool(block));
}

QtEventListener QtAddEventListener (
        void*       obj_ptr,
        size_t      kind,
        QtBool      prevent,
        callback_t  cb,
        size_t      payload
) {
    QObject* obj = (QObject*) obj_ptr;
    QEvent::Type q_kind = (QEvent::Type) kind;
    EventListener* l = new EventListener(q_kind, prevent, cb, payload);
    obj->installEventFilter(l);
    QtEventListener wrapped;
    wrapped.ptr = l;
    return wrapped;
}

QtEvent QtGetCurrentEvent(QtEventListener listener) {
    EventListener* l = (EventListener*) listener.ptr;
    QtEvent wrapped;
    wrapped.ptr = l->current_event;
    return wrapped;
}

void QtRemoveEventListener(void* obj_ptr, QtEventListener listener) {
    QObject* obj = (QObject*) obj_ptr;
    EventListener* l = (EventListener*) listener.ptr;
    obj->removeEventFilter(l);
    delete l;
}

size_t QtResizeEventGetWidth(QtEvent ev) {
    QResizeEvent* resize = (QResizeEvent*) ev.ptr;
    return resize->size().width();
}

size_t QtResizeEventGetHeight(QtEvent ev) {
    QResizeEvent* resize = (QResizeEvent*) ev.ptr;
    return resize->size().height();
}

QtPoint QtMakePoint(int x, int y) {
    return { x, y };
}

int QtPointGetX(QtPoint p) {
    return p.x;
}

int QtPointGetY(QtPoint p) {
    return p.y;
}

QtString QtNewStringUTF8(const uint8_t* buf, size_t len) {
    QString* ptr = new QString;
    *ptr = QString::fromUtf8((const char*)(buf), len);
    return { (void*) ptr };
}

QtString QtNewStringUTF32(const uint32_t* buf, size_t len) {
    QString* ptr = new QString;
    *ptr = QString::fromUcs4(buf, len);
    return { (void*) ptr };
}

void QtDeleteString(QtString str) {
    delete (QString*)(str.ptr);
}

size_t QtStringUTF16Length(QtString str) {
    return QtUnwrapString(str).length();
}

size_t QtStringWriteToUTF32Buffer(QtString str, uint32_t *buf) {
    QVector<uint> vec = QtUnwrapString(str).toUcs4();
    size_t len = 0;
    for(auto rune: vec) {
        *buf = rune;
        buf += 1;
        len += 1;
    }
    return len;
}

size_t QtStringListGetSize(QtStringList list) {
    QStringList* ptr = (QStringList*) (list.ptr);
    return ptr->size();
}

QtString QtStringListGetItem(QtStringList list, size_t index) {
    QStringList* ptr = (QStringList*) (list.ptr);
    return QtWrapString(ptr->at(index));
}

void QtDeleteStringList(QtStringList list) {
    delete (QStringList*)(list.ptr);
}

QtVariantList QtNewVariantList() {
    return { (void*) new QVariantList() };
}

void QtVariantListAppendNumber(QtVariantList l, double n) {
    QVariantList* ptr = (QVariantList*) l.ptr;
    ptr->append(n);
}

void QtVariantListAppendString(QtVariantList l, QtString str) {
    QVariantList* ptr = (QVariantList*) l.ptr;
    ptr->append(QtUnwrapString(str));
}

void QtDeleteVariantList(QtVariantList l) {
    delete (QVariantList*)(l.ptr);
}

QtString QtVariantMapGetString(QtVariantMap m, QtString key) {
    QVariantMap* ptr = (QVariantMap*) m.ptr;
    QString key_ = QtUnwrapString(key);
    QVariant val_ = (*ptr)[key_];
    QtString val = QtWrapString(val_.toString());
    return val;
}

double QtVariantMapGetFloat(QtVariantMap m, QtString key) {
    QVariantMap* ptr = (QVariantMap*) m.ptr;
    QString key_ = QtUnwrapString(key);
    QVariant val_ = (*ptr)[key_];
    double val = val_.toDouble();
    return val;
}

QtBool QtVariantMapGetBool(QtVariantMap m, QtString key) {
    QVariantMap* ptr = (QVariantMap*) m.ptr;
    QString key_ = QtUnwrapString(key);
    QVariant val_ = (*ptr)[key_];
    int val = val_.toBool();
    return val;
}

void QtDeleteVariantMap(QtVariantMap m) {
    delete (QVariantMap*)(m.ptr);
}

QtIcon QtNewIcon(QtPixmap pm) {
    QIcon* ptr = new QIcon(*(QPixmap*)(pm.ptr));
    return { (void*) ptr };
}

void QtDeleteIcon(QtIcon icon) {
    delete (QIcon*)(icon.ptr);
}

QtPixmap QtNewPixmap(const uint8_t* buf, size_t len, const char* format) {
    QPixmap *ptr = new QPixmap;
    ptr->loadFromData(buf, len, format);
    return { (void*) ptr };
}

QtPixmap QtNewPixmapPNG(const uint8_t* buf, size_t len) {
    return QtNewPixmap(buf, len, "PNG");
}

QtPixmap QtNewPixmapJPEG(const uint8_t* buf, size_t len) {
    return QtNewPixmap(buf, len, "JPG");
}

void QtDeletePixmap(QtPixmap pm) {
    delete (QPixmap*)(pm.ptr);
}

void QtListWidgetClear(void* widget_ptr) {
    QListWidget* widget = (QListWidget*) widget_ptr;
    widget->clear();
}

void QtListWidgetAddItem(void* widget_ptr, QtString key_, QtString label_, QtBool as_current) {
    QListWidget* widget = (QListWidget*) widget_ptr;
    QString key = QtUnwrapString(key_);
    QString label = QtUnwrapString(label_);
    QListWidgetItem* item = new QListWidgetItem(label, widget);
    item->setData(Qt::UserRole, key);
    widget->addItem(item);
    if (as_current) {
        widget->setCurrentItem(item);
    }
}

void QtListWidgetAddItemWithIcon(void* widget_ptr, QtString key_, QtIcon icon_, QtString label_, QtBool as_current) {
    QListWidget* widget = (QListWidget*) widget_ptr;
    QString key = QtUnwrapString(key_);
    QString label = QtUnwrapString(label_);
    QIcon* icon = (QIcon*) icon_.ptr;
    QListWidgetItem* item = new QListWidgetItem(*icon, label, widget);
    item->setData(Qt::UserRole, key);
    widget->addItem(item);
    if (as_current) {
        widget->setCurrentItem(item);
    }
}

QtBool QtListWidgetHasCurrentItem(void* widget_ptr) {
    QListWidget* widget = (QListWidget*) widget_ptr;
    return (widget->currentRow() != -1);
}

QtString QtListWidgetGetCurrentItemKey(void* widget_ptr) {
    QListWidget* widget = (QListWidget*) widget_ptr;
    QVariant key_v = widget->currentItem()->data(Qt::UserRole);
    return QtWrapString(key_v.toString());
}

void QtWebViewDisableContextMenu(void* widget_ptr) {
    QWebView* widget = (QWebView*) widget_ptr;
    widget->setContextMenuPolicy(Qt::NoContextMenu);
}

void QtWebViewEnableLinkDelegation(void* widget_ptr) {
    QWebView* widget = (QWebView*) widget_ptr;
    widget->page()->setLinkDelegationPolicy(QWebPage::DelegateAllLinks);
}

void QtWebViewRecordClickedLink(void* widget_ptr) {
    QWebView* widget = (QWebView*) widget_ptr;
    QObject::connect(widget, &QWebView::linkClicked, [widget](const QUrl& url)->void {
        widget->setProperty("qtbindingClickedLinkUrl", url.toString());
    });
}

void QtWebViewSetHTML(void* widget_ptr, QtString html) {
    QWebView* widget = (QWebView*) widget_ptr;
    widget->page()->mainFrame()->setHtml(QtUnwrapString(html));
}

void QtWebViewScrollToAnchor(void* widget_ptr, QtString anchor) {
    QWebView* widget = (QWebView*) widget_ptr;
    widget->page()->mainFrame()->scrollToAnchor(QtUnwrapString(anchor));
}

QtPoint QtWebViewGetScroll(void* widget_ptr) {
    QWebView* widget = (QWebView*) widget_ptr;
    return QtObjectGetPropPoint(widget->page()->mainFrame(), "scrollPosition");
}

void QtWebViewSetScroll(void* widget_ptr, QtPoint pos) {
    QWebView* widget = (QWebView*) widget_ptr;
    QtObjectSetPropPoint(widget->page()->mainFrame(), "scrollPosition", pos);
}

QtString QtFileDialogOpen(void* parent_ptr, QtString title, QtString cwd, QtString filter) {
   QWidget* parent = (QWidget*) parent_ptr;
   QString path = QFileDialog::getOpenFileName(
               parent,
               QtUnwrapString(title),
               QtUnwrapString(cwd),
               QtUnwrapString(filter));
   return QtWrapString(path);
}

QtStringList QtFileDialogOpenMultiple(void* parent_ptr,  QtString title, QtString cwd, QtString filter) {
    QWidget* parent = (QWidget*) parent_ptr;
    QStringList* path_list = new QStringList;
    *path_list = QFileDialog::getOpenFileNames(
                parent,
                QtUnwrapString(title),
                QtUnwrapString(cwd),
                QtUnwrapString(filter));
    QtStringList wrapped = { path_list };
    return wrapped;
}

QtString QtFileDialogSelectDirectory(void *parent_ptr, QtString title, QtString cwd) {
    QWidget* parent = (QWidget*) parent_ptr;
    QString path = QFileDialog::getExistingDirectory(
                parent,
                QtUnwrapString(title),
                QtUnwrapString(cwd));
    return QtWrapString(path);
}

QtString QtFileDialogSave(void *parent_ptr, QtString title, QtString cwd, QtString filter) {
    QWidget* parent = (QWidget*) parent_ptr;
    QString path = QFileDialog::getSaveFileName(
                parent,
                QtUnwrapString(title),
                QtUnwrapString(cwd),
                QtUnwrapString(filter));
    return QtWrapString(path);
}

QMetaObject::Connection QtDynamicConnect (
        QObject* emitter , const QString& signalName,
        QObject* receiver, const QString& slotName
) {
    /* ref: https://stackoverflow.com/questions/26208851/qt-connecting-signals-and-slots-from-text */
    int index = emitter->metaObject()
                ->indexOfSignal(QMetaObject::normalizedSignature(qPrintable(signalName)));
    if (index == -1) {
        qWarning("Wrong signal name: %s", qPrintable(signalName));
        return QMetaObject::Connection();
    }
    QMetaMethod signal = emitter->metaObject()->method(index);
    index = receiver->metaObject()
            ->indexOfSlot(QMetaObject::normalizedSignature(qPrintable(slotName)));
    if (index == -1) {
        qWarning("Wrong slot name: %s", qPrintable(slotName));
        return QMetaObject::Connection();
    }
    QMetaMethod slot = receiver->metaObject()->method(index);
    return QObject::connect(emitter, signal, receiver, slot);
}
