package checker


func FillArgs (t Type, ctx_args []Type) Type {
	switch T := t.(type) {
	case ParameterType:
		return ctx_args[T.Index]
	case NamedType:
		var filled = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			filled[i] = FillArgs(arg, ctx_args)
		}
		return NamedType {
			Name: T.Name,
			Args: filled,
		}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return t
		case Tuple:
			var filled = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				filled[i] = FillArgs(element, ctx_args)
			}
			return AnonymousType {
				Repr: Tuple {
					Elements: filled,
				},
			}
		case Bundle:
			var filled = make(map[string]Type, len(r.Fields))
			for name, field := range r.Fields {
				filled[name] = FillArgs(field, ctx_args)
			}
			return AnonymousType {
				Repr: Bundle {
					Fields: filled,
					Index:  r.Index,
				},
			}
		case Func:
			return AnonymousType {
				Repr:Func {
					Input:  FillArgs(r.Input, ctx_args),
					Output: FillArgs(r.Output, ctx_args),
				},
			}
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}
