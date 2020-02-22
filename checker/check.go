package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
	"math/big"
)


type ModuleInfo struct {
	Module     *loader.Module
	Types      TypeRegistry
	Constants  ConstantCollection   // TODO: check naming conflict between
	Functions  FunctionCollection   //       constants and functions
}

type ExprContext struct {
	ModuleInfo   ModuleInfo
	TypeParams   [] string
	TypeArgs     map[uint] Type
	LocalValues  map[string] Type
}

type Sym interface { Sym() }
func (impl SymLocalValue) Sym() {}
type SymLocalValue struct { ValueType Type }
func (impl SymConst) Sym() {}
type SymConst struct { Const *Constant }
func (impl SymTypeParam) Sym() {}
type SymTypeParam struct { Index uint }
func (impl SymType) Sym() {}
type SymType struct { Type *GenericType }
func (impl SymFunctions) Sym() {}
type SymFunctions struct { Functions []*GenericFunction }


func (ctx ExprContext) GetTypeContext() TypeContext {
	return TypeContext {
		Module: ctx.ModuleInfo.Module,
		Params: ctx.TypeParams,
		Ireg:   ctx.ModuleInfo.Types,
	}
}

func (ctx ExprContext) LookupSymbol(raw loader.Symbol) (Sym, bool) {
	var mod_name = raw.ModuleName
	var sym_name = raw.SymbolName
	if mod_name == "" {
		var t, exists = ctx.LocalValues[sym_name]
		if exists {
			return SymLocalValue { ValueType: t }, true
		}
		for index, param_name := range ctx.TypeParams {
			if param_name == sym_name {
				return SymTypeParam { Index: uint(index) }, true
			}
		}
		f_refs, exists := ctx.ModuleInfo.Functions[sym_name]
		if exists {
			var functions = make([]*GenericFunction, len(f_refs))
			for i, ref := range f_refs {
				functions[i] = ref.Function
			}
			return SymFunctions { Functions: functions }, true
		}
		return nil, false
	} else {
		var g, exists = ctx.ModuleInfo.Types[raw]
		if exists {
			return SymType { Type: g }, true
		}
		constant, exists := ctx.ModuleInfo.Constants[raw]
		if exists {
			return SymConst { Const: constant }, true
		}
		return nil, false
	}
}

func (ctx ExprContext) WithLocalValues(added map[string]Type) (ExprContext, string) {
	var merged = make(map[string]Type)
	for name, t := range ctx.LocalValues {
		var _, exists = added[name]
		if exists {
			return ExprContext{}, name
		}
		merged[name] = t
	}
	for name, t := range added {
		merged[name] = t
	}
	return ExprContext {
		ModuleInfo:  ctx.ModuleInfo,
		TypeParams:  ctx.TypeParams,
		TypeArgs:    ctx.TypeArgs,
		LocalValues: merged,
	}, ""
}


type SemiExpr interface { SemiExpr() }

func (impl TypedExpr) SemiExpr() {}
type TypedExpr Expr

func (impl UntypedLambda) SemiExpr() {}
type UntypedLambda struct {
	Input   Pattern
	Output  node.Expr
}

func (impl UntypedInteger) SemiExpr() {}
type UntypedInteger big.Int

func (impl SemiTypedTuple) SemiExpr() {}
type SemiTypedTuple struct {
	Values  [] SemiExpr
}

func (impl SemiTypedBundle) SemiExpr() {}
type SemiTypedBundle struct {
	Index   map[string] uint
	Values  [] SemiExpr
}


func SemiExprFromTerm(t node.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	var get_err_point = func(node node.Node) ErrorPoint {
		return ErrorPoint{
			AST:  ctx.ModuleInfo.Module.AST,
			Node: node,
		}
	}
	var info = ExprInfo { ErrorPoint: get_err_point(t.Node) }
	switch term := t.Term.(type) {
	case node.Tuple:
		var L = len(term.Elements)
		if L == 0 {
			return TypedExpr(Expr {
				Type:  AnonymousType { Unit {} },
				Value: UnitValue {},
				Info:  info,
			}), nil
		} else if L == 1 {
			var expr, err = SemiExprFrom(term.Elements[0], ctx)
			if err != nil { return nil, err }
			return expr, nil
		} else {
			var el_exprs = make([]SemiExpr, L)
			var el_typed_exprs = make([]Expr, L)
			var el_types = make([]Type, L)
			var typed_count = 0
			for i, el := range term.Elements {
				var expr, err = SemiExprFrom(el, ctx)
				if err != nil { return nil, err }
				el_exprs[i] = expr
				switch typed := expr.(type) {
				case TypedExpr:
					el_typed_exprs[i] = Expr(typed)
					el_types[i] = typed.Type
					typed_count += 1
				}
			}
			if typed_count == L {
				return TypedExpr(Expr {
					Type:  AnonymousType { Tuple { el_types } },
					Value: Product { el_typed_exprs },
					Info:  info,
				}), nil
			} else {
				return SemiTypedTuple { el_exprs }, nil
			}
		}
	case node.Bundle:
		var L = len(term.Values)
		if L == 0 {
			return TypedExpr(Expr {
				Type:  AnonymousType { Unit {} },
				Value: UnitValue {},
				Info:  info,
			}), nil
		} else {
			var f_exprs = make([]SemiExpr, L)
			var f_index_map = make(map[string]uint, L)
			var f_type_map = make(map[string]Type, L)
			var f_typed_exprs = make([]Expr, L)
			var typed_count = 0
			for i, field := range term.Values {
				var name = loader.Id2String(field.Key)
				var _, exists = f_index_map[name]
				if exists { return nil, &ExprError {
					Point:    get_err_point(field.Key.Node),
					Concrete: E_DuplicateFieldValue { name },
				} }
				var value = DesugarOmittedFieldValue(field)
				var expr, err = SemiExprFrom(value, ctx)
				if err != nil { return nil, err }
				f_exprs[i] = expr
				f_index_map[name] = uint(i)
				switch typed := expr.(type) {
				case TypedExpr:
					f_type_map[name] = typed.Type
					f_typed_exprs[i] = Expr(typed)
					typed_count += 1
				}
			}
			if typed_count == L {
				return TypedExpr(Expr {
					Type:  AnonymousType { Bundle {
						Fields: f_type_map,
						Index:  f_index_map,
					} },
					Value: Product {
						Values: f_typed_exprs,
					},
					Info:  info,
				}), nil
			} else {
				return SemiTypedBundle {
					Index:  f_index_map,
					Values: f_exprs,
				}, nil
			}
		}
		// TODO: other kinds of terms
	default:
		panic("impossible branch")
	}
}

func SemiExprFrom(e node.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	// TODO
	return nil, nil
}

func DesugarOmittedFieldValue(field node.FieldValue) node.Expr {
	switch val_expr := field.Value.(type) {
	case node.Expr:
		return val_expr
	default:
		return node.Expr {
			Node:  field.Node,
			Pipes: []node.Pipe {{
				Node:  field.Node,
				Terms: []node.VariousTerm {{
					Node: field.Node,
					Term: node.Ref {
						Node:     field.Node,
						Module:   node.Identifier{
							Node: field.Node,
							Name: []rune(""),
						},
						Specific: false,
						Id:       field.Key,
						TypeArgs: make([]node.VariousType, 0),
					},
				}},
			}},
		}
	}
}

/*

func ExprFromPipe(p node.Pipe, ctx ExprContext, input Type) (Expr, *ExprError) {
	// TODO
	// if input == nil { ...
	return Expr{}, nil
}

*/
