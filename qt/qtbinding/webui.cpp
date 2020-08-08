#include "adapt.hpp"
#include "webui.hpp"
#include "qtbinding.h"


static WebUiWindow* window = nullptr;

void QtWebUiInit(QtString title) {
    if (window == nullptr) {
        window = new WebUiWindow(QtUnwrapString(title));
    };
}

void* QtWebUiGetWindow() {
    return (void*) (window);
}

void QtWebUiUpdateVDOM(QtWebUiNode root) {
    Node* ptr = (Node*) (root.ptr);
    window->updateVDOM(ptr);
}

QtWebUiNode QtWebUiNewNode(QtString tagName) {
    Node* ptr = new Node;
    ptr->tagName = QtUnwrapString(tagName);
    return { (void*) ptr };
};

void QtWebUiNodeAddStyle(QtWebUiNode node, QtString key, QtString value) {
    Node* ptr = (Node*) (node.ptr);
    ptr->style[QtUnwrapString(key)] = QtUnwrapString(value);
}

void QtWebUiNodeAddEvent(QtWebUiNode node, QtString name, QtBool prevent, QtBool stop, size_t handler) {
    Node* ptr = (Node*) (node.ptr);
    EventOptions* opts = new EventOptions(ptr);
    opts->prevent = prevent;
    opts->stop = stop;
    opts->handler = handler;
    ptr->events[QtUnwrapString(name)] = opts;
}

void QtWebUiNodeSetText(QtWebUiNode node, QtString text) {
    Node* ptr = (Node*) (node.ptr);
    assert(ptr->children.length() == 0);
    ptr->is_text = true;
    ptr->text = QtUnwrapString(text);
}

void QtWebUiNodeAppendChild(QtWebUiNode node, QtWebUiNode child) {
    Node* ptr = (Node*) (node.ptr);
    assert(!(ptr->is_text));
    Node* child_ptr = (Node*) (child.ptr);
    ptr->children.append(child_ptr);
}
