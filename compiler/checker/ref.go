package checker

import (
	"kumachan/lang"
	"kumachan/lang/parser/ast"
	. "kumachan/misc/util/error"
)

// TODO: refactor this file


func (impl UntypedRef) SemiExprVal() {}
type UntypedRef struct {
	RefBody   UntypedRefBody
	TypeArgs  [] Type
}

type UntypedRefBody interface { UntypedRefBody() }
func (impl UntypedRefToType) UntypedRefBody() {}
type UntypedRefToType struct {
	TypeName    lang.Symbol
	Type        *GenericType
	ForceExact  bool
}
func (impl UntypedRefToFunctions) UntypedRefBody() {}
type UntypedRefToFunctions struct {
	FuncName    string
	Functions   [] SymFunctionReference
	TypeExists  bool
	RefToType   UntypedRefToType
}
func (impl UntypedRefToFunctionsAndLocalValue) UntypedRefBody() {}
type UntypedRefToFunctionsAndLocalValue struct {
	RefToFunctions   UntypedRefToFunctions
	RefToLocalValue  Expr
}

type Ref interface { ExprVal; Ref() }

func (impl RefConstant) ExprVal() {}
func (impl RefConstant) Ref() {}
type RefConstant struct {
	Name  lang.Symbol
}

func (impl RefFunction) ExprVal() {}
func (impl RefFunction) Ref() {}
type RefFunction struct {
	Name      string
	Index     uint
	AbsRef    AbsRefFunction
	Implicit  [] Ref
}
type AbsRefFunction struct {
	Module  string
	Name    string
	Index   uint
}
func MakeRefFunction(name string, index uint, type_args ([] Type), node ast.Node, ctx ExprContext) (RefFunction, *ExprError) {
	var raw_ref = ctx.ModuleInfo.Functions[name][index]
	var abs_ref = AbsRefFunction {
		Module: raw_ref.ModuleName,
		Name:   name,
		Index:  raw_ref.Index,
	}
	var f = raw_ref.Function
	var err = CheckTypeArgsBounds(type_args, f.TypeParams, nil, f.TypeBounds, node, ctx)
	if err != nil { return RefFunction{}, err }
	var implicit_count = uint(len(f.Implicit))
	var implicit_refs = make([] Ref, implicit_count)
	if implicit_count > 0 {
		var wrap_error = func(name string, err *ExprError) *ExprError {
			return &ExprError {
				Point:    ErrorPointFrom(node),
				Concrete: E_ImplicitContextNotFound {
					Name:   name,
					Detail: err,
				},
			}
		}
		for name, field := range f.Implicit {
			var field_t = FillTypeArgs(field.Type, type_args)
			var ref = CraftAstRef(name, node)
			var ref_term = ast.VariousTerm {
				Node: node,
				Term: ref,
			}
			var ref_semi, err1 = CheckTerm(ref_term, ctx)
			if err1 != nil { return RefFunction{}, wrap_error(name, err1) }
			var ref_expr, err2 = AssignTo(field_t, ref_semi, ctx)
			if err2 != nil { return RefFunction{}, wrap_error(name, err2) }
			var ref_val = ref_expr.Value.(Ref)
			implicit_refs[field.Index] = ref_val
		}
	}
	return RefFunction {
		Name:     name,
		Index:    index,
		AbsRef:   abs_ref,
		Implicit: implicit_refs,
	}, nil
}
func CraftAstRef(name string, node ast.Node) ast.InlineRef {
	return ast.InlineRef {
		Node:     node,
		Module:   ast.Identifier {
			Node: node,
			Name: [] rune {},
		},
		Id:       ast.Identifier {
			Node: node,
			Name: ([] rune)(name),
		},
		TypeArgs: make([] ast.VariousType, 0),
	}
}
func CraftAstRefTerm(name string, node ast.Node) ast.VariousTerm {
	return ast.VariousTerm {
		Node: node,
		Term: CraftAstRef(name, node),
	}
}

func (impl RefLocal) ExprVal() {}
func (impl RefLocal) Ref() {}
type RefLocal struct {
	Name  string
}


func CheckRef(ref ast.InlineRef, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(ref.Node)
	var maybe_symbol = ctx.ModuleInfo.Module.SymbolFromInlineRef(ref)
	var symbol, ok = maybe_symbol.(lang.Symbol)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ErrorPointFrom(ref.Module.Node),
		Concrete: E_ModuleNotFound { ast.Id2String(ref.Module) },
	} }
	var sym_concrete, exists = ctx.LookupSymbol(symbol)
	if !exists { return SemiExpr{}, &ExprError {
		Point:    ErrorPointFrom(ref.Id.Node),
		Concrete: E_TypeOrValueNotFound { symbol },
	} }
	var type_ctx = ctx.GetTypeContext()
	var type_args = make([] Type, len(ref.TypeArgs))
	for i, arg_node := range ref.TypeArgs {
		var t, err = TypeFrom(arg_node, type_ctx)
		if err != nil { return SemiExpr{}, &ExprError {
			Point:    err.Point,
			Concrete: E_TypeErrorInExpr { err },
		} }
		type_args[i] = t
	}
	switch s := sym_concrete.(type) {
	case SymLocalValue:
		if len(type_args) > 0 { return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_SuperfluousTypeArgs {},
		} }
		return LiftTyped(Expr {
			Type:  s.ValueType,
			Value: RefLocal { symbol.SymbolName },
			Info:  info,
		}), nil
	case SymTypeParam:
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(ref.Id.Node),
			Concrete: E_TypeParamInExpr { symbol.SymbolName },
		}
	case SymType:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToType {
					TypeName:   s.Name,
					Type:       s.Type,
					ForceExact: s.ForceExact,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymFunctions:
		var ref_to_type UntypedRefToType
		var type_exists = s.TypeExists
		if type_exists {
			ref_to_type = UntypedRefToType {
				TypeName:   s.TypeSym.Name,
				Type:       s.TypeSym.Type,
				ForceExact: s.TypeSym.ForceExact,
			}
		}
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToFunctions {
					FuncName:   s.Name,
					Functions:  s.Functions,
					TypeExists: type_exists,
					RefToType:  ref_to_type,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymLocalAndFunc:
		var ref_to_type UntypedRefToType
		var type_exists = s.Func.TypeExists
		if type_exists {
			ref_to_type = UntypedRefToType {
				TypeName:   s.Func.TypeSym.Name,
				Type:       s.Func.TypeSym.Type,
				ForceExact: s.Func.TypeSym.ForceExact,
			}
		}
		if len(type_args) > 0 {
			return SemiExpr {
				Value: UntypedRef {
					RefBody:  UntypedRefToFunctions {
						FuncName:   s.Func.Name,
						Functions:  s.Func.Functions,
						TypeExists: type_exists,
						RefToType:  ref_to_type,
					},
					TypeArgs: type_args,
				},
				Info:  info,
			}, nil
		} else {
			return SemiExpr {
				Value: UntypedRef {
					RefBody:  UntypedRefToFunctionsAndLocalValue {
						RefToFunctions:  UntypedRefToFunctions {
							FuncName:   s.Func.Name,
							Functions:  s.Func.Functions,
							TypeExists: type_exists,
							RefToType:  ref_to_type,
						},
						RefToLocalValue: Expr {
							Type:  s.Local.ValueType,
							Value: RefLocal { symbol.SymbolName },
							Info:  info,
						},
					},
					TypeArgs: type_args,
				},
				Info:  info,
			}, nil
		}
	default:
		panic("impossible branch")
	}
}


func AssignRefTo(expected Type, ref UntypedRef, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	switch r := ref.RefBody.(type) {
	case UntypedRefToType:
		var g = r.Type
		var g_name = r.TypeName
		var force_exact = r.ForceExact
		var type_args = ref.TypeArgs
		var unit = LiftTyped(Expr {
			Type:  &AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		})
		var boxed_unit, err = Box (
			unit, g, g_name, info, type_args,
			force_exact, info, ctx,
		)
		if err != nil { return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_TypeUsedAsValue { r.TypeName },
		} }
		return TypedAssignTo(expected, boxed_unit, ctx)
	case UntypedRefToFunctions:
		var name = r.FuncName
		var functions = r.Functions
		var type_args = ref.TypeArgs
		return OverloadedAssignTo (
			expected, functions, name, type_args, info, ctx,
		)
	case UntypedRefToFunctionsAndLocalValue:
		var local = r.RefToLocalValue
		var expr, err = TypedAssignTo(expected, local, ctx)
		if err == nil {
			return expr, nil
		} else {
			var functions = UntypedRef {
				RefBody:  r.RefToFunctions,
				TypeArgs: ref.TypeArgs,
			}
			var expr_f, err_f = AssignRefTo(expected, functions, info, ctx)
			if err_f != nil {
				return Expr{}, err  // throw the error of local value
			}
			return expr_f, nil
		}
	default:
		panic("impossible branch")
	}
}


func CallUntypedRef (
	arg        SemiExpr,
	ref        UntypedRef,
	ref_info   ExprInfo,
	call_info  ExprInfo,
	ctx        ExprContext,
) (SemiExpr, *ExprError) {
	switch ref_body := ref.RefBody.(type) {
	case UntypedRefToType:
		var g = ref_body.Type
		var g_name = ref_body.TypeName
		var force_exact = ref_body.ForceExact
		var type_args = ref.TypeArgs
		var typed, err = Box (
			arg, g, g_name, ref_info, type_args,
			force_exact, call_info, ctx,
		)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(typed), nil
	case UntypedRefToFunctions:
		var functions = ref_body.Functions
		var name = ref_body.FuncName
		var type_args = ref.TypeArgs
		var semi, err = OverloadedCall (
			functions, name, type_args,
			arg, ref_info, call_info, ctx,
		)
		if err != nil {
			if ref_body.TypeExists {
				var rt = ref_body.RefToType
				var expr, box_err = Box (
					arg, rt.Type, rt.TypeName, ref_info, type_args,
					rt.ForceExact, call_info, ctx,
				)
				if box_err != nil { return SemiExpr{}, err }
				return LiftTyped(expr), nil
			} else {
				return SemiExpr{}, err
			}
		}
		return semi, nil
	case UntypedRefToFunctionsAndLocalValue:
		var local = ref_body.RefToLocalValue
		var _, is_func = UnboxFunc(local.Type, ctx).(Func)
		if is_func {
			var expr, err = CallTyped(local, arg, call_info, ctx)
			if err != nil { return SemiExpr{}, err }
			return LiftTyped(expr), nil
		} else {
			var functions = UntypedRef {
				RefBody:  ref_body.RefToFunctions,
				TypeArgs: ref.TypeArgs,
			}
			return CallUntypedRef(arg, functions, ref_info, call_info, ctx)
		}
	default:
		panic("impossible branch")
	}
}
