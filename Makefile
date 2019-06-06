.PHONY: runtime compiler

all: runtime compiler

runtime: compiler
	runtime/build.py runtime/runtime.js > build/dev/runtime.js
	echo "let KumaChan = module.exports" >> build/dev/runtime.js
	build/dev/compiler module runtime/preload/preload.k >> build/dev/runtime.js

compiler:
	go build -o build/dev/compiler compiler/main.go

