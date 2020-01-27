package loader

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	"kumachan/parser"
	"kumachan/transformer"
	"kumachan/transformer/node"
	."kumachan/error"
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
	var file_content, file_info, err1 = ReadFile(path)
	if err1 != nil { return nil, E_ReadFileFailed {
		FilePath:  path,
		Message:   err1.Error(),
		Context:   ctx,
	} }
	var code_string = string(file_content)
	var code = []rune(code_string)
	var ast, err2 = parser.Parse(code, "module", path)
	if err2 != nil { return nil, E_ParseFailed {
		PartialAST:   ast,
		ParserError:  err2,
		Context:      ctx,
	} }
	var module_node = transformer.Transform(ast)
	var module_name = string(module_node.Name.Name)
	for _, ancestor := range ctx.BreadCrumbs {
		if ancestor.ModuleName == module_name {
			if os.SameFile(ancestor.FileInfo, file_info) {
				return nil, E_CircularImport {
					ModuleName: module_name,
					Context:    ctx,
				}
			} else {
				return nil, E_NameConflict {
					ModuleName: module_name,
					FilePath1:  ancestor.FilePath,
					FilePath2:  path,
					Context:    ctx,
				}
			}
		}
	}
	var existing, exists = idx[module_name]
	if exists {
		if os.SameFile(existing.FileInfo, file_info) {
			return existing, nil
		} else {
			return nil, E_NameConflict {
				ModuleName:  module_name,
				FilePath1:   existing.AST.Name,
				FilePath2:   path,
				Context:     ctx,
			}
		}
	} else {
		var imported_map = make(map[string]*Module)
		for _, cmd := range module_node.Commands {
			switch c := cmd.Command.(type) {
			case node.Import:
				var local_alias = string(c.Name.Name)
				var relpath = strings.Trim(string(c.Path.Value), "'") // fixme: ugly, should be handled before loader phase
				var impath = filepath.Join(filepath.Dir(path), relpath)
				var immod, err = LoadModule (
					impath,
					Context{
						ImportPoint: ErrorPoint {
							AST:  ast,
							Node: c.Node,
						},
						LocalAlias:  local_alias,
						BreadCrumbs: append(ctx.BreadCrumbs, Ancestor {
							ModuleName: module_name,
							FileInfo:   file_info,
							FilePath:   path,
						}),
					},
					idx,
				)
				if err != nil {
					return nil, err
				} else {
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