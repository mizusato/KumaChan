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

func Unbox(t Type, ctx ExprContext) UnboxResult {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var ctx_mod = ctx.GetModuleName()
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


/**
 *  Generic Template:
 *  (${Repr},${ReprBrief}) = (Tuple,T) | (Bundle,B) | (Func,F)
 *
type ${Repr}ReprResult interface { ${Repr}ReprResult() }
func (impl ${Repr}) ${Repr}ReprResult() {}
func (impl ${ReprBrief}R_Non${Repr}) ${Repr}ReprResult() {}
type ${ReprBrief}R_Non${Repr} struct {}
func (impl ${ReprBrief}R_${Repr}ButOpaque) ${Repr}ReprResult() {}
type ${ReprBrief}R_${Repr}ButOpaque struct {}

func Unbox${Repr}(t Type, ctx ExprContext) ${Repr}ReprResult {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case ${Repr}:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return ${ReprBrief}R_${Repr}ButOpaque {}
					} else {
						return inner_repr
					}
				}
			}
		}
		return ${ReprBrief}R_Non${Repr} {}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case ${Repr}:
			return r
		default:
			return ${ReprBrief}R_Non${Repr} {}
		}
	default:
		return ${ReprBrief}R_Non${Repr} {}
	}
}
*/

type TupleReprResult interface { TupleReprResult() }
func (impl Tuple) TupleReprResult() {}
func (impl TR_NonTuple) TupleReprResult() {}
type TR_NonTuple struct {}
func (impl TR_TupleButOpaque) TupleReprResult() {}
type TR_TupleButOpaque struct {}

func UnboxTuple(t Type, ctx ExprContext) TupleReprResult {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Tuple:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return TR_TupleButOpaque {}
					} else {
						return inner_repr
					}
				}
			}
		}
		return TR_NonTuple {}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Tuple:
			return r
		default:
			return TR_NonTuple {}
		}
	default:
		return TR_NonTuple {}
	}
}

type BundleReprResult interface { BundleReprResult() }
func (impl Bundle) BundleReprResult() {}
func (impl BR_NonBundle) BundleReprResult() {}
type BR_NonBundle struct {}
func (impl BR_BundleButOpaque) BundleReprResult() {}
type BR_BundleButOpaque struct {}

func UnboxBundle(t Type, ctx ExprContext) BundleReprResult {
	switch T := t.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Bundle:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = T.Name.ModuleName
					if gv.Opaque && ctx_mod != type_mod {
						return BR_BundleButOpaque {}
					} else {
						return inner_repr
					}
				}
			}
		}
		return BR_NonBundle {}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Bundle:
			return r
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
	case NamedType:
		var g = ctx.ModuleInfo.Types[T.Name]
		switch gv := g.Value.(type) {
		case Boxed:
			var inner = FillTypeArgs(gv.InnerType, T.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
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
			}
		}
		return FR_NonFunc {}
	case AnonymousType:
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
