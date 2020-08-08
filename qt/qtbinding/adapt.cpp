#include "adapt.hpp"


QtString QtWrapString(QString str) {
    QString* ptr = new QString;
    *ptr = str;
    return { (void*) ptr };
}

QString QtUnwrapString(QtString str) {
    return QString(*(QString*)(str.ptr));
}
