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
	cp -P qt/build/libqtbinding* build/

stdlib:
	@echo -e '\033[1mCopying Standard Library Files...\033[0m'
	if [ -d build/stdlib ]; then rm -r build/stdlib; fi
	cp -rP stdlib build/
	rm build/stdlib/*.go

interpreter: qt stdlib
	@echo -e '\033[1mCompiling the Interpreter...\033[0m'
	go build -o ./build/kumachan main.go

all: check qt stdlib interpreter
