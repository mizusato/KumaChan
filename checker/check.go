package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/node"
	"math/big"
	"strconv"
	"strings"
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
	ArgInfCtx    TypeArgsInferringContext
	LocalValues  map[string] Type
}

type TypeArgsInferringContext struct {
	Enabled   bool
	Inferred  map[uint] Type  // mutable
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
		ArgInfCtx:   ctx.ArgInfCtx,
		LocalValues: merged,
	}, ""
}

func (ctx ExprContext) WithPatternMatching (
	input    Type,
	pattern  Pattern,
	strict   bool,
) (ExprContext, *ExprError) {
	var err_result = func(e ConcreteExprError) (ExprContext, *ExprError) {
		return ExprContext{}, &ExprError {
			Point:    pattern.Point,
			Concrete: e,
		}
	}
	var check = func(added map[string]Type) (ExprContext, *ExprError) {
		var new_ctx, shadowed = ctx.WithAddedLocalValues(added)
		if shadowed != "" && !strict {
			return err_result(E_DuplicateBinding {shadowed })
		} else {
			return new_ctx, nil
		}
	}
	switch p := pattern.Concrete.(type) {
	case TrivialPattern:
		if p.ValueName == IgnoreMark {
			if strict {
				return err_result(E_EntireValueIgnored {})
			} else {
				return ctx, nil
			}
		} else {
			var added = make(map[string]Type)
			added[p.ValueName] = input
			return check(added)
		}
	case TuplePattern:
		switch tuple := UnboxTuple(input, ctx).(type) {
		case Tuple:
			var required = len(p.ValueNames)
			var given = len(tuple.Elements)
			if given != required {
				return err_result(E_TupleSizeNotMatching {
					Required:  required,
					Given:     given,
					GivenType: ctx.DescribeType(AnonymousType { tuple }),
				})
			} else {
				var added = make(map[string]Type)
				var ignored = 0
				for i, name := range p.ValueNames {
					if name == IgnoreMark {
						ignored += 1
					} else {
						var _, exists = added[name]
						if exists {
							return ExprContext{}, &ExprError {
								Point:    p.Points[i],
								Concrete: E_DuplicateBinding { name },
							}
						}
						added[name] = tuple.Elements[i]
					}
				}
				if ignored == len(p.ValueNames) {
					return err_result(E_EntireValueIgnored {})
				} else {
					return check(added)
				}
			}
		case TR_NonTuple:
			return err_result(E_MatchingNonTupleType {})
		case TR_TupleButOpaque:
			return err_result(E_MatchingOpaqueTupleType {})
		default:
			panic("impossible branch")
		}
	case BundlePattern:
		switch bundle := UnboxBundle(input, ctx).(type) {
		case Bundle:
			var added = make(map[string]Type)
			for i, name := range p.ValueNames {
				if name == IgnoreMark { panic("something went wrong") }
				var field_type, exists = bundle.Fields[name]
				if !exists {
					return ExprContext{}, &ExprError{
						Point:    p.Points[i],
						Concrete: E_FieldDoesNotExist {
							Field:  name,
							Target: ctx.DescribeType(input),
						},
					}
				}
				_, exists = added[name]
				if exists {
					return ExprContext{}, &ExprError {
						Point:    p.Points[i],
						Concrete: E_DuplicateBinding { name },
					}
				}
				added[name] = field_type
			}
			return check(added)
		case BR_NonBundle:
			return err_result(E_MatchingNonBundleType {})
		case BR_BundleButOpaque:
			return err_result(E_MatchingOpaqueBundleType {})
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
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

func (impl UntypedLambda) SemiExprVal() {}
type UntypedLambda struct {
	Input      Pattern
	Output     node.Expr
}

func (impl UntypedInteger) SemiExprVal() {}
type UntypedInteger struct {
	Value   *big.Int
}

func (impl SemiTypedTuple) SemiExprVal() {}
type SemiTypedTuple struct {
	Values  [] SemiExpr
}

func (impl SemiTypedBundle) SemiExprVal() {}
type SemiTypedBundle struct {
	Index     map[string] uint
	Values    [] SemiExpr
	KeyNodes  [] node.Node
}

func (impl SemiTypedArray) SemiExprVal() {}
type SemiTypedArray struct {
	Items  [] SemiExpr
}

func (impl SemiTypedBlock) SemiExprVal() {}
type SemiTypedBlock struct {
	Bindings  [] Binding
	Returned  SemiExpr
}

func (impl SemiTypedMatch) SemiExprVal() {}
type SemiTypedMatch struct {
	Argument  Expr
	Branches  [] SemiTypedBranch
}
type SemiTypedBranch struct {
	IsDefault  bool
	Index      uint
	Pattern    MaybePattern
	Value      SemiExpr
}


func SemiExprFromIntLiteral(i node.IntegerLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(i.Node)
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
		return SemiExpr {
			Value: UntypedInteger { value },
			Info:  info,
		}, nil
	} else {
		return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_InvalidInteger { str },
		}
	}
}

func ExprFromFloatLiteral(f node.FloatLiteral, ctx ExprContext) (Expr, *ExprError) {
	var info = ctx.GetExprInfo(f.Node)
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
	var info = ctx.GetExprInfo(s.Node)
	return Expr {
		Type:  NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { s.Value },
	}, nil
}

func SemiExprFromTuple(tuple node.Tuple, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(tuple.Node)
	var L = len(tuple.Elements)
	if L == 0 {
		return LiftTyped(Expr {
			Type:  AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		}), nil
	} else if L == 1 {
		var expr, err = SemiExprFrom(tuple.Elements[0], ctx)
		if err != nil { return SemiExpr{}, err }
		return expr, nil
	} else {
		var el_exprs = make([]SemiExpr, L)
		var el_types = make([]Type, L)
		for i, el := range tuple.Elements {
			var expr, err = SemiExprFrom(el, ctx)
			if err != nil { return SemiExpr{}, err }
			el_exprs[i] = expr
			switch typed := expr.Value.(type) {
			case TypedExpr:
				el_types[i] = typed.Type
			}
		}
		return SemiExpr {
			Value: SemiTypedTuple { el_exprs },
			Info: info,
		}, nil
	}
}

func SemiExprFromBundle(bundle node.Bundle, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(bundle.Node)
	switch update := bundle.Update.(type) {
	case node.Update:
		var base_semi, err = SemiExprFrom(update.Base, ctx)
		if err != nil { return SemiExpr{}, err }
		switch b := base_semi.Value.(type) {
		case TypedExpr:
			if IsBundleLiteral(Expr(b)) { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			} }
			var L = len(bundle.Values)
			if !(L >= 1) { panic("something went wrong") }
			var base = Expr(b)
			switch target := UnboxBundle(base.Type, ctx).(type) {
			case Bundle:
				var occurred_names = make(map[string] bool)
				var current_base = base
				for _, field := range bundle.Values {
					var name = loader.Id2String(field.Key)
					var index, exists = target.Index[name]
					if !exists {
						return SemiExpr{}, &ExprError {
							Point: ctx.GetErrorPoint(field.Key.Node),
							Concrete: E_FieldDoesNotExist {
								Field:  name,
								Target: ctx.DescribeType(base.Type),
							},
						}
					}
					var _, duplicate = occurred_names[name]
					if duplicate {
						return SemiExpr{}, &ExprError {
							Point:    ctx.GetErrorPoint(field.Key.Node),
							Concrete: E_ExprDuplicateField { name },
						}
					}
					occurred_names[name] = true
					var value_node = DesugarOmittedFieldValue(field)
					var value_semi, err1 = SemiExprFrom(value_node, ctx)
					if err1 != nil { return SemiExpr{}, err1 }
					var field_type = target.Fields[name]
					var value, err2 = AssignSemiTo(field_type, value_semi, ctx)
					if err2 != nil { return SemiExpr{}, err2 }
					current_base = Expr {
						Type:  current_base.Type,
						Value: Set {
							Product:  current_base,
							Index:    index,
							NewValue: value,
						},
						Info:  current_base.Info,
					}
				}
				var final = current_base
				return SemiExpr {
					Value: TypedExpr(final),
					Info:  info,
				}, nil
			case BR_BundleButOpaque:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToOpaqueBundle {},
				}
			case BR_NonBundle:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToNonBundle {},
				}
			default:
				panic("impossible branch")
			}
		case SemiTypedBundle:
			return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			}
		default:
			return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(update.Base.Node),
				Concrete: E_SetToNonBundle {},
			}
		}
	default:
		var L = len(bundle.Values)
		if L == 0 {
			return LiftTyped(Expr {
				Type:  AnonymousType { Unit {} },
				Value: UnitValue {},
				Info:  info,
			}), nil
		} else {
			var f_exprs = make([]SemiExpr, L)
			var f_index_map = make(map[string]uint, L)
			var f_key_nodes = make([]node.Node, L)
			for i, field := range bundle.Values {
				var name = loader.Id2String(field.Key)
				var _, exists = f_index_map[name]
				if exists { return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(field.Key.Node),
					Concrete: E_ExprDuplicateField { name },
				} }
				var value = DesugarOmittedFieldValue(field)
				var expr, err = SemiExprFrom(value, ctx)
				if err != nil { return SemiExpr{}, err }
				f_exprs[i] = expr
				f_index_map[name] = uint(i)
				f_key_nodes[i] = field.Key.Node
			}
			return SemiExpr {
				Value: SemiTypedBundle {
					Index:    f_index_map,
					Values:   f_exprs,
					KeyNodes: f_key_nodes,
				},
				Info: info,
			}, nil
		}
	}
}

func ExprFromGet(get node.Get, ctx ExprContext) (Expr, *ExprError) {
	var base_semi, err = SemiExprFrom(get.Base, ctx)
	if err != nil { return Expr{}, err }
	switch b := base_semi.Value.(type) {
	case TypedExpr:
		if IsBundleLiteral(Expr(b)) { return Expr{}, &ExprError {
			Point:    ctx.GetErrorPoint(get.Base.Node),
			Concrete: E_GetFromLiteralBundle {},
		} }
		var L = len(get.Path)
		if !(L >= 1) { panic("something went wrong") }
		var base = Expr(b)
		for _, member := range get.Path {
			switch bundle := UnboxBundle(base.Type, ctx).(type) {
			case Bundle:
				var key = loader.Id2String(member.Name)
				var index, exists = bundle.Index[key]
				if !exists { return Expr{}, &ExprError {
					Point:    ctx.GetErrorPoint(member.Node),
					Concrete: E_FieldDoesNotExist {
						Field:  key,
						Target: ctx.DescribeType(AnonymousType{bundle}),
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


func ExprFromText(text node.Text, ctx ExprContext) (Expr, *ExprContext) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(text.Node) }
	var template = text.Template
	var segments = make([]string, 0)
	var arity = 0
	var buf strings.Builder
	for _, char := range template {
		if char == TextPlaceholder {
			var seg = buf.String()
			buf.Reset()
			segments = append(segments, seg)
			arity += 1
		} else {
			buf.WriteRune(char)
		}
	}
	var last = buf.String()
	if last != "" {
		segments = append(segments, last)
	}
	var format = func(args []string) string {
		var buf strings.Builder
		for i, seg := range segments {
			buf.WriteString(seg)
			if i < arity {
				buf.WriteString(args[i])
			}
		}
		return buf.String()
	}
	var elements = make([]Type, arity)
	for i := 0; i < arity; i += 1 {
		elements[i] = NamedType { Name: __String, Args: make([]Type, 0) }
	}
	var t Type = AnonymousType { Func {
		Input:  AnonymousType { Tuple { elements } },
		Output: NamedType { Name: __String, Args: make([]Type, 0) },
	} }
	return Expr {
		Type:  t,
		Value: NativeFunction { format },
		Info:  info,
	}, nil
}

func SemiExprFromArray(array node.Array, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(array.Node) }
	var L = len(array.Items)
	if L == 0 {
		return SemiExpr {
			Value: SemiTypedArray { make([]SemiExpr, 0) },
			Info: info,
		}, nil
	} else {
		var item_exprs = make([]SemiExpr, L)
		for i, item_node := range array.Items {
			var item, err = SemiExprFrom(item_node, ctx)
			if err != nil { return SemiExpr{}, err }
			item_exprs[i] = item
		}
		return SemiExpr {
			Value: SemiTypedArray { item_exprs },
			Info:  info,
		}, nil
	}
}

func SemiExprFromBlock(block node.Block, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(block.Node) }
	var type_ctx = ctx.GetTypeContext()
	var current_ctx = ctx
	var bindings = make([]Binding, len(block.Bindings))
	for i, b := range block.Bindings {
		var pattern = PatternFrom(b.Pattern, current_ctx)
		var t Type
		switch type_node := b.Type.(type) {
		case node.VariousType:
			var some_t, err = TypeFrom(type_node.Type, type_ctx)
			if err != nil { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(type_node.Node),
				Concrete: E_TypeErrorInExpr { err },
			}}
			t = some_t
		default:
			t = nil
		}
		if b.Recursive {
			if t == nil {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(b.Value.Node),
					Concrete: E_ExplicitTypeRequired {},
				}
			}
			var rec_ctx, err1 = current_ctx.WithPatternMatching (
				t, pattern, true,
			)
			if err1 != nil { return SemiExpr{}, err1 }
			var semi, err2 = SemiExprFrom(b.Value, rec_ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			switch semi.Value.(type) {
			case UntypedLambda:
				var typed, err = AssignSemiTo(t, semi, rec_ctx)
				if err != nil { return SemiExpr{}, err }
				bindings[i] = Binding {
					Pattern: pattern,
					Value:   typed,
				}
				current_ctx = rec_ctx
			default:
				return SemiExpr{}, &ExprError {
					Point:    semi.Info.ErrorPoint,
					Concrete: E_RecursiveMarkUsedOnNonLambda {},
				}
			}
		} else {
			var semi, err1 = SemiExprFrom(b.Value, current_ctx)
			if err1 != nil { return SemiExpr{}, err1 }
			var typed, err2 = AssignSemiTo(t, semi, current_ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			var final_t = typed.Type
			var next_ctx, err3 = current_ctx.WithPatternMatching (
				final_t, pattern, true,
			)
			if err3 != nil { return SemiExpr{}, err3 }
			bindings[i] = Binding {
				Pattern: pattern,
				Value:   typed,
			}
			current_ctx = next_ctx
		}
	}
	var ret, err = SemiExprFrom(block.Return, current_ctx)
	if err != nil { return SemiExpr{}, err }
	return SemiExpr {
		Info:  info,
		Value: SemiTypedBlock {
			Bindings: bindings,
			Returned: ret,
		},
	}, nil
}

func ExprFromCast(cast node.Cast, ctx ExprContext) (Expr, *ExprError) {
	var type_ctx = ctx.GetTypeContext()
	var target, err1 = TypeFrom(cast.Target.Type, type_ctx)
	if err1 != nil { return Expr{}, &ExprError {
		Point:    ctx.GetErrorPoint(cast.Target.Node),
		Concrete: E_TypeErrorInExpr { err1 },
	} }
	var semi, err2 = SemiExprFrom(cast.Expr, ctx)
	if err2 != nil { return Expr{}, err2 }
	var typed, err3 = AssignSemiTo(target, semi, ctx)
	if err3 != nil { return Expr{}, err3 }
	return typed, nil
}

func SemiExprFromMatch(match node.Match, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(match.Node) }
	var arg_semi, err = SemiExprFrom(match.Argument, ctx)
	if err != nil { return SemiExpr{}, err }
	var arg_typed, ok = arg_semi.Value.(TypedExpr)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(match.Argument.Node),
		Concrete: E_ExplicitTypeRequired {},
	} }
	var arg_type = arg_typed.Type
	var union, union_args, is_union = UnboxUnion(arg_type, ctx)
	if !is_union { return SemiExpr{}, &ExprError {
		Point:    arg_typed.Info.ErrorPoint,
		Concrete: E_InvalidMatchArgType {
			ArgType: ctx.DescribeType(arg_typed.Type),
		},
	} }
	var checked = make(map[loader.Symbol]bool)
	var has_default = false
	var branches = make([]SemiTypedBranch, len(match.Branches))
	for i, branch := range match.Branches {
		switch t := branch.Type.(type) {
		case node.Ref:
			if len(t.TypeArgs) > 0 {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(t.Node),
					Concrete: E_TypeParametersUnnecessary {},
				}
			}
			var maybe_type_sym = ctx.ModuleInfo.Module.SymbolFromRef(t)
			var maybe_pattern = MaybePatternFrom(branch.Pattern, ctx)
			var type_sym, ok = maybe_type_sym.(loader.Symbol)
			if !ok { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(t.Module.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ctx.GetErrorPoint(t.Module.Node),
					Concrete: E_ModuleOfTypeRefNotFound {
						Name: loader.Id2String(t.Module),
					},
				} },
			}}
			var g, exists = ctx.ModuleInfo.Types[type_sym]
			if !exists { return SemiExpr{}, &ExprError {
				Point:    ctx.GetErrorPoint(t.Node),
				Concrete: E_TypeErrorInExpr { &TypeError {
					Point:    ctx.GetErrorPoint(t.Node),
					Concrete: E_TypeNotFound {
						Name: type_sym,
					},
				} },
			} }
			var index, is_subtype = union.GetSubtypeIndex(type_sym)
			if !is_subtype { return SemiExpr{}, &ExprError{
				Point:    ctx.GetErrorPoint(t.Node),
				Concrete: E_NotSubtype{
					Union:    ctx.DescribeType(arg_type),
					TypeName: type_sym.String(),
				},
			} }
			if g.Arity != uint(len(union_args)) {
				panic("something went wrong")
			}
			var subtype = NamedType {
				Name: type_sym,
				Args: union_args,
			}
			var branch_ctx ExprContext
			switch pattern := maybe_pattern.(type) {
			case Pattern:
				var new_ctx, err = ctx.WithPatternMatching (
					subtype, pattern, false,
				)
				if err != nil { return SemiExpr{}, err }
				branch_ctx = new_ctx
			default:
				branch_ctx = ctx
			}
			var semi, err = SemiExprFrom (
				branch.Expr, branch_ctx,
			)
			if err != nil { return SemiExpr{}, err }
			branches[i] = SemiTypedBranch {
				IsDefault: false,
				Index:     index,
				Pattern:   maybe_pattern,
				Value:     semi,
			}
			checked[type_sym] = true
		default:
			if has_default {
				return SemiExpr{}, &ExprError {
					Point:    ctx.GetErrorPoint(branch.Node),
					Concrete: E_DuplicateDefaultBranch {},
				}
			}
			switch branch.Pattern.(type) {
			case node.VariousPattern:
				panic("something went wrong")
			}
			var semi, err = SemiExprFrom(branch.Expr, ctx)
			if err != nil { return SemiExpr{}, nil }
			branches[i] = SemiTypedBranch {
				IsDefault: true,
				Index:     -1,
				Pattern:   Pattern {},
				Value:     semi,
			}
			has_default = true
		}
	}
	if !has_default && len(checked) != len(union.SubTypes) {
		var missing = make([]string, 0)
		for _, subtype := range union.SubTypes {
			if !checked[subtype] {
				missing = append(missing, subtype.String())
			}
		}
		return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(match.Node),
			Concrete: E_IncompleteMatch { missing },
		}
	} else {
		return SemiExpr {
			Value: SemiTypedMatch {
				Argument: Expr(arg_typed),
				Branches: branches,
			},
			Info: info,
		}, nil
	}
}

func SemiExprFromIf(if_node node.If, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(if_node.Node) }
	var cond_semi, err = SemiExprFrom(if_node.Condition, ctx)
	if err != nil { return SemiExpr{}, err }
	var cond_typed, ok = cond_semi.Value.(TypedExpr)
	if !ok { return SemiExpr{}, &ExprError{
		Point:    ctx.GetErrorPoint(if_node.Condition.Node),
		Concrete: E_NonBooleanCondition { Typed:false },
	} }
	switch T := cond_typed.Type.(type) {
	case NamedType:
		if T.Name == __Bool {
			if len(T.Args) != 0 { panic("something went wrong") }
			var yes_semi, err1 = SemiExprFrom(if_node.YesBranch, ctx)
			if err1 != nil { return SemiExpr{}, err1 }
			var yes_branch = SemiTypedBranch {
				IsDefault: false,
				Index:     __Yes,
				Pattern:   nil,
				Value:     yes_semi,
			}
			var no_semi, err2 = SemiExprFrom(if_node.NoBranch, ctx)
			if err2 != nil { return SemiExpr{}, err2 }
			var no_branch = SemiTypedBranch {
				IsDefault: true,
				Index:     -1,
				Pattern:   nil,
				Value:     no_semi,
			}
			return SemiExpr {
				Info: info,
				Value: SemiTypedMatch {
					Argument: Expr(cond_typed),
					Branches: []SemiTypedBranch {
						yes_branch, no_branch,
					},
				},
			}, nil
		}
	}
	return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(if_node.Condition.Node),
		Concrete: E_NonBooleanCondition {
			Typed: true,
			Type:  ctx.DescribeType(cond_typed.Type),
		},
	}
}


func PatternFrom(p_node node.VariousPattern, ctx ExprContext) Pattern {
	switch p := p_node.Pattern.(type) {
	case node.PatternTrivial:
		return Pattern {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: TrivialPattern {
				ValueName: loader.Id2String(p.Name),
				Point:     ctx.GetErrorPoint(p.Name.Node),
			},
		}
	case node.PatternTuple:
		var names = make([]string, len(p.Names))
		var points = make([]ErrorPoint, len(p.Names))
		for i, identifier := range p.Names {
			names[i] = loader.Id2String(identifier)
			points[i] = ctx.GetErrorPoint(p.Names[i].Node)
		}
		return Pattern {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: TuplePattern {
				ValueNames: names,
				Points:     points,
			},
		}
	case node.PatternBundle:
		var names = make([]string, len(p.Names))
		var points = make([]ErrorPoint, len(p.Names))
		for i, identifier := range p.Names {
			names[i] = loader.Id2String(identifier)
			points[i] = ctx.GetErrorPoint(p.Names[i].Node)
		}
		return Pattern{
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: BundlePattern {
				ValueNames: names,
				Points:     points,
			},
		}
	default:
		panic("impossible branch")
	}
}

func MaybePatternFrom(p node.MaybePattern, ctx ExprContext) MaybePattern {
	switch p_node := p.(type) {
	case node.VariousPattern:
		return PatternFrom(p_node, ctx)
	default:
		return nil
	}
}


func SemiExprFrom(e node.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	// TODO
	return SemiExpr{}, nil
}

func SemiExprFromTerm(term node.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	// TODO
	return SemiExpr{}, nil
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
						Module:   node.Identifier {
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
