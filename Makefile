.PHONY: runtime transpiler

all: runtime transpiler

runtime:
	runtime/build.py runtime/runtime.js > build/dev/runtime.js

transpiler:
	go build -o build/dev/transpiler transpiler/transpiler.go

