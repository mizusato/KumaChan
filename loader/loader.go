package loader

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	. "kumachan/error"
	"kumachan/parser"
	"kumachan/transformer"
	"kumachan/transformer/ast"
	"kumachan/parser/syntax"
)


const ManifestFileName = "module.json"
const StandaloneScriptModuleName = "Main"
const SourceSuffix = ".km"
const StdlibFolder = "stdlib"

type Module struct {
	Name      string
	Path      string
	Node      ast.Root
	ImpMap    map[string] *Module
	FileInfo  os.FileInfo
}
type Index  map[string] *Module

type RawModule struct {
	FileInfo  os.FileInfo
	Manifest  RawModuleManifest
	Content   RawModuleContent
}
type RawModuleManifest struct {
	Name  string   `json:"name"`
}
type RawModuleContent interface {
	Load(ctx Context)  (ast.Root, *Error)
}
type M_StandaloneScript struct {
	File  SourceFile
}
type M_ModuleFolder struct {
	Files  [] SourceFile
}
type SourceFile struct {
	Path     string
	Content  [] byte
}

func (mod M_StandaloneScript) Load(ctx Context) (ast.Root, *Error) {
	const syntax_root = syntax.RootPartName
	var code_string = string(mod.File.Content)
	var code = []rune(code_string)
	var tree, err = parser.Parse(code, syntax_root, mod.File.Path)
	if err != nil { return ast.Root{}, &Error {
		Context:  ctx,
		Concrete: E_ParseFailed {
			ParserError: err,
		},
	} }
	var root = transformer.Transform(tree)
	return root, nil
}
func (mod M_ModuleFolder) Load(ctx Context) (ast.Root, *Error) {
	const syntax_root = syntax.RootPartName
	var empty_tree, err = parser.Parse (
		([] rune)("#"), syntax_root, "(Module Folder)",
	)
	if err != nil { panic("something went wrong") }
	var mod_root = transformer.Transform(empty_tree)
	for _, f := range mod.Files {
		var code_string = string(f.Content)
		var code = []rune(code_string)
		var tree, err = parser.Parse(code, syntax_root, f.Path)
		if err != nil { return ast.Root{}, &Error {
			Context:  ctx,
			Concrete: E_ParseFailed {
				ParserError: err,
			},
		} }
		var f_root = transformer.Transform(tree)
		for _, cmd := range f_root.Commands {
			mod_root.Commands = append(mod_root.Commands, cmd)
		}
	}
	return mod_root, nil
}

func ReadModulePath(path string) (RawModule, error) {
	fd, err := os.Open(path)
	if err != nil { return RawModule{}, err }
	defer func() {
		_ = fd.Close()
	} ()
	fd_info, err := fd.Stat()
	if err != nil { return RawModule{}, err }
	if fd_info.IsDir() {
		items, err := fd.Readdir(0)
		if err != nil { return RawModule{}, err }
		var has_manifest = false
		var manifest_content ([] byte)
		var manifest_path string
		var source_files = make([] SourceFile, 0)
		for _, item := range items {
			var item_name = item.Name()
			var item_path = filepath.Join(path, item_name)
			if item_name == ManifestFileName {
				var item_content, err = ioutil.ReadFile(item_path)
				if err != nil { return RawModule{}, err }
				manifest_content = item_content
				manifest_path = item_path
				has_manifest = true
			} else if strings.HasSuffix(item_name, SourceSuffix) {
				var item_content, err = ioutil.ReadFile(item_path)
				if err != nil { return RawModule{}, err }
				source_files = append(source_files, SourceFile {
					Path:    item_path,
					Content: item_content,
				})
			}
		}
		if !(has_manifest) {
			return RawModule{}, errors.New (
				fmt.Sprintf("missing manifest(%s) in module folder %s",
					ManifestFileName, path))
		}
		var manifest RawModuleManifest
		err = json.Unmarshal(manifest_content, &manifest)
		if err != nil {
			return RawModule{}, errors.New (
				fmt.Sprintf("error decoding module manifest %s: %s",
					manifest_path, err.Error()))
		}
		var mod_name = manifest.Name
		if mod_name == "" {
			return RawModule{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "name" cannot be empty or omitted`))
		}
		if !(syntax.GetIdentifierFullRegexp().MatchString(mod_name)) {
			return RawModule{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "name" has an invalid value`))
		}
		return RawModule {
			FileInfo: fd_info,
			Manifest: manifest,
			Content:  M_ModuleFolder { source_files },
		}, nil
	} else {
		var content, err = ioutil.ReadAll(fd)
		if err != nil { return RawModule{}, err }
		return RawModule {
			FileInfo: fd_info,
			Manifest: RawModuleManifest {
				Name: StandaloneScriptModuleName,
			},
			Content:  M_StandaloneScript {
				File: SourceFile {
					Path:    path,
					Content: content,
				},
			},
		}, nil
	}
}

func LoadModule(path string, ctx Context, idx Index) (*Module, *Error) {
	/* 1. Try to read the content of given source file/folder */
	var raw_mod, err1 = ReadModulePath(path)
	if err1 != nil { return nil, &Error {
		Context:  ctx,
		Concrete: E_ReadFileFailed {
			FilePath:  path,
			Message:   err1.Error(),
		},
	} }
	/* 2. Try to parse the content to get an AST */
	var module_name = raw_mod.Manifest.Name
	var file_info = raw_mod.FileInfo
	var module_node, err2 = raw_mod.Content.Load(ctx)
	if err2 != nil { return nil, err2 }
	/* 3. Check the module name according to ancestor modules */
	for _, ancestor := range ctx.BreadCrumbs {
		if ancestor.ModuleName == module_name {
			/* 3.1. If there is an ancestor with the same name: */
			if os.SameFile(ancestor.FileInfo, file_info) {
				/* 3.1.1. If it corresponds to the same source file, */
				/*        throw an error of circular import. */
				return nil, &Error {
					Context:  ctx,
					Concrete: E_CircularImport {
						ModuleName: module_name,
					},
				}
			} else {
				/* 3.1.2. Otherwise, throw an error of module name conflict. */
				return nil, &Error {
					Context:  ctx,
					Concrete: E_NameConflict {
						ModuleName: module_name,
						FilePath1:  ancestor.FilePath,
						FilePath2:  path,
					},
				}
			}
		}
	}
	/* 4. Check the module name according to previous sibling (sub)modules */
	var existing, exists = idx[module_name]
	if exists {
		/* 4.1. If there is a sibling (sub)module with the same name */
		if os.SameFile(existing.FileInfo, file_info) {
			/* 4.1.1. If it corresponds to the same source file, */
			/*        which indicates the module has already been loaded, */
			/*        return the loaded module. */
			return existing, nil
		} else {
			/* 4.1.2. Otherwise, throw an error of module name conflict. */
			return nil, &Error {
				Context: ctx,
				Concrete: E_NameConflict {
					ModuleName: module_name,
					FilePath1:  existing.Path,
					FilePath2:  path,
				},
			}
		}
	} else {
		/* 4.2. Otherwise, load all submodules of current module */
		/*      and then return the current module. */
		var imported_map = make(map[string] *Module)
		var imported_set = make(map[string] bool)
		ImportStdLib(imported_map, imported_set)
		var current_breadcrumbs = append(ctx.BreadCrumbs, Ancestor {
			ModuleName: module_name,
			FileInfo:   file_info,
			FilePath:   path,
		})
		for _, cmd := range module_node.Commands {
			switch c := cmd.Command.(type) {
			case ast.Import:
				// Execute each `import` command
				var local_alias = string(c.Name.Name)
				var relpath = string(c.Path.Value)
				var imctx = Context {
					ImportPoint: ErrorPointFrom(c.Node),
					LocalAlias:  local_alias,
					BreadCrumbs: current_breadcrumbs,
				}
				var _, exists = imported_map[local_alias]
				if exists {
					return nil, &Error {
						Context: imctx,
						Concrete: E_ConflictAlias {
							LocalAlias: local_alias,
						},
					}
				}
				var impath string
				if file_info.IsDir() {
					impath = filepath.Join(path, relpath)
				} else {
					impath = filepath.Join(filepath.Dir(path), relpath)
				}
				var immod, err = LoadModule(impath, imctx, idx)
				if err != nil {
					// Bubble errors
					return nil, err
				} else {
					// Register the imported module
					var immod_name = string(immod.Name)
					if imported_set[immod_name] { return nil, &Error {
						Context: imctx,
						Concrete: E_DuplicateImport {
							ModuleName: immod_name,
						},
					} }
					imported_set[immod_name] = true
					imported_map[local_alias] = immod
					idx[immod_name] = immod
				}
			default:
				// do nothing
			}
		}
		var mod = &Module {
			Name:     module_name,
			Path:     path,
			Node:     module_node,
			ImpMap:   imported_map,
			FileInfo: file_info,
		}
		idx[module_name] = mod
		return mod, nil
	}
}

func LoadEntry (path string) (*Module, Index, *Error) {
	var idx = __StdLibIndex
	var ctx = MakeEntryContext()
	var mod, err = LoadModule(path, ctx, idx)
	return mod, idx, err
}

var __StdLibModules = []string { "core", "io", "os" }
var __StdLibIndex = make(map[string] *Module)
var _ = __Init()

func __Init() interface{} {
	LoadStdLib()
	return nil
}

func LoadStdLib() {
	var exe_path, err = os.Executable()
	if err != nil { panic(err) }
	var ctx = MakeEntryContext()
	for _, name := range __StdLibModules {
		var file = filepath.Join (
			filepath.Dir(exe_path), StdlibFolder, name,
		)
		var _, err = LoadModule(file, ctx, __StdLibIndex)
		if err != nil {
			fmt.Fprintf (
				os.Stderr,
				"%v*** Failed to Load Standard Library%v\n*\n%s\n",
				"\033[1m", "\033[0m", err.Error(),
			)
			os.Exit(3)
		}
	}
}

func ImportStdLib (imp_map map[string]*Module, imp_set map[string]bool) {
	for name, mod := range __StdLibIndex {
		imp_map[name] = mod
		imp_set[name] = true
	}
}
