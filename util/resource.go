package util

import (
	"os"
	"path/filepath"
	"io/ioutil"
)


var interpreterResourceFolderPath = (func() string {
	var exe_dir = "."
	var exe_path, err = os.Executable()
	if err == nil {
		exe_dir = filepath.Dir(exe_path)
	}
	return filepath.Join(exe_dir, "resources")
})()

func InterpreterResourceFolderPath() string {
	return interpreterResourceFolderPath
}

func ReadInterpreterResource(file string) ([] byte) {
	var full_path = filepath.Join(interpreterResourceFolderPath, file)
	content, err := ioutil.ReadFile(full_path)
	if err != nil { panic(err) }
	return content
}

