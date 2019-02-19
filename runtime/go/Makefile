all:
	mkdir build 2>/dev/null || printf ''
	cat header.go object.go singleton.go linearlist.go hashtable.go container.go scope.go main.go > build/kumago.go
	go build -o build/kumago build/kumago.go 2>&1 | ./show_error.sh

