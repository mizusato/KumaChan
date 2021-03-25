package loader

import (
	"os"
	"time"
	"io/ioutil"
)


type FileSystem interface {
	Open(path string) (File, error)
}

type File interface {
	Close() error
	Info() (os.FileInfo, error)
	ReadDir() ([] os.FileInfo, error)
	ReadContent() ([] byte, error)
}

func ReadFile(path string, fs FileSystem) ([] byte, error) {
	var fd, err = fs.Open(path)
	if err != nil { return nil, err }
	return fd.ReadContent()
}

type RealFileSystem struct {}
type RealFile struct {
	Fd  *os.File
}
func (_ RealFileSystem) Open(path string) (File, error) {
	var fd, err = os.Open(path)
	if err != nil { return nil, err }
	return RealFile { fd }, nil
}
func (f RealFile) Close() error {
	return f.Fd.Close()
}
func (f RealFile) Info() (os.FileInfo, error) {
	return f.Fd.Stat()
}
func (f RealFile) ReadDir() ([] os.FileInfo, error) {
	return f.Fd.Readdir(0)
}
func (f RealFile) ReadContent() ([] byte, error) {
	return ioutil.ReadAll(f.Fd)
}

type craftedFileInfo struct {
	name     string
	size     int64
	mode     os.FileMode
	modTime  time.Time
	isDir    bool
	sys      interface {}
}
func (info craftedFileInfo) Name() string { return info.name }
func (info craftedFileInfo) Size() int64 { return info.size }
func (info craftedFileInfo) Mode() os.FileMode { return info.mode }
func (info craftedFileInfo) ModTime() time.Time { return info.modTime }
func (info craftedFileInfo) IsDir() bool { return info.isDir }
func (info craftedFileInfo) Sys() interface{} { return info.sys }

