CONFIG += c++11
HEADERS += qtbinding.hpp \
    adapt.hpp \
    vdom.hpp \
    webui.hpp
SOURCES += qtbinding.cpp \
    adapt.cpp \
    vdom.cpp \
    webui.cpp
TARGET = qtbinding
TEMPLATE = lib
QT += core widgets uitools webkitwidgets

RESOURCES += \
    qtbinding.qrc

