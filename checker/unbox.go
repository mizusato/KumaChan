package checker


type UnboxResult interface { UnboxResult() }
func (impl Unboxed) UnboxResult() {}
type Unboxed struct {
	Type  Type
}
func (impl UnboxedButOpaque) UnboxResult() {}
type UnboxedButOpaque struct {}
func (impl UnboxFailed) UnboxResult() {}
type UnboxFailed struct {}

func Unbox(t Type, ctx_mod string, reg TypeRegistry) UnboxResult {
	switch T := t.(type) {
	case *NamedType:
		var g = reg[T.Name]
		switch gv := g.Value.(type) {
		case *Boxed:
			var type_mod = T.Name.ModuleName
			if gv.Opaque && ctx_mod != type_mod {
				return UnboxedButOpaque {}
			} else {
				return Unboxed {
					Type: FillTypeArgs(gv.InnerType, T.Args),
				}
			}
		}
	}
	return UnboxFailed {}
}

func UnboxAsIs(t Type, reg TypeRegistry) Type {
	switch T := t.(type) {
	case *NamedType:
		var g = reg[T.Name]
		switch gv := g.Value.(type) {
		case *Boxed:
			if gv.Weak {
				var filled_inner = FillTypeArgs(gv.InnerType, T.Args)
				return UnboxAsIs(filled_inner, reg)
			}
		}
	}
	return t
}

type TupleReprResult interface { TupleReprResult() }
func (impl TR_Tuple) TupleReprResult() {}
type TR_Tuple struct {
	Tuple           Tuple
	AcrossReactive  bool
}
func (impl TR_NonTuple) TupleReprResult() {}
type TR_NonTuple struct {}
func (impl TR_TupleButOpaque) TupleReprResult() {}
type TR_TupleButOpaque struct {}

func UnboxTuple(t Type, ctx ExprContext, across_reactive bool) TupleReprResult {
	switch T := t.(type) {
	case *NamedType:
		if across_reactive && T.Name == __Reactive {
			if !(len(T.Args) == 1) { panic("something went wrong") }
			var result = UnboxTuple(T.Args[0], ctx, false)
			switch r := result.(type) {
			case TR_Tuple:
				return TR_Tuple {
					Tuple:          r.Tuple,
					AcrossReactive: true,
				}
			default:
				return result
			}
		}
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case *Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case *AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Tuple:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return TR_TupleButOpaque {}
					} else {
						return TR_Tuple { Tuple: inner_repr }
					}
				}
			case *NamedType:
				return UnboxTuple(inner, ctx, false)
			}
		}
		return TR_NonTuple {}
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Tuple:
			return TR_Tuple { Tuple: r }
		default:
			return TR_NonTuple {}
		}
	default:
		return TR_NonTuple {}
	}
}

type BundleReprResult interface { BundleReprResult() }
func (impl BR_Bundle) BundleReprResult() {}
type BR_Bundle struct {
	Bundle          Bundle
	AcrossReactive  bool
}
func (impl BR_NonBundle) BundleReprResult() {}
type BR_NonBundle struct {}
func (impl BR_BundleButOpaque) BundleReprResult() {}
type BR_BundleButOpaque struct {}

func UnboxBundle(t Type, ctx ExprContext, across_reactive bool) BundleReprResult {
	switch T := t.(type) {
	case *NamedType:
		if across_reactive && T.Name == __Reactive {
			if !(len(T.Args) == 1) { panic("something went wrong") }
			var result = UnboxBundle(T.Args[0], ctx, false)
			switch r := result.(type) {
			case BR_Bundle:
				return BR_Bundle {
					Bundle:         r.Bundle,
					AcrossReactive: true,
				}
			default:
				return result
			}
		}
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case *Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case *AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Bundle:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return BR_BundleButOpaque {}
					} else {
						return BR_Bundle { Bundle: inner_repr }
					}
				}
			case *NamedType:
				return UnboxBundle(inner, ctx, false)
			}
		}
		return BR_NonBundle {}
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Bundle:
			return BR_Bundle { Bundle: r }
		default:
			return BR_NonBundle {}
		}
	default:
		return BR_NonBundle {}
	}
}

type FuncReprResult interface { FuncReprResult() }
func (impl Func) FuncReprResult() {}
func (impl FR_NonFunc) FuncReprResult() {}
type FR_NonFunc struct {}
func (impl FR_FuncButOpaque) FuncReprResult() {}
type FR_FuncButOpaque struct {}

func UnboxFunc(t Type, ctx ExprContext) FuncReprResult {
	switch T := t.(type) {
	case *NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case *Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case *AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Func:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return FR_FuncButOpaque {}
					} else {
						return inner_repr
					}
				}
			case *NamedType:
				UnboxFunc(inner, ctx)
			}
		}
		return FR_NonFunc {}
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Func:
			return r
		default:
			return FR_NonFunc {}
		}
	default:
		return FR_NonFunc {}
	}
}

