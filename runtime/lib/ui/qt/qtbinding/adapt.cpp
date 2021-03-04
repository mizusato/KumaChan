#include <QByteArray>
#include "adapt.hpp"


QtString QtWrapString(QString str) {
    QString* ptr = new QString;
    *ptr = str;
    return { (void*) ptr };
}

QString QtUnwrapString(QtString str) {
    return QString(*(QString*)(str.ptr));
}

QString QtEncodeBase64(QString str) {
    return QString::fromUtf8(str.toUtf8().toBase64(QByteArray::Base64UrlEncoding));
}

QString QtDecodeBase64(QString str) {
    return QString::fromUtf8(QByteArray::fromBase64(str.toUtf8(), QByteArray::Base64UrlEncoding));
}

