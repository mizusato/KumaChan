CONFIG += c++11
HEADERS += qtbinding.hpp \
    adapt.hpp \
    webui.hpp
SOURCES += qtbinding.cpp \
    adapt.cpp \
    webui.cpp
TARGET = qtbinding
TEMPLATE = lib
QT += core widgets uitools webkitwidgets

RESOURCES += \
    qtbinding.qrc

