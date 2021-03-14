#include <QGuiApplication>
#include <QScreen>
#include "util.hpp"


QtString WrapString(QString str) {
    QString* ptr = new QString;
    *ptr = str;
    return { (void*) ptr };
}
QString UnwrapString(QtString str) {
    return QString(*(QString*)(str.ptr));
}

QString EncodeBase64(QString str) {
    return QString::fromUtf8(str.toUtf8().toBase64(QByteArray::Base64UrlEncoding));
}
QString DecodeBase64(QString str) {
    return QString::fromUtf8(QByteArray::fromBase64(str.toUtf8(), QByteArray::Base64UrlEncoding));
}

int Get1remPixels() {
    QScreen *screen = QGuiApplication::primaryScreen();
    QRect screenGeometry = screen->geometry();
    int screenHeight = screenGeometry.height();
    int screenWidth = screenGeometry.width();
    int minEdgeLength = std::min(screenHeight, screenWidth);
    return int(round(BaseSize * (((double) minEdgeLength) / BaseScreen)));
}
QSize GetSizeFromRelative(QSize size_rem) {
    int unit = Get1remPixels();
    return QSize((unit * size_rem.width()), (unit * size_rem.height()));
}

bool debugEnabled
    = false;
bool DebugEnabled() {
    return debugEnabled;
}
void EnableDebug() {
    debugEnabled = true;
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

