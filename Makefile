.PHONY: runtime compiler dist

default: dev

K_COMPILER_BIN = build/dev/compiler
K_RUNTIME_JS = build/dev/runtime.js

compiler:
	go build -o $(K_COMPILER_BIN) compiler/main.go

runtime: compiler
	./bundle.py runtime/runtime.js \
		> $(K_RUNTIME_JS)
	echo ";(function(Runtime) { let KumaChan = Runtime;" \
		>> $(K_RUNTIME_JS)
	$(K_COMPILER_BIN) module runtime/preload/preload.k \
		>> $(K_RUNTIME_JS)
	echo "})((typeof KumaChan != 'undefined')? KumaChan: module.exports)" \
		>> $(K_RUNTIME_JS)

dev: compiler runtime

dist: dev
	cp $(K_COMPILER_BIN) build/dist/kumachan-build
	terser -c -m -- $(K_RUNTIME_JS) > build/dist/kumachan-runtime.js

clean:
	rm build/dev/*
	rm build/dist/*
