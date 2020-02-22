package checker


func FillArgs(t Type, ctx_args []Type) Type {
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

func InferArgs(template Type, given Type, inferred map[uint]Type) {
	switch T := template.(type) {
	case ParameterType:
		var existing, exists = inferred[T.Index]
		if !exists || AreTypesEqualInSameCtx(existing, given) {
			inferred[T.Index] = given
		}
	case NamedType:
		switch G := given.(type) {
		case NamedType:
			var L1 = len(T.Args)
			var L2 = len(G.Args)
			if L1 != L2 { panic("type registration went wrong") }
			var L = L1
			for i := 0; i < L; i += 1 {
				InferArgs(T.Args[i], G.Args[i], inferred)
			}
		}
	case AnonymousType:
		switch G := given.(type) {
		case AnonymousType:
			switch Tr := T.Repr.(type) {
			case Tuple:
				switch Gr := G.Repr.(type) {
				case Tuple:
					var L1 = len(Tr.Elements)
					var L2 = len(Gr.Elements)
					if L1 == L2 {
						var L = L1
						for i := 0; i < L; i += 1 {
							InferArgs(Tr.Elements[i], Gr.Elements[i], inferred)
						}
					}
				}
			case Bundle:
				switch Gr := G.Repr.(type) {
				case Bundle:
					for name, Tf := range Tr.Fields {
						var Gf, exists = Gr.Fields[name]
						if exists {
							InferArgs(Tf, Gf, inferred)
						}
					}
				}
			case Func:
				switch Gr := G.Repr.(type) {
				case Func:
					InferArgs(Tr.Input, Gr.Input, inferred)
					InferArgs(Tr.Output, Gr.Output, inferred)
				}
			default:
				panic("impossible branch")
			}
		}
	default:
		panic("impossible branch")
	}
}
