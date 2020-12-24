package loader

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"
	"io/ioutil"
	"archive/zip"
	"path/filepath"
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

func LoadZipFile(content [] byte, base_path string) (ZipFileSystem, error) {
	if base_path == "" { base_path = "." }
	var content_reader = bytes.NewReader(content)
	var zip_reader, err = zip.NewReader(content_reader, int64(len(content)))
	if err != nil { return ZipFileSystem {}, err }
	var index = make(map[string] *ZipItemFile)
	var dir_files = make(map[string] ([] *ZipItemFile))
	index[base_path] = &ZipItemFile {
		isRootFile:   true,
		rootFileInfo: craftedFileInfo {
			name:    filepath.Base(base_path),
			isDir:   true,
		},
	}
	for _, f := range zip_reader.File {
		var rel_path, err = validateZipItemFileName(f.Name)
		if err != nil { return ZipFileSystem {}, err }
		var abs_path = filepath.Join(base_path, rel_path)
		var _, exists = index[abs_path]
		if exists { return ZipFileSystem {}, errors.New(fmt.Sprintf(
			"duplicate file: %s", abs_path))}
		var item = &ZipItemFile {
			underlyingFile:  f,
		}
		index[abs_path] = item
		var dir_path = filepath.Dir(abs_path)
		_, exists = dir_files[dir_path]
		if !(exists) { dir_files[dir_path] = make([] *ZipItemFile, 0) }
		dir_files[dir_path] = append(dir_files[dir_path], item)
	}
	for abs_path, files := range dir_files {
		var item, exists = index[abs_path]
		if !(exists) {
			return ZipFileSystem {}, errors.New(fmt.Sprintf(
				"missing directory item: %s", abs_path))
		}
		if !(item.info().IsDir()) {
			return ZipFileSystem {}, errors.New(fmt.Sprintf(
				"invalid directory item: %s", abs_path))
		}
		item.containingFiles = files
	}
	return ZipFileSystem { files: index }, nil
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


type ZipFileSystem struct {
	files  map[string] *ZipItemFile
}
type ZipItemFile struct {
	isRootFile       bool
	rootFileInfo     craftedFileInfo
	underlyingFile   *zip.File
	containingFiles  [] *ZipItemFile
}
func (fs ZipFileSystem) AllFilesPathList() ([] string) {
	var path_list = make([] string, 0, len(fs.files))
	for path, _ := range fs.files {
		path_list = append(path_list, path)
	}
	return path_list
}
func (fs ZipFileSystem) Open(path string) (File, error) {
	var f, exists = fs.files[path]
	if !(exists) {
		var path_sep = string([] rune { os.PathSeparator })
		f, exists = fs.files[strings.TrimRight(path, path_sep)]
	}
	if !(exists) {
		return nil, errors.New(fmt.Sprintf("file does not exist: %s", path))
	}
	return f, nil
}
func (_ ZipItemFile) Close() error {
	return nil
}
func (f ZipItemFile) info() os.FileInfo {
	if f.isRootFile {
		return f.rootFileInfo
	} else {
		return f.underlyingFile.FileInfo()
	}
}
func (f ZipItemFile) Info() (os.FileInfo, error) {
	return f.info(), nil
}
func (f ZipItemFile) ReadDir() ([] os.FileInfo, error) {
	var info_list = make([] os.FileInfo, len(f.containingFiles))
	for i, f := range f.containingFiles {
		info_list[i] = f.info()
	}
	return info_list, nil
}
func (f ZipItemFile) ReadContent() ([] byte, error) {
	if f.isRootFile {
		return nil, errors.New(fmt.Sprint(
			"cannot read binary content of zip root folder"))
	}
	var read_closer, err = f.underlyingFile.Open()
	if err != nil { return nil, err }
	defer func() {
		_ = read_closer.Close()
	}()
	return ioutil.ReadAll(read_closer)
}


func validateZipItemFileName(name string) (string, error) {
	if strings.HasPrefix(name, "/") {
		return "", errors.New(fmt.Sprintf(
			"the name of this zip item file is not a relative path: %s", name))
	}
	var segments = strings.Split(name, "/")
	for _, s := range segments {
		if s == ".." {
			return "", errors.New(fmt.Sprintf(
				"the name of this zip item file is not valid: %s", name) +
				" (using '..' could lead to path traversal)")
		} else if s == "." {
			return "", errors.New(fmt.Sprintf(
				"the name of this zip item file is not valid: %s", name) +
				" (using '.' is suspicious)")
		}
	}
	return filepath.Clean(name), nil
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

