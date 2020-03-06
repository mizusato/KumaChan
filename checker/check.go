package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/node"
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
	LocalValues  map[string] Type
}

type Expr struct {
	Type   Type
	Info   ExprInfo
	Value  ExprVal
}
type ExprInfo struct {
	ErrorPoint  ErrorPoint
}
type ExprVal interface { ExprVal() }

type SemiExpr struct {
	Info   ExprInfo
	Value  SemiExprVal
}
type SemiExprVal interface { SemiExprVal() }
func (impl TypedExpr) SemiExprVal() {}
type TypedExpr Expr
func LiftTyped(expr Expr) SemiExpr {
	return SemiExpr {
		Info:  expr.Info,
		Value: TypedExpr(expr),
	}
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

func (ctx ExprContext) DescribeType(t Type) string {
	return DescribeType(t, ctx.GetTypeContext())
}

func (ctx ExprContext) GetModuleName() string {
	return loader.Id2String(ctx.ModuleInfo.Module.Node.Name)
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

func (ctx ExprContext) WithAddedLocalValues(added map[string]Type) (ExprContext, string) {
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
		LocalValues: merged,
	}, ""
}

func (ctx ExprContext) GetErrorPoint(node node.Node) ErrorPoint {
	return ErrorPoint {
		AST:  ctx.ModuleInfo.Module.AST,
		Node: node,
	}
}

func (ctx ExprContext) GetExprInfo(node node.Node) ExprInfo {
	return ExprInfo { ErrorPoint: ctx.GetErrorPoint(node) }
}


func Check(expr node.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	return CheckCall(DesugarExpr(expr), ctx)
}

func CheckTerm(term node.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	switch t := term.Term.(type) {
	case node.Cast:
		return CheckCast(t, ctx)
	case node.Match:
		return CheckMatch(t, ctx)
	case node.If:
		return CheckIf(t, ctx)
	case node.Block:
		return CheckBlock(t, ctx)
	case node.Tuple:
		return CheckTuple(t, ctx)
	case node.Bundle:
		return CheckBundle(t, ctx)
	case node.Get:
		return CheckGet(t, ctx)
	case node.Array:
		return CheckArray(t, ctx)
	case node.Text:
		return CheckText(t, ctx)
	case node.VariousLiteral:
		switch l := t.Literal.(type) {
		case node.IntegerLiteral:
			return CheckInteger(l, ctx)
		case node.FloatLiteral:
			return CheckFloat(l, ctx)
		case node.StringLiteral:
			return CheckString(l, ctx)
		default:
			panic("impossible branch")
		}
	case node.Ref:
		return CheckRef(t, ctx)
	default:
		panic("impossible branch")
	}
}
