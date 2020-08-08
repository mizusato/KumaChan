#include "adapt.hpp"
#include "webui.hpp"
#include "qtbinding.h"


static WebUiWindow* window = nullptr;

void QtWebUiInit(QtString title) {
    if (window == nullptr) {
        window = new WebUiWindow(QtUnwrapString(title));
    };
}
