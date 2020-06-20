package checker

import (
	"strings"
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/ast"
)


type Macro struct {
	Node    ast.Node
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

func (impl UntypedMacroInflation) SemiExprVal() {}
type UntypedMacroInflation struct {
	Macro      *Macro
	MacroName  string
	Arguments  [] ast.Expr
	Point      ErrorPoint
	Context    ExprContext
}


func CollectMacros(mod *loader.Module, store MacroStore) (MacroCollection, *MacroError) {
	var mod_name = mod.Name
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var collection = make(MacroCollection)
	for _, imported := range mod.ImpMap {
		var imp_mod_name = imported.Name
		var imp_col, err = CollectMacros(imported, store)
		if err != nil { return nil, err }
		for name, macro_ref := range imp_col {
			if !(macro_ref.IsImported) && macro_ref.Macro.Public {
				var existing, exists = collection[name]
				if exists { return nil, &MacroError {
					Point:    ErrorPointFrom(mod.Node.Node),  // TODO: more specific info
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
	for _, stmt := range mod.Node.Statements {
		switch decl := stmt.Statement.(type) {
		case ast.DeclMacro:
			var name = loader.Id2String(decl.Name)
			if name == IgnoreMark || strings.HasSuffix(name, MacroSuffix) {
				return nil, &MacroError {
					Point:    ErrorPointFrom(decl.Name.Node),
					Concrete: E_InvalidMacroName { name },
				}
			}
			var existing, exists = collection[name]
			if exists {
				var point = ErrorPointFrom(decl.Name.Node)
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
			var input = make([]string, len(decl.Input))
			for i, item := range decl.Input {
				input[i] = loader.Id2String(item)
			}
			var m = &Macro {
				Node:   decl.Node,
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


func AssignMacroInflationTo(expected Type, e UntypedMacroInflation, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var m = e.Macro
	var name = e.MacroName
	var args = e.Arguments
	var point = e.Point
	var e_ctx = e.Context
	var wrap_error = func(err *ExprError) *ExprError {
		return &ExprError {
			Point:    point,
			Concrete: E_MacroExpandingFailed {
				MacroName: name,
				Deeper:    err,
			},
		}
	}
	var m_ctx, err1 = e_ctx.WithMacroExpanded(m, name, args, point)
	if err1 != nil { return Expr{}, wrap_error(err1) }
	var semi, err2 = Check(m.Output, m_ctx)
	if err2 != nil { return Expr{}, wrap_error(err2) }
	var expr, err3 = AssignTo(expected, semi, m_ctx.WithInferringInfoFrom(ctx))
	if err3 != nil { return Expr{}, wrap_error(err3) }
	return Expr {
		Type:  expr.Type,
		Value: expr.Value,
		Info:  info,
	}, nil
}


func AdaptMacroArgs(call ast.Call) [] ast.Expr {
	var _, has_arg = call.Arg.(ast.Call)
	if has_arg {
		return [] ast.Expr { ast.WrapCallAsExpr(call) }
	} else {
		var tuple, is_tuple = call.Func.Term.(ast.Tuple)
		if is_tuple {
			return tuple.Elements
		} else {
			return [] ast.Expr{ ast.WrapCallAsExpr(call) }
		}
	}
}


func (ctx ExprContext) WithMacroExpanded (
	m      *Macro,
	name   string,
	args   [] ast.Expr,
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
	var arg_map = make(map[string] ast.Expr)
	for i, name := range m.Input {
		arg_map[name] = args[i]
	}
	new_ctx.MacroPath = append(ctx.MacroPath, MacroExpanding {
		Name:  name,
		Macro: m,
		Point: point,
		Args:  arg_map,
	})
	return new_ctx, nil
}

func (ctx ExprContext) WithMacroExpandingUnwound() ExprContext {
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	if len(ctx.MacroPath) > 0 {
		var L = len(ctx.MacroPath)
		var new_path = make([] MacroExpanding, L-1)
		copy(new_path, ctx.MacroPath[:L-1])
		new_ctx.MacroPath = new_path
	} else {
		panic("something went wrong")
	}
	return new_ctx
}

func (ctx ExprContext) FindMacroArg(name string) (ast.Expr, ExprContext, bool) {
	if L := len(ctx.MacroPath); L > 0 {
		var args = ctx.MacroPath[L-1].Args
		var arg, exists = args[name]
		if exists {
			return arg, ctx.WithMacroExpandingUnwound(), true
		}
	}
	return ast.Expr{}, ExprContext{}, false
}
