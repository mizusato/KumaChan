package loader

import (
	"os"
	"fmt"
	"time"
	"errors"
	"strings"
	"reflect"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	. "kumachan/error"
	"kumachan/stdlib"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/parser/syntax"
	"kumachan/parser/transformer"
	"kumachan/loader/common"
	"kumachan/loader/kinds"
)


const ManifestFileName = "module.json"
const StandaloneScriptModuleName = "Main"
const SourceSuffix = ".km"
const StdlibFolder = "stdlib"
const RenamePrefix = "rename:"

var __UnitFileLoaders = [] common.UnitFileLoader {
	kinds.DataLoader(),
	kinds.QtUiLoader(),
	kinds.PNG_Loader(),
}

type Module struct {
	Vendor    string
	Project   string
	Name      string
	Path      string
	Node      ast.Root
	ImpMap    map[string] *Module
	FileInfo  os.FileInfo
}
type Index  map[string] *Module

type RawModule struct {
	FilePath    string
	FileInfo    os.FileInfo
	Manifest    RawModuleManifest
	Content     RawModuleContent
	Standalone  bool
}
type RawModuleManifest struct {
	Vendor   string            `json:"vendor"`
	Project  string            `json:"project"`
	Version  string            `json:"version"`
	Name     string            `json:"name"`
	Config   RawModuleConfig   `json:"config"`
}
type RawModuleConfig struct {
	UI    kinds.QtUiConfig   `json:"ui"`
	PNG   kinds.PNG_Config   `json:"png"`
	Data  kinds.DataConfig   `json:"data"`
}
func DefaultManifest(path string) RawModuleManifest {
	var abs_path, err = filepath.Abs(path)
	if err == nil { path = abs_path }
	var dir = filepath.Dir(path)
	var base = filepath.Base(path)
	return RawModuleManifest {
		Vendor:  "",
		Project: dir,
		Version: "",
		Name:    base,
		Config:  RawModuleConfig {
			UI:   kinds.QtUiConfig { Public: true },
			PNG:  kinds.PNG_Config { Public: true },
			Data: kinds.DataConfig { Public: true },
		},
	}
}

type RawModuleContent interface {
	Load(ctx Context)  (ast.Root, *Error)
}
func (m M_PredefinedAST) Load(Context) (ast.Root, *Error) { return m.Root, nil }
type M_PredefinedAST struct {
	Root  ast.Root
}
type M_StandaloneScript struct {
	File  SourceFile
}
type M_ModuleFolder struct {
	Files  [] common.UnitFile
}
type SourceFile struct {
	Path     string
	Content  [] byte
}
func (sf SourceFile) GetAST() (ast.Root, *parser.Error) {
	var code_string = string(sf.Content)
	var code = []rune(code_string)
	var tree, err = parser.Parse(code, syntax.RootPartName, sf.Path)
	if err != nil { return ast.Root{}, err }
	return transformer.Transform(tree).(ast.Root), nil
}

func (mod M_StandaloneScript) Load(ctx Context) (ast.Root, *Error) {
	var root, err = mod.File.GetAST()
	if err != nil { return ast.Root{}, &Error {
		Context:  ctx,
		Concrete: E_ParseFailed {
			ParserError: err,
		},
	} }
	return root, nil
}
func (mod M_ModuleFolder) Load(ctx Context) (ast.Root, *Error) {
	var ast_root = common.CreateEmptyAST("(Module Folder)")
	for _, f := range mod.Files {
		var f_root, err = f.GetAST()
		if err != nil { return ast.Root{}, &Error {
			Context:  ctx,
			Concrete: E_ParseFailed {
				ParserError: err,
			},
		} }
		for _, cmd := range f_root.Statements {
			ast_root.Statements = append(ast_root.Statements, cmd)
		}
	}
	return ast_root, nil
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
		var unit_files = make([] common.UnitFile, 0)
		for _, item := range items {
			var item_name = item.Name()
			var item_path = filepath.Join(path, item_name)
			if item_name == ManifestFileName {
				var item_content, err = ioutil.ReadFile(item_path)
				if err != nil { return RawModule{}, err }
				manifest_content = item_content
				manifest_path = item_path
				has_manifest = true
			}
		}
		var manifest RawModuleManifest
		if has_manifest {
			var err = json.Unmarshal(manifest_content, &manifest)
			if err != nil {
				return RawModule{}, errors.New (
					fmt.Sprintf("error decoding module manifest %s: %s",
						manifest_path, err.Error()))
			}
		} else {
			manifest = DefaultManifest(path)
		}
		for _, item := range items {
			var item_name = item.Name()
			var item_path = filepath.Join(path, item_name)
			if strings.HasSuffix(item_name, SourceSuffix) {
				var item_content, err = ioutil.ReadFile(item_path)
				if err != nil { return RawModule{}, err }
				unit_files = append(unit_files, SourceFile {
					Path:    item_path,
					Content: item_content,
				})
			} else {
				var loader common.UnitFileLoader
				var loader_exists = false
				outer: for _, l := range __UnitFileLoaders {
					for _, ext := range l.Extensions {
						if strings.HasSuffix(item_name, ("." + ext)) {
							loader = l
							loader_exists = true
							break outer
						}
					}
				}
				if loader_exists {
					var item_config interface{} = nil
					var config_rv = reflect.ValueOf(manifest.Config)
					var config_t = config_rv.Type()
					for i := 0; i < config_t.NumField(); i += 1 {
						var field_t = config_t.Field(i)
						if field_t.Tag.Get("json") == loader.Name {
							item_config = config_rv.Field(i).Interface()
						}
					}
					f, err := loader.Load(item_path, item_config)
					if err != nil { return RawModule{}, err }
					unit_files = append(unit_files, f)
				}
			}
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
		if strings.ContainsAny(manifest.Project, ": \r\n") {
			return RawModule{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "project" has an invalid value`))
		}
		if strings.ContainsAny(manifest.Version, ": \r\n") {
			return RawModule{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "version" has an invalid value`))
		}
		if strings.ContainsAny(manifest.Vendor, ": \r\n") {
			return RawModule{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "vendor" has an invalid value`))
		}
		return RawModule {
			FilePath: path,
			FileInfo: fd_info,
			Manifest: manifest,
			Content:  M_ModuleFolder { unit_files },
		}, nil
	} else {
		var content, err = ioutil.ReadAll(fd)
		if err != nil { return RawModule{}, err }
		return RawModule {
			FilePath: path,
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
			Standalone: true,
		}, nil
	}
}

func LoadModule(path string, ctx Context, idx Index) (*Module, *Error) {
	// Try to read the content of given source file/folder
	var raw_mod, err1 = ReadModulePath(path)
	if err1 != nil { return nil, &Error {
		Context:  ctx,
		Concrete: E_ReadFileFailed {
			FilePath:  path,
			Message:   err1.Error(),
		},
	} }
	return LoadRawModule(raw_mod, ctx, idx)
}

func LoadRawModule(raw_mod RawModule, ctx Context, idx Index) (*Module, *Error) {
	/* 1. Check for validity of standalone module */
	if raw_mod.Standalone && len(ctx.BreadCrumbs) > 0 {
		return nil, &Error {
			Context:  ctx,
			Concrete: E_StandaloneImported {},
		}
	}
	var file_path = raw_mod.FilePath
	/* 2. Try to parse the content to get an AST */
	var manifest = raw_mod.Manifest
	var module_name = (func() string {
		if manifest.Vendor != "" {
			var v = manifest.Vendor
			if manifest.Project != "" {
				var p = manifest.Project
				var ver = "dev"
				if manifest.Version != "" {
					ver = manifest.Version
				}
				// org.bar.foo:App:dev::Main
				return fmt.Sprintf("%s:%s:%s::%s", v, p, ver, manifest.Name)
			} else {
				// org.bar.foo::Toolkit
				return fmt.Sprintf("%s::%s", v, manifest.Name)
			}
		} else {
			return manifest.Name
		}
	})()
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
						FilePath2:  file_path,
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
					FilePath2:  file_path,
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
			FilePath:   file_path,
		})
		for _, cmd := range module_node.Statements {
			switch c := cmd.Statement.(type) {
			case ast.Import:
				// Execute each `import` command
				var local_alias = string(c.Name.Name)
				var rel_path = string(c.Path.Value)
				var im_ctx = Context {
					ImportPoint: ErrorPointFrom(c.Node),
					LocalAlias:  local_alias,
					BreadCrumbs: current_breadcrumbs,
				}
				var _, exists = imported_map[local_alias]
				if exists {
					return nil, &Error {
						Context: im_ctx,
						Concrete: E_ConflictAlias {
							LocalAlias: local_alias,
						},
					}
				}
				if strings.HasPrefix(rel_path, RenamePrefix) {
					var target = strings.TrimPrefix(rel_path, RenamePrefix)
					var target_mod, exists = __StdLibIndex[target]
					if exists {
						if imported_map[target] != target_mod {
							panic("something went wrong")
						}
						// rename the local alias of a stdlib module
						delete(imported_map, target)
						imported_map[local_alias] = target_mod
						continue
					}
				}
				var im_path string
				if file_info.IsDir() {
					im_path = filepath.Join(file_path, rel_path)
				} else {
					im_path = filepath.Join(filepath.Dir(file_path), rel_path)
				}
				var im_mod, err = LoadModule(im_path, im_ctx, idx)
				if err != nil {
					// Bubble errors
					return nil, err
				} else {
					// Register the imported module
					var im_name = string(im_mod.Name)
					if imported_set[im_name] { return nil, &Error {
						Context: im_ctx,
						Concrete: E_DuplicateImport {
							ModuleName: im_name,
						},
					} }
					imported_set[im_name] = true
					imported_map[local_alias] = im_mod
					idx[im_name] = im_mod
				}
			default:
				// do nothing
			}
		}
		var mod = &Module {
			Vendor:   manifest.Vendor,
			Project:  manifest.Project,
			Name:     module_name,
			Path:     file_path,
			Node:     module_node,
			ImpMap:   imported_map,
			FileInfo: file_info,
		}
		idx[module_name] = mod
		return mod, nil
	}
}

func LoadEntry(path string) (*Module, Index, *Error) {
	var idx = make(map[string] *Module)
	for k, v := range __StdLibIndex {
		idx[k] = v
	}
	var ctx = MakeEntryContext()
	var mod, err = LoadModule(path, ctx, idx)
	return mod, idx, err
}

func LoadEntryRawModule(raw_mod RawModule) (*Module, Index, *Error) {
	var idx = make(map[string] *Module)
	for k, v := range __StdLibIndex {
		idx[k] = v
	}
	var ctx = MakeEntryContext()
	var mod, err = LoadRawModule(raw_mod, ctx, idx)
	return mod, idx, err
}

var __StdLibModules = stdlib.GetModuleDirectories()
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

func CraftRawModule(manifest RawModuleManifest, path string, tree ast.Root) RawModule {
	return RawModule {
		FilePath:   path,
		FileInfo:   craftModuleFileInfo(filepath.Base(path)),
		Manifest:   manifest,
		Content:    M_PredefinedAST { tree },
		Standalone: false,
	}
}

func CraftRawEmptyModule(manifest RawModuleManifest, path string) RawModule {
	return CraftRawModule(manifest, path, common.CreateEmptyAST(path))
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
func craftModuleFileInfo(name string) os.FileInfo {
	return craftedFileInfo {
		name: name,
	}
}

