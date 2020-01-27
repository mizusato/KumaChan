package loader

import (
	"io/ioutil"
	. "kumachan/error"
	"kumachan/parser"
	"kumachan/transformer"
	"kumachan/transformer/node"
	"os"
	"path/filepath"
)


type Module struct {
	Node      node.Module
	AST       *parser.Tree
	ImpMap    map[string] *Module
	FileInfo  os.FileInfo
}

type Index  map[string] *Module

func ReadFile(path string) ([]byte, os.FileInfo, error) {
	var fd, err1 = os.Open(path)
	if err1 != nil { return nil, nil, err1 }
	defer fd.Close()
	var info, err2 = fd.Stat()
	if err2 != nil { return nil, nil, err2 }
	var content, err3 = ioutil.ReadAll(fd)
	if err3 != nil { return nil, nil, err3 }
	return content, info, nil
}

func LoadModule(path string, ctx Context, idx Index) (*Module, Error) {
	/* 1. Try to read the content of given source file */
	var file_content, file_info, err1 = ReadFile(path)
	if err1 != nil { return nil, E_ReadFileFailed {
		FilePath:  path,
		Message:   err1.Error(),
		Context:   ctx,
	} }
	/* 2. Try to parse the content and generate an AST */
	var code_string = string(file_content)
	var code = []rune(code_string)
	var ast, err2 = parser.Parse(code, "module", path)
	if err2 != nil { return nil, E_ParseFailed {
		PartialAST:   ast,
		ParserError:  err2,
		Context:      ctx,
	} }
	/* 3. Transform the AST to typed structures */
	var module_node = transformer.Transform(ast)
	/* 4. Extract the module name */
	var module_name = string(module_node.Name.Name)
	/* 5. Check the module name according to ancestor modules */
	for _, ancestor := range ctx.BreadCrumbs {
		if ancestor.ModuleName == module_name {
			/* 5.1. If there is an ancestor with the same name: */
			if os.SameFile(ancestor.FileInfo, file_info) {
				/* 5.1.1. If it corresponds to the same source file, */
				/*        throw an error of circular import. */
				return nil, E_CircularImport {
					ModuleName: module_name,
					Context:    ctx,
				}
			} else {
				/* 5.1.2. Otherwise, throw an error of module name conflict. */
				return nil, E_NameConflict {
					ModuleName: module_name,
					FilePath1:  ancestor.FilePath,
					FilePath2:  path,
					Context:    ctx,
				}
			}
		}
	}
	/* 6. Check the module name according to previous sibling (sub)modules */
	var existing, exists = idx[module_name]
	if exists {
		/* 6.1. If there is a sibling (sub)module with the same name */
		if os.SameFile(existing.FileInfo, file_info) {
			/* 6.1.1. If it corresponds to the same source file, */
			/*        which indicates the module has already been loaded, */
			/*        return the loaded module. */
			return existing, nil
		} else {
			/* 6.1.2. Otherwise, throw an error of module name conflict. */
			return nil, E_NameConflict {
				ModuleName:  module_name,
				FilePath1:   existing.AST.Name,
				FilePath2:   path,
				Context:     ctx,
			}
		}
	} else {
		/* 6.2. Otherwise, load all submodules of current module */
		/*      and then return the current module. */
		var imported_map = make(map[string]*Module)
		var current_breadcrumbs = append(ctx.BreadCrumbs, Ancestor {
			ModuleName: module_name,
			FileInfo:   file_info,
			FilePath:   path,
		})
		for _, cmd := range module_node.Commands {
			switch c := cmd.Command.(type) {
			case node.Import:
				// Execute each `import` command
				var local_alias = string(c.Name.Name)
				var relpath = string(c.Path.Value)
				var imctx = Context {
					ImportPoint: ErrorPoint {
						AST:  ast,
						Node: c.Node,
					},
					LocalAlias:  local_alias,
					BreadCrumbs: current_breadcrumbs,
				}
				var _, exists = imported_map[local_alias]
				if exists {
					return nil, E_ConflictAlias {
						LocalAlias:  local_alias,
						Context:     imctx,
					}
				}
				var impath = filepath.Join(filepath.Dir(path), relpath)
				var immod, err = LoadModule(impath, imctx, idx)
				if err != nil {
					// Bubble errors
					return nil, err
				} else {
					// Register the imported module
					var immod_name = string(immod.Node.Name.Name)
					imported_map[local_alias] = immod
					idx[immod_name] = immod
				}
			default:
				// do nothing
			}
		}
		return &Module {
			Node:      module_node,
			ImpMap:    imported_map,
			AST:       ast,
			FileInfo:  file_info,
		}, nil
	}
}

func LoadEntry (path string) (*Module, Index, Error) {
	var idx = make(Index)
	var ctx = MakeEntryContext()
	var mod, err = LoadModule(path, ctx, idx)
	return mod, idx, err
}