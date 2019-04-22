.PHONY: runtime compiler

all: runtime compiler

runtime:
	runtime/build.py runtime/runtime.js > build/dev/runtime.js

compiler:
	go build -o build/dev/compiler compiler/main.go

