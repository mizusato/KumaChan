ifdef OS
	WINDOWS_WORKAROUND = cp runtime/lib/ui/qt/build/release/libqtbinding* runtime/lib/ui/qt/build/
	LIBBIN = runtime/lib/ui/qt/build/release/qtbinding.dll
	EXENAME = kumachan.exe
else
	WINDOWS_WORKAROUND = $(NOOP)
	LIBBIN = runtime/lib/ui/qt/build/libqtbinding*
	EXENAME = kumachan
endif

default: all

check:
	@echo -e '\033[1mChecking for Qt...\033[0m'
	qmake -v
	@echo -e '\033[1mChecking for Go...\033[0m'
	go version

qt:
	@echo -e '\033[1mCompiling CGO Qt Binding...\033[0m'
	cd runtime/lib/ui/qt/build && qmake ../qtbinding/qtbinding.pro && $(MAKE)
	$(WINDOWS_WORKAROUND)
	cp -P $(LIBBIN) build/

stdlib:
	@echo -e '\033[1mCopying Standard Library Files...\033[0m'
	if [ -d build/stdlib ]; then rm -r build/stdlib; fi
	cp -rP stdlib build/
	rm build/stdlib/*.go

resources:
	@echo -e '\033[1mCopying Resources Files...\033[0m'
	if [ -d build/resources ]; then rm -r build/resources; fi
	mkdir build/resources
	cp support/docs/api_doc.css build/resources/
	cp support/docs/api_browser.ui build/resources/
	cp -r support/docs/icons build/resources/

deps: check qt stdlib resources
	$(NOOP)

interpreter: deps
	@echo -e '\033[1mCompiling the Interpreter...\033[0m'
	go build -o ./build/$(EXENAME) main.go

.PHONY: check qt stdlib resources deps interpreter
all: interpreter

