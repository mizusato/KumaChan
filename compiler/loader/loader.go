package loader

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	"kumachan/rpc"
	"kumachan/lang"
	"kumachan/compiler/loader/parser/ast"
)


const SourceSuffix = ".km"
const BundleSuffix = ".kmx"
const ManifestFileName = "module.json"
const StandaloneScriptModuleName = "Main"
const ModuleKind_Service = "service"

type Module struct {
	Vendor    string
	Project   string
	Name      string
	Path      string
	AST       ast.Root
	ImpMap    map[string] *Module
	FileInfo  os.FileInfo
	Manifest  Manifest
	ModuleServiceInfo
}
type ModuleServiceInfo struct {
	IsService           bool
	ServiceIdentifier   rpc.ServiceIdentifier
	ServiceArgTypeName  string
	ServiceMethodNames  [] string
}
type Index     map[string] *Module
type ResIndex  map[string] lang.Resource

func loadModule (
	path  string,
	fs    FileSystem,
	ctx   Context,
	idx   Index,
	res   ResIndex,
) (*Module, *Error) {
	// Try to read the content of given source file/folder
	var is_project_root = (len(ctx.BreadCrumbs) == 0)
	var mod_thunk, err1 = readModulePath(path, fs, is_project_root)
	if err1 != nil { return nil, &Error {
		Context:  ctx,
		Concrete: E_ReadFileFailed {
			FilePath: path,
			Message:  err1.Error(),
		},
	} }
	var mod, loaded, err2 = buildHierarchy(mod_thunk, fs, ctx, idx, res)
	if err2 != nil { return nil, err2 }
	if !(loaded) {
		// merge resources from newly loaded modules
		for k, v := range mod_thunk.Resources {
			res[k] = v
		}
	}
	return mod, nil
}

func loadEntry(path string, fs FileSystem) (*Module, Index, ResIndex, *Error) {
	var idx = make(Index)
	for k, v := range __StdLibIndex {
		idx[k] = v
	}
	var ctx = MakeEntryContext()
	var res = make(ResIndex)
	var mod, err = loadModule(path, fs, ctx, idx, res)
	return mod, idx, res, err
}

func loadEntryThunk(t ModuleThunk, fs FileSystem) (*Module, Index, ResIndex, *Error) {
	var idx = make(Index)
	for k, v := range __StdLibIndex {
		idx[k] = v
	}
	var ctx = MakeEntryContext()
	var res = make(ResIndex)
	var mod, _, err = buildHierarchy(t, fs, ctx, idx, res)
	return mod, idx, res, err
}

func entryPathToAbsPath(path string) (string, *Error) {
	var abs_path, err = filepath.Abs(path)
	if err != nil { return "", &Error {
		Context:  MakeEntryContext(),
		Concrete: E_ReadFileFailed {
			FilePath: path,
			Message:  "cannot get absolute path of the given file",
		},
	} }
	return abs_path, nil
}

func LoadEntry(path string) (*Module, Index, ResIndex, *Error) {
	var abs_path, e = entryPathToAbsPath(path)
	if e != nil { return nil, nil, nil, e }
	path = abs_path
	if strings.HasSuffix(path, BundleSuffix) {
		return loadEntryZipFile(path)
	} else {
		return loadEntry(path, RealFileSystem {})
	}
}

func LoadEntryThunk(raw_mod ModuleThunk) (*Module, Index, ResIndex, *Error) {
	return loadEntryThunk(raw_mod, RealFileSystem {})
}

func loadEntryZipFile(path string) (*Module, Index, ResIndex, *Error) {
	var content, err = ioutil.ReadFile(path)
	if err != nil { return nil, nil, nil, &Error {
		Context:  MakeEntryContext(),
		Concrete: E_ReadFileFailed {
			FilePath: path,
			Message:  err.Error(),
		},
	} }
	return LoadEntryZipData(content, path)
}

func LoadEntryWithinFileSystem(path string, fs FileSystem) (*Module, Index, ResIndex, *Error) {
	var _, is_real_fs = fs.(RealFileSystem)
	if is_real_fs {
		var abs_path, err = entryPathToAbsPath(path)
		if err != nil { return nil, nil, nil, err }
		path = abs_path
	}
	return loadEntry(path, fs)
}

func LoadEntryZipData(data ([] byte), dummy_path string) (*Module, Index, ResIndex, *Error) {
	var fs, err = LoadZipFile(data, dummy_path)
	if err != nil { return nil, nil, nil, &Error {
		Context:  MakeEntryContext(),
		Concrete: E_ReadFileFailed {
			FilePath: dummy_path,
			Message:  err.Error(),
		},
	} }
	return loadEntry(dummy_path, fs)
}

