package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/node"
	"math/big"
	"strconv"
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

func (ctx ExprContext) GetErrorPoint(node node.Node) ErrorPoint {
	return ErrorPoint {
		AST:  ctx.ModuleInfo.Module.AST,
		Node: node,
	}
}


type SemiExpr interface { SemiExpr() }

func (impl TypedExpr) SemiExpr() {}
type TypedExpr Expr

func (impl UntypedLambda) SemiExpr() {}
type UntypedLambda struct {
	Node    node.Node
	Input   Pattern
	Output  node.Expr
}

func (impl UntypedInteger) SemiExpr() {}
type UntypedInteger struct {
	Node    node.Node
	Value  *big.Int
}

func (impl SemiTypedTuple) SemiExpr() {}
type SemiTypedTuple struct {
	Node    node.Node
	Values  [] SemiExpr
}

func (impl SemiTypedBundle) SemiExpr() {}
type SemiTypedBundle struct {
	Node    node.Node
	Index   map[string] uint
	Values  [] SemiExpr
}

func (impl SemiTypedArray) SemiExpr() {}
type SemiTypedArray struct {
	Node   node.Node
	Items  [] SemiExpr
}

func (impl SemiSet) SemiExpr() {}
type SemiSet struct {
	Base    Expr
	Bundle  Bundle
	Ops     [] SemiSetOp
}
type SemiSetOp struct {
	Node   node.Node
	Index  uint
	Value  SemiExpr
}

func SemiExprFromIntLiteral(i node.IntegerLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var chars = i.Value
	var abs_chars []rune
	if chars[0] == '-' {
		abs_chars = chars[1:]
	} else {
		abs_chars = chars
	}
	var has_base_prefix = false
	if len(abs_chars) >= 2 {
		var c1 = abs_chars[0]
		var c2 = abs_chars[1]
		if c1 == '0' && (c2 == 'x' || c2 == 'o' || c2 == 'b' || c2 == 'X' || c2 == 'O' || c2 == 'B') {
			has_base_prefix = true
		}
	}
	var str = string(chars)
	var value *big.Int
	var ok bool
	if has_base_prefix {
		value, ok = big.NewInt(0).SetString(str, 0)
	} else {
		// avoid "0" prefix to be recognized as octal with base 0
		value, ok = big.NewInt(0).SetString(str, 10)
	}
	if ok {
		return UntypedInteger {
			Node:  i.Node,
			Value: value,
		}, nil
	} else {
		return nil, &ExprError {
			Point:    ctx.GetErrorPoint(i.Node),
			Concrete: E_InvalidInteger { str },
		}
	}
}

func ExprFromFloatLiteral(f node.FloatLiteral, ctx ExprContext) (Expr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(f.Node) }
	var value, err = strconv.ParseFloat(string(f.Value), 64)
	if err != nil { panic("invalid float literal got from parser") }
	return Expr {
		Type:  NamedType {
			Name: __Float,
			Args: make([]Type, 0),
		},
		Value: FloatLiteral { value },
		Info:  info,
	}, nil
}

func ExprFromStringLiteral(s node.StringLiteral, ctx ExprContext) (Expr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(s.Node) }
	return Expr{
		Type:  NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { s.Value },
	}, nil
}

func SemiExprFromTuple(tuple node.Tuple, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(tuple.Node) }
	var L = len(tuple.Elements)
	if L == 0 {
		return TypedExpr(Expr {
			Type:  AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		}), nil
	} else if L == 1 {
		var expr, err = SemiExprFrom(tuple.Elements[0], ctx)
		if err != nil { return nil, err }
		return expr, nil
	} else {
		var el_exprs = make([]SemiExpr, L)
		var el_typed_exprs = make([]Expr, L)
		var el_types = make([]Type, L)
		var typed_count = 0
		for i, el := range tuple.Elements {
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
			return SemiTypedTuple {
				Node:   tuple.Node,
				Values: el_exprs,
			}, nil
		}
	}
}

func SemiExprFromBundle(bundle node.Bundle, ctx ExprContext) (SemiExpr, *ExprError) {
	switch update := bundle.Update.(type) {
	case node.Update:
		var base_semi, err = SemiExprFrom(update.Base, ctx)
		if err != nil { return nil, err }
		switch b := base_semi.(type) {
		case TypedExpr:
			if IsBundleLiteral(Expr(b)) { return nil, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			} }
			var L = len(bundle.Values)
			if !(L >= 1) { panic("something went wrong") }
			var base = Expr(b)
			switch target := GetBundleRepr(base.Type, ctx).(type) {
			case Bundle:
				var occurred_names = make(map[string] bool)
				var ops = make([]SemiSetOp, L)
				for i, field := range bundle.Values {
					var name = loader.Id2String(field.Key)
					var index, exists = target.Index[name]
					if !exists {
						return nil, &ExprError {
							Point: ctx.GetErrorPoint(field.Key.Node),
							Concrete: E_FieldDoesNotExist {
								Field:  name,
								Bundle: ctx.DescribeType(AnonymousType{target}),
							},
						}
					}
					var _, duplicate = occurred_names[name]
					if duplicate {
						return nil, &ExprError {
							Point:    ctx.GetErrorPoint(field.Key.Node),
							Concrete: E_ExprDuplicateField { name },
						}
					}
					occurred_names[name] = true
					var value_node = DesugarOmittedFieldValue(field)
					var value, err = SemiExprFrom(value_node, ctx)
					if err != nil { return nil, err }
					ops[i] = SemiSetOp {
						Node:  value_node.Node,
						Index: index,
						Value: value,
					}
				}
				return SemiSet {
					Base:   base,
					Bundle: target,
					Ops:    ops,
				}, nil
			case BR_BundleButOpaque:
				return nil, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToOpaqueBundle {},
				}
			case BR_NonBundle:
				return nil, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToNonBundle {},
				}
			default:
				panic("impossible branch")
			}
		case SemiTypedBundle:
			return nil, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			}
		default:
			return nil, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToNonBundle {},
			}
		}
	default:
		var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(bundle.Node) }
		var L = len(bundle.Values)
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
			for i, field := range bundle.Values {
				var name = loader.Id2String(field.Key)
				var _, exists = f_index_map[name]
				if exists { return nil, &ExprError {
					Point:    ctx.GetErrorPoint(field.Key.Node),
					Concrete: E_ExprDuplicateField { name },
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
					Node:   bundle.Node,
					Index:  f_index_map,
					Values: f_exprs,
				}, nil
			}
		}
	}
}

func ExprFromGet(get node.Get, ctx ExprContext) (Expr, *ExprError) {
	var base_semi, err = SemiExprFrom(get.Base, ctx)
	if err != nil { return Expr{}, err }
	switch b := base_semi.(type) {
	case TypedExpr:
		if IsBundleLiteral(Expr(b)) { return Expr{}, &ExprError {
			Point:    ctx.GetErrorPoint(get.Base.Node),
			Concrete: E_GetFromLiteralBundle {},
		} }
		var L = len(get.Path)
		if !(L >= 1) { panic("something went wrong") }
		var base = Expr(b)
		for _, member := range get.Path {
			switch bundle := GetBundleRepr(base.Type, ctx).(type) {
			case Bundle:
				var key = loader.Id2String(member.Name)
				var index, exists = bundle.Index[key]
				if !exists { return Expr{}, &ExprError {
					Point:    ctx.GetErrorPoint(member.Node),
					Concrete: E_FieldDoesNotExist {
						Field:  key,
						Bundle: ctx.DescribeType(AnonymousType{bundle}),
					},
				} }
				var expr = Expr {
					Type: bundle.Fields[key],
					Value: Get {
						Product: Expr(base),
						Index:   index,
					},
					Info: ExprInfo {
						ErrorPoint: ctx.GetErrorPoint(member.Node),
					},
				}
				base = expr
			case BR_BundleButOpaque:
				return Expr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_GetFromOpaqueBundle {},
				}
			case BR_NonBundle:
				return Expr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_GetFromNonBundle {},
				}
			default:
				panic("impossible branch")
			}
		}
		var final = base
		return final, nil
	case SemiTypedBundle:
		return Expr{}, &ExprError {
			Point:    ctx.GetErrorPoint(get.Base.Node),
			Concrete: E_GetFromLiteralBundle {},
		}
	default:
		return Expr{}, &ExprError {
			Point:    ctx.GetErrorPoint(get.Base.Node),
			Concrete: E_GetFromNonBundle {},
		}
	}
}

type BundleReprResult interface { BundleReprResult() }
func (impl Bundle) BundleReprResult() {}
func (impl BR_NonBundle) BundleReprResult() {}
type BR_NonBundle struct {}
func (impl BR_BundleButOpaque) BundleReprResult() {}
type BR_BundleButOpaque struct {}

func GetBundleRepr(type_ Type, ctx ExprContext) BundleReprResult {
	switch t := type_.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[t.Name]
		switch gv := g.Value.(type) {
		case Wrapped:
			var inner = FillArgs(gv.InnerType, t.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Bundle:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = t.Name.ModuleName
					if gv.IsOpaque && ctx_mod != type_mod {
						return BR_BundleButOpaque {}
					} else {
						return inner_repr
					}
				}
			}
		}
		return BR_NonBundle {}
	case AnonymousType:
		switch r := t.Repr.(type) {
		case Bundle:
			return r
		default:
			return BR_NonBundle {}
		}
	default:
		return BR_NonBundle {}
	}
}

func IsBundleLiteral(expr Expr) bool {
	switch expr.Value.(type) {
	case Product:
		switch t := expr.Type.(type) {
		case AnonymousType:
			switch t.Repr.(type) {
			case Bundle:
				return true
			}
		}
	}
	return false
}

func SemiExprFromArray(array node.Array, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(array.Node) }
	var L = len(array.Items)
	if L == 0 {
		return SemiTypedArray {
			Node:  array.Node,
			Items: make([]SemiExpr, 0),
		}, nil
	} else {
		var item_exprs = make([]SemiExpr, L)
		var item_typed_exprs = make([]Expr, L)
		var typed_count = 0
		for i, item_node := range array.Items {
			var item, err = SemiExprFrom(item_node, ctx)
			if err != nil { return nil, err }
			item_exprs[i] = item
			switch typed := item.(type) {
			case TypedExpr:
				item_typed_exprs[i] = Expr(typed)
				typed_count += 1
			}
		}
		if typed_count == L {
			var lifted, item_type, ok = LiftToMaxType(item_typed_exprs, ctx)
			if ok {
				return TypedExpr(Expr {
					Type: NamedType {
						Name: __Array,
						Args: []Type { item_type },
					},
					Value: Array { Items: lifted },
					Info:  info,
				}), nil
			} else {
				return nil, &ExprError {
					Point:    ctx.GetErrorPoint(array.Node),
					Concrete: E_HeterogeneousArray {},
				}
			}
		} else {
			return SemiTypedArray {
				Node:  array.Node,
				Items: item_exprs,
			}, nil
		}
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
