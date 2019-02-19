all:
	mkdir build 2>/dev/null || printf ''
	cat header.go object.go linearlist.go hashtable.go main.go > build/kumago.go
	go build -o build/kumago build/kumago.go 2>&1 | ./show_error.sh

