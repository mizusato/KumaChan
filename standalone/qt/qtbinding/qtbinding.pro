CONFIG += c++11
HEADERS += qtbinding.h \
    util.hpp \
    web.hpp
SOURCES += util.cpp \
    api.cpp
TARGET = qtbinding
TEMPLATE = lib
QT += core widgets uitools webenginecore webenginewidgets

RESOURCES += \
    qtbinding.qrc

