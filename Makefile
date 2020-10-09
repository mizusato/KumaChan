ifdef OS	
	WINDOWS_WORKAROUND = cp qt/build/release/libqtbinding* qt/build/
	LIBBIN = qt/build/release/qtbinding.dll
else
	WINDOWS_WORKAROUND = $(NOOP)
	LIBBIN = qt/build/libqtbinding*
endif

.PHONY: check qt interpreter stdlib

default: all

check:
	@echo -e '\033[1mChecking for Qt...\033[0m'
	qmake -v
	@echo -e '\033[1mChecking for Go...\033[0m'
	go version

qt:
	@echo -e '\033[1mCompiling CGO Qt Binding...\033[0m'
	cd qt/build && qmake ../qtbinding/qtbinding.pro && $(MAKE)
	$(WINDOWS_WORKAROUND)
	cp -P $(LIBBIN) build/

stdlib:
	@echo -e '\033[1mCopying Standard Library Files...\033[0m'
	if [ -d build/stdlib ]; then rm -r build/stdlib; fi
	cp -rP stdlib build/
	rm build/stdlib/*.go

interpreter: qt stdlib
	@echo -e '\033[1mCompiling the Interpreter...\033[0m'
	go build -o ./build/kumachan main.go

all: check qt stdlib interpreter
