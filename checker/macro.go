package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/cst"
	"kumachan/transformer/ast"
)


type Macro struct {
	Node    ast.Node
	CST     *cst.Tree
	Public  bool
	Input   [] string
	Output  ast.Expr
}

type MacroReference struct {
	Macro       *Macro
	ModuleName  string
	IsImported  bool
}

type MacroCollection  map[string] MacroReference

type MacroStore map[string] MacroCollection


func CollectMacros(mod *loader.Module, functions FunctionStore, store MacroStore) (MacroCollection, *MacroError) {
	var mod_name = mod.Name
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var mod_functions = functions[mod_name]
	var collection = make(MacroCollection)
	for _, imported := range mod.ImpMap {
		var imp_mod_name = imported.Name
		var imp_col, err = CollectMacros(imported, functions, store)
		if err != nil { return nil, err }
		for name, macro_ref := range imp_col {
			if !(macro_ref.IsImported) && macro_ref.Macro.Public {
				var existing, exists = collection[name]
				if exists { return nil, &MacroError {
					Point:    ErrorPoint {
						CST:  mod.CST,
						Node: mod.Node.Name.Node,
					},
					Concrete: E_MacroConflictBetweenModules {
						Macro:   name,
						Module1: existing.ModuleName,
						Module2: imp_mod_name,
					},
				} }
				collection[name] = MacroReference {
					Macro:      macro_ref.Macro,
					ModuleName: imp_mod_name,
					IsImported: true,
				}
			}
		}
	}
	for _, cmd := range mod.Node.Commands {
		switch decl := cmd.Command.(type) {
		case ast.DeclMacro:
			var name = loader.Id2String(decl.Name)
			if name == IgnoreMark { return nil, &MacroError {
				Point:    ErrorPoint { CST: mod.CST, Node: decl.Name.Node },
				Concrete: E_InvalidMacroName { name },
			} }
			var existing, exists = collection[name]
			if exists {
				var point = ErrorPoint { CST: mod.CST, Node: decl.Name.Node }
				if existing.ModuleName == mod_name {
					return nil, &MacroError {
						Point:    point,
						Concrete: E_DuplicateMacroName { name },
					}
				} else {
					return nil, &MacroError {
						Point:    point,
						Concrete: E_MacroConflictWithImported {
							Macro:  name,
							Module: existing.ModuleName,
						},
					}
				}
			}
			var _, f_exists = mod_functions[name]
			if f_exists { return nil, &MacroError {
				Point:    ErrorPoint { CST: mod.CST, Node: decl.Name.Node },
				Concrete: E_MacroConflictWithFunction { name },
			} }
			var input = make([]string, len(decl.Input))
			for i, item := range decl.Input {
				input[i] = loader.Id2String(item)
			}
			var m = &Macro {
				Node:   decl.Node,
				CST:    mod.CST,
				Public: decl.Public,
				Input:  input,
				Output: decl.Output,
			}
			collection[name] = MacroReference {
				Macro:      m,
				ModuleName: mod_name,
				IsImported: false,
			}
		}
	}
	store[mod_name] = collection
	return collection, nil
}


func (ctx ExprContext) WithMacroExpanded (
	m      *Macro,
	name   string,
	args   [] SemiExpr,
	point  ErrorPoint,
) (ExprContext, *ExprError) {
	var given = uint(len(args))
	var required = uint(len(m.Input))
	if given != required {
		return ExprContext{}, &ExprError {
			Point:    point,
			Concrete: E_MacroWrongArgsQuantity {
				MacroName: name,
				Given:     given,
				Required:  required,
			},
		}
	}
	for _, expanded := range ctx.MacroPath {
		if expanded.Macro == m {
			return ExprContext{}, &ExprError {
				Point:    point,
				Concrete: E_MacroCircularExpanding {
					MacroName: name,
				},
			}
		}
	}
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	var arg_map = make(map[string] SemiExpr)
	for i, name := range m.Input {
		arg_map[name] = args[i]
	}
	new_ctx.MacroArgs = arg_map
	new_ctx.MacroPath = append(ctx.MacroPath, MacroExpanding {
		Name:  name,
		Macro: m,
		Point: point,
	})
	return new_ctx, nil
}
