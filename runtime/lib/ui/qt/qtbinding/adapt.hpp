#ifndef ADAPT_HPP
#define ADAPT_HPP

#include <QString>
#include "qtbinding.h"

QtString QtWrapString(QString str);
QString QtUnwrapString(QtString str);
QString QtEncodeBase64(QString str);
QString QtDecodeBase64(QString str);

#endif // ADAPT_HPP
