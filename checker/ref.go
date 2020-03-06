package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
)


func (impl UntypedRef) SemiExprVal() {}
type UntypedRef struct {
	RefBody   UntypedRefBody
	TypeArgs  [] Type
}

type UntypedRefBody interface { UntypedRefBody() }
func (impl UntypedRefToType) UntypedRefBody() {}
type UntypedRefToType struct {
	TypeName  loader.Symbol
	Type      *GenericType
}
func (impl UntypedRefToFunctions) UntypedRefBody() {}
type UntypedRefToFunctions struct {
	FuncName   string
	Functions  [] *GenericFunction
}

func (impl RefConst) ExprVal() {}
type RefConst struct {
	Name  loader.Symbol
}

func (impl RefFunction) ExprVal() {}
type RefFunction struct {
	Name   string
	Index  uint
}

func (impl RefLocal) ExprVal() {}
type RefLocal struct {
	Name  string
}

func (impl NativeFunction) ExprVal() {}
type NativeFunction struct {
	Function  interface {}
}


func CheckRef(ref node.Ref, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(ref.Node) }
	var maybe_symbol = ctx.ModuleInfo.Module.SymbolFromRef(ref)
	var symbol, ok = maybe_symbol.(loader.Symbol)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(ref.Module.Node),
		Concrete: E_ModuleNotFound { loader.Id2String(ref.Module) },
	} }
	var sym_concrete, exists = ctx.LookupSymbol(symbol)
	if !exists { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(ref.Id.Node),
		Concrete: E_TypeOrValueNotFound { symbol },
	} }
	var type_ctx = ctx.GetTypeContext()
	var type_args = make([]Type, len(ref.TypeArgs))
	for i, arg_node := range ref.TypeArgs {
		var t, err = TypeFrom(arg_node.Type, type_ctx)
		if err != nil { return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(arg_node.Node),
			Concrete: E_TypeErrorInExpr { err },
		} }
		type_args[i] = t
	}
	switch s := sym_concrete.(type) {
	case SymLocalValue:
		return LiftTyped(Expr {
			Type:  s.ValueType,
			Value: RefLocal { symbol.SymbolName },
			Info:  info,
		}), nil
	case SymConst:
		return LiftTyped(Expr {
			Type:  s.Const.DeclaredType,
			Value: RefConst { symbol },
			Info:  info,
		}), nil
	case SymTypeParam:
		return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(ref.Id.Node),
			Concrete: E_TypeParamInExpr { symbol.SymbolName },
		}
	case SymType:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToType {
					TypeName: symbol,
					Type:     s.Type,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymFunctions:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToFunctions {
					FuncName:  symbol.SymbolName,
					Functions: s.Functions,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	default:
		panic("impossible branch")
	}
}

func AssignRefTo(expected Type, ref UntypedRef, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	switch r := ref.RefBody.(type) {
	case UntypedRefToType:
		return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_TypeUsedAsValue { r.TypeName },
		}
	case UntypedRefToFunctions:
		var name = r.FuncName
		var functions = r.Functions
		var type_args = ref.TypeArgs
		if len(functions) == 0 { panic("something went wrong") }
		if len(functions) == 1 {
			var f = functions[0]
			return GenericFunctionAssignTo (
				expected, name, 0, f, type_args, info, ctx,
			)
		} else {
			var errors = make([]*ExprError, 0)
			for i, f := range functions {
				var index = uint(i)
				var expr, err = GenericFunctionAssignTo (
					expected, name, index, f, type_args, info, ctx,
				)
				if err != nil {
					errors = append(errors, err)
				} else {
					return expr, nil
				}
			}
			if expected == nil {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_ExplicitTypeRequired {},
				}
			} else {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_NoneOfFunctionsAssignable {
						To:     ctx.DescribeType(expected),
						Errors: errors,
					},
				}
			}
		}
	default:
		panic("impossible branch")
	}
}

func GenericFunctionAssignTo (
	expected   Type,
	name       string,
	index      uint,
	f          *GenericFunction,
	type_args  []Type,
	info       ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	var type_arity = len(f.TypeParams)
	if type_arity == len(type_args) {
		var f_raw_type = AnonymousType { f.DeclaredType }
		var f_type = FillArgs(f_raw_type, type_args)
		var f_expr = Expr {
			Type:  f_type,
			Value: RefFunction {
				Name:  name,
				Index: index,
			},
			Info:  info,
		}
		return AssignTypedTo(expected, f_expr, ctx, true)
	} else if len(type_args) == 0 {
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		}
		var f_raw_type = AnonymousType { f.DeclaredType }
		// Note: Deep inferring is not required since function types
		//       are anonymous types. InferArgs() is capable.
		var inferred = make(map[uint]Type)
		InferArgs(f_raw_type, expected, inferred)
		if len(inferred) == type_arity {
			var args = make([]Type, type_arity)
			for i, t := range inferred {
				args[i] = t
			}
			var f_type = FillArgs(f_raw_type, args)
			if !(AreTypesEqualInSameCtx(f_type, expected)) {
				panic("something went wrong")
			}
			return Expr {
				Type:  f_type,
				Value: RefFunction {
					Name:  name,
					Index: index,
				},
				Info:  info,
			}, nil
		} else {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {
					FuncName:  name,
					TypeArity: uint(type_arity),
				},
			}
		}
	} else {
		return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_FunctionWrongTypeParamsQuantity {
				FuncName: name,
				Given:    uint(len(type_args)),
				Required: uint(type_arity),
			},
		}
	}
}
