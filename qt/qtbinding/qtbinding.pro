CONFIG += c++11
HEADERS += qtbinding.hpp \
    vdom.hpp \
    webui.hpp
SOURCES += qtbinding.cpp \
    vdom.cpp
TARGET = qtbinding
TEMPLATE = lib
QT += core widgets uitools webkitwidgets

