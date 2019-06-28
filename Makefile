.PHONY: runtime compiler test dist

default: dev

K_COMPILER_BIN = build/dev/compiler
K_RUNTIME_JS = build/dev/runtime.js

compiler:
	go build -o $(K_COMPILER_BIN) main.go

runtime: compiler
	./bundle.py runtime/runtime.js \
		> $(K_RUNTIME_JS)
	echo ";(function(Runtime) { let KumaChan = Runtime;" \
		>> $(K_RUNTIME_JS)
	$(K_COMPILER_BIN) module runtime/modules/Std/Std.k \
		>> $(K_RUNTIME_JS)
	echo "})((typeof KumaChan != 'undefined')? KumaChan: module.exports)" \
		>> $(K_RUNTIME_JS)

test: compiler runtime
	./test.js

dev: compiler runtime test

dist: dev
	cp $(K_COMPILER_BIN) build/dist/kumachan
	terser -c -m -- $(K_RUNTIME_JS) > build/dist/kumachan-runtime.js

clean:
	rm build/dev/*
	rm build/dist/*
