package checker

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

func Unbox${Repr}(type_ Type, ctx ExprContext) ${Repr}ReprResult {
	switch t := type_.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[t.Name]
		switch gv := g.Value.(type) {
		case Wrapped:
			var inner = FillArgs(gv.InnerType, t.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case ${Repr}:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = t.Name.ModuleName
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
		switch r := t.Repr.(type) {
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

func UnboxTuple(type_ Type, ctx ExprContext) TupleReprResult {
	switch t := type_.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[t.Name]
		switch gv := g.Value.(type) {
		case Wrapped:
			var inner = FillArgs(gv.InnerType, t.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Tuple:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = t.Name.ModuleName
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
		switch r := t.Repr.(type) {
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

func UnboxBundle(type_ Type, ctx ExprContext) BundleReprResult {
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

type FuncReprResult interface { FuncReprResult() }
func (impl Func) FuncReprResult() {}
func (impl FR_NonFunc) FuncReprResult() {}
type FR_NonFunc struct {}
func (impl FR_FuncButOpaque) FuncReprResult() {}
type FR_FuncButOpaque struct {}

func UnboxFunc(type_ Type, ctx ExprContext) FuncReprResult {
	switch t := type_.(type) {
	case NamedType:
		var g = ctx.ModuleInfo.Types[t.Name]
		switch gv := g.Value.(type) {
		case Wrapped:
			var inner = FillArgs(gv.InnerType, t.Args)
			switch inner_type := inner.(type) {
			case AnonymousType:
				switch inner_repr := inner_type.Repr.(type) {
				case Func:
					var ctx_mod = ctx.GetModuleName()
					var type_mod = t.Name.ModuleName
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
		switch r := t.Repr.(type) {
		case Func:
			return r
		default:
			return FR_NonFunc {}
		}
	default:
		return FR_NonFunc {}
	}
}
