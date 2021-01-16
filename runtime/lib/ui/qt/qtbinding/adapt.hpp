#ifndef ADAPT_HPP
#define ADAPT_HPP

#include <QString>
#include "qtbinding.h"

QtString QtWrapString(QString str);
QString QtUnwrapString(QtString str);

#endif // ADAPT_HPP
