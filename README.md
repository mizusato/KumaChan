# 熊語言・KumaChan Programming Language

KumaChan (ja. くまちゃん) is an experimental programming language, and is still under development.

## Yet Another Programming Language?

I have been using JavaScript for several years and can't tolerate the BAD DESIGN of it anymore, so I created a new language.

However, the new language does NOT intend to be a substitution of JavaScript. It's designed for GUI (aka. client-side) programming only, and does NOT intend to be used on server-side. Its design focuses more on abstraction ability rather than performance.

For more details, see [documentation](https://mizusato.gitbook.io/kumachan-documentation/).

## Development Progress

This repository contains a `parser` which defines the syntax of KumaChan, and a `transpiler` which is capable to transpile KumaChan code into JavaScript, along with a necessary `runtime` library.

The current implementation of `runtime` (based on JavaScript) is created for **DEMO Purpose** and is extraordinarily naive. **I don't want to maintain it, and is going to reimplement the whole runtime using Golang.** Finally, all JavaScript code in this repository will be wiped. I want the Golang implementation to be the official implementation. 

Rewriting the `runtime` using Golang may cause difficulty when developing Web applications, because WebAssembly is not widely supported in browsers, meanwhile the binary file of Golang program is very large. (although it does not matter when developing off-line apps using Electron.)

But using JavaScript implementation of the `runtime` in production also requires a rewrite of the current code, because the current implementation is too naive to be used in production, and is not integrated with Webpack ecosystem.

Currently I am working on the rewriting of `runtime`. After I finished the new `runtime`, I am going to integrate GTK2 with it using [go-gtk](https://github.com/mattn/go-gtk) to enable desktop GUI programming for this language. I hope it can start the age of declarative programming for native desktop GUI.

## Usage

### Build

Make sure you have `make`, `bash`, `golang`, `python3` and `nodejs` installed on your system before building.

```console
$ git clone https://notabug.org/mizusato/KumaChan
$ cd KumaChan
$ make
```

### Run the REPL

```console
$ ./repl.js
```

### Run a script file

Note: shebang is NOT supported currently.

```console
$ cat << EOF > /tmp/test.k
module main
print('Change The World')
EOF
$ ./run.js /tmp/test.k
```

## Syntax Highlighting

There is a plugin called `language-kumachan` for the Atom Editor. It is still naive, therefore incorrect hightlight could happen.

## LICENSE

LGPL-3.0
