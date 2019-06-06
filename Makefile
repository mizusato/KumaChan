.PHONY: runtime compiler

all: runtime compiler

runtime: compiler
	runtime/build.py runtime/runtime.js > build/dev/runtime.js
	echo ";(function(Runtime) { let KumaChan = Runtime;" >> build/dev/runtime.js
	build/dev/compiler module runtime/preload/preload.k >> build/dev/runtime.js
	echo "})((typeof KumaChan != 'undefined')? KumaChan: module.exports)" >> build/dev/runtime.js

compiler:
	go build -o build/dev/compiler compiler/main.go

