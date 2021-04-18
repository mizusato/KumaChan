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
		switch gv := g.Definition.(type) {
		case *Boxed:
			var type_mod = T.Name.ModuleName
			if gv.Opaque && ctx_mod != type_mod {
				return UnboxedButOpaque {}
			} else {
				return Unboxed {
					Type: FillTypeArgsWithDefaults(gv.InnerType, T.Args, g.Defaults),
				}
			}
		}
	}
	return UnboxFailed {}
}

func UnboxWeak(t Type, reg TypeRegistry) Type {
	switch T := t.(type) {
	case *NamedType:
		var g = reg[T.Name]
		switch gv := g.Definition.(type) {
		case *Boxed:
			if gv.Weak {
				var filled_inner = FillTypeArgsWithDefaults(gv.InnerType, T.Args, g.Defaults)
				return UnboxWeak(filled_inner, reg)
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

func UnboxTuple(t Type, ctx ExprContext, cross_reactive bool) TupleReprResult {
	switch T := t.(type) {
	case *NamedType:
		if cross_reactive && IsReactive(T) {
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
		switch gv := g.Definition.(type) {
		case *Boxed:
			var inner = FillTypeArgsWithDefaults(gv.InnerType, T.Args, g.Defaults)
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
				return UnboxTuple(inner, ctx, cross_reactive)
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

type RecordReprResult interface { RecordReprResult() }
func (impl BR_Record) RecordReprResult() {}
type BR_Record struct {
	Record          Record
	AcrossReactive  bool
}
func (impl BR_NonRecord) RecordReprResult() {}
type BR_NonRecord struct {}
func (impl BR_RecordButOpaque) RecordReprResult() {}
type BR_RecordButOpaque struct {}

func UnboxRecord(t Type, ctx ExprContext, cross_reactive bool) RecordReprResult {
	switch T := t.(type) {
	case *NamedType:
		if cross_reactive && IsReactive(T) {
			if !(len(T.Args) == 1) { panic("something went wrong") }
			var result = UnboxRecord(T.Args[0], ctx, false)
			switch r := result.(type) {
			case BR_Record:
				return BR_Record {
					Record:         r.Record,
					AcrossReactive: true,
				}
			default:
				return result
			}
		}
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Definition.(type) {
		case *Boxed:
			var inner = FillTypeArgsWithDefaults(gv.InnerType, T.Args, g.Defaults)
			switch inner_type := inner.(type) {
			case *AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Record:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return BR_RecordButOpaque {}
					} else {
						return BR_Record { Record: inner_repr }
					}
				}
			case *NamedType:
				return UnboxRecord(inner, ctx, cross_reactive)
			}
		}
		return BR_NonRecord {}
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Record:
			return BR_Record { Record: r }
		default:
			return BR_NonRecord {}
		}
	default:
		return BR_NonRecord {}
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
		switch gv := g.Definition.(type) {
		case *Boxed:
			var inner = FillTypeArgsWithDefaults(gv.InnerType, T.Args, g.Defaults)
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
				default:
					return FR_NonFunc {}
				}
			case *NamedType:
				return UnboxFunc(inner, ctx)
			default:
				return FR_NonFunc {}
			}
		default:
			return FR_NonFunc {}
		}
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

