#include <QApplication>
#include <QDesktopWidget>
#include "adapt.hpp"
#include "webui.hpp"
#include "qtbinding.h"


static WebUiWindow* window = nullptr;

void WebUiInit(QtString title) {
    if (window == nullptr) {
        window = new WebUiWindow(QtUnwrapString(title));
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

QtString WebUiGetEventHandler() {
    return QtWrapString(window->getEmittedEventHandler());
}

QtVariantMap WebUiGetEventPayload() {
    QVariantMap* m = new QVariantMap;
    *m = window->getEmittedEventPayload();
    return { m };
}

void WebUiCallMethod(QtString id, QtString name, QtVariantList args) {
    QVariantList args_copy = *(QVariantList*)(args.ptr);
    window->bridge->CallMethod(QtUnwrapString(id), QtUnwrapString(name), args_copy);
}

void WebUiEraseStyle(QtString id, QtString key) {
    window->bridge->EraseStyle(QtUnwrapString(id), QtUnwrapString(key));
}

void WebUiApplyStyle(QtString id, QtString key, QtString value) {
    window->bridge->ApplyStyle(QtUnwrapString(id), QtUnwrapString(key), QtUnwrapString(value));
}

void WebUiDetachEvent(QtString id, QtString event) {
    window->bridge->DetachEvent(QtUnwrapString(id), QtUnwrapString(event));
}

void WebUiModifyEvent(QtString id, QtString event, QtBool prevent, QtBool stop) {
    window->bridge->ModifyEvent(QtUnwrapString(id), QtUnwrapString(event), bool(prevent), bool(stop));
}

void WebUiAttachEvent(QtString id, QtString event, QtBool prevent, QtBool stop, QtString handler) {
    window->bridge->AttachEvent(QtUnwrapString(id), QtUnwrapString(event), bool(prevent), bool(stop), QtUnwrapString(handler));
}

void WebUiSetText(QtString id, QtString text) {
    window->bridge->SetText(QtUnwrapString(id), QtUnwrapString(text));
}

// void WebUiInsertNode(QtString parent, QtString ref, QtString id, QtString tag) {
//     window->bridge->InsertNode(QtUnwrapString(parent), QtUnwrapString(ref), QtUnwrapString(id), QtUnwrapString(tag));
// }

void WebUiAppendNode(QtString parent, QtString id, QtString tag) {
    window->bridge->AppendNode(QtUnwrapString(parent), QtUnwrapString(id), QtUnwrapString(tag));
}

void WebUiRemoveNode(QtString parent, QtString id) {
    window->bridge->RemoveNode(QtUnwrapString(parent), QtUnwrapString(id));
}

void WebUiUpdateNode(QtString old_id, QtString new_id) {
    window->bridge->UpdateNode(QtUnwrapString(old_id), QtUnwrapString(new_id));
}

void WebUiReplaceNode(QtString parent, QtString old_id, QtString id, QtString tag) {
    window->bridge->ReplaceNode(QtUnwrapString(parent), QtUnwrapString(old_id), QtUnwrapString(id), QtUnwrapString(tag));
}
