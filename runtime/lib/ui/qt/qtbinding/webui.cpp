#include <QApplication>
#include <QDesktopWidget>
#include <QUuid>
#include "adapt.hpp"
#include "webui.hpp"
#include "qtbinding.h"


static WebUiWindow* window = nullptr;

void WebUiInit(QtString title, QtBool debug) {
    if (window == nullptr) {
        window = new WebUiWindow(QtUnwrapString(title), bool(debug));
    };
}
void WebUiLoadView() {
    window->loadView();
    window->move(QApplication::desktop()->rect().center()
                 - window->frameGeometry().center());
}
void* WebUiGetWindow() {
    return (void*) (window);
}

void WebUiRegisterAsset(QtString path, QtString mime, const uint8_t* buf, size_t len) {
    QByteArray data = QByteArray::fromRawData((const char*)(buf), len);
    window->store->InsertItem(QtUnwrapString(path), QtUnwrapString(mime), data);
}
QtString WebUiInjectCSS(QtString path) {
    QString uuid = QUuid::createUuid().toString();
    QString path_base64 = QtEncodeBase64(QtUnwrapString(path));
    emit window->bridge->InjectCSS(uuid, path_base64);
    return QtWrapString(uuid);
}
QtString WebUiInjectJS(QtString path) {
    QString uuid = QUuid::createUuid().toString();
    QString path_base64 = QtEncodeBase64(QtUnwrapString(path));
    emit window->bridge->InjectJS(uuid, path_base64);
    return QtWrapString(uuid);
}
QtString WebUiInjectTTF(QtString path, QtString family, QtString weight, QtString style) {
    QString uuid = QUuid::createUuid().toString();
    QString path_base64 = QtEncodeBase64(QtUnwrapString(path));
    emit window->bridge->InjectTTF(uuid, path_base64, QtUnwrapString(family), QtUnwrapString(weight), QtUnwrapString(style));
    return QtWrapString(uuid);
}
void WebUiCallMethod(QtString id, QtString name, QtVariantList args) {
    QVariantList args_copy = *(QVariantList*)(args.ptr);
    emit window->bridge->CallMethod(QtUnwrapString(id), QtUnwrapString(name), args_copy);
}

QtString WebUiGetCurrentEventHandler() {
    return QtWrapString(window->getEmittedEventHandler());
}
QtVariantMap WebUiGetCurrentEventPayload() {
    QVariantMap* m = new QVariantMap;
    *m = window->getEmittedEventPayload();
    return { m };
}
void WebUiPatchActualDOM(QtString operations) {
    emit window->bridge->PatchActualDOM(QtUnwrapString(operations));
}

