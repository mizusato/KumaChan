package loader

import (
	"fmt"
	"os"
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/parser/ast"
	"strings"
	"path/filepath"
)


const RenamePrefix = "rename:"

func buildHierarchy (
	raw_mod  ModuleThunk,
	fs       FileSystem,
	ctx      Context,
	idx      Index,
	res      ResIndex,
) (*Module, bool, *Error) {
	/* 1. Check for validity of standalone module */
	if raw_mod.Standalone && len(ctx.BreadCrumbs) > 0 {
		return nil, false, &Error {
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
				var ver = DefaultVersion
				if manifest.Version != "" {
					ver = manifest.Version
				}
				// org.bar.foo:App::Main:dev
				return fmt.Sprintf("%s:%s::%s:%s", v, p, manifest.Name, ver)
			} else {
				// org.bar.foo::Toolkit
				return fmt.Sprintf("%s::%s", v, manifest.Name)
			}
		} else {
			return manifest.Name
		}
	})()
	var file_info = raw_mod.FileInfo
	var ast_root_node, err2 = raw_mod.Content.Load(ctx)
	if err2 != nil { return nil, false, err2 }
	ast_root_node, service_info, err3 :=
		DecorateServiceModule(ast_root_node, manifest, ctx)
	if err3 != nil { return nil, false, err3 }
	/* 3. Check the module name according to ancestor modules */
	for _, ancestor := range ctx.BreadCrumbs {
		if ancestor.ModuleName == module_name {
			/* 3.1. If there is an ancestor with the same name: */
			if os.SameFile(ancestor.FileInfo, file_info) {
				/* 3.1.1. If it corresponds to the same source file, */
				/*        throw an error of circular import. */
				return nil, false, &Error {
					Context:  ctx,
					Concrete: E_CircularImport {
						ModuleName: module_name,
					},
				}
			} else {
				/* 3.1.2. Otherwise, throw an error of module name conflict. */
				return nil, false, &Error {
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
			return existing, true, nil
		} else {
			/* 4.1.2. Otherwise, throw an error of module name conflict. */
			return nil, false, &Error {
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
		for _, cmd := range ast_root_node.Statements {
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
				if exists || local_alias == SelfModule {
					return nil, false, &Error {
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
				var im_mod, err = loadModule(im_path, fs, im_ctx, idx, res)
				if err != nil {
					// Bubble errors
					return nil, false, err
				} else {
					// Register the imported module
					var im_name = string(im_mod.Name)
					if imported_set[im_name] { return nil, false, &Error {
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
			AST:      ast_root_node,
			ImpMap:   imported_map,
			FileInfo: file_info,
			Manifest: manifest,
			ModuleServiceInfo: service_info,
		}
		idx[module_name] = mod
		return mod, false, nil
	}
}

