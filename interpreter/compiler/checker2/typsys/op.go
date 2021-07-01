package typsys


func TypeOpMap(t Type, f func(t Type)(Type,bool)) Type {
	var mapped, ok = f(t)
	if ok {
		return mapped
	} else {
		var nested, is_nested = t.(*NestedType)
		if is_nested {
			switch N := nested.Content.(type) {
			case Ref:
				var mapped_args = make([] Type, len(N.Args))
				for i, arg := range N.Args {
					mapped_args[i] = TypeOpMap(arg, f)
				}
				return &NestedType { Ref {
					Def:  N.Def,
					Args: mapped_args,
				} }
			case Tuple:
				var mapped_elements = make([] Type, len(N.Elements))
				for i, e := range N.Elements {
					mapped_elements[i] = TypeOpMap(e, f)
				}
				return &NestedType { Tuple { mapped_elements } }
			case Record:
				var mapped_fields = make([] Field, len(N.Fields))
				for i, field := range N.Fields {
					var field_t = field.Type
					mapped_fields[i] = Field {
						Attr: field.Attr,
						Name: field.Name,
						Type: TypeOpMap(field_t, f),
					}
				}
				return &NestedType { Record {
					FieldIndexMap: N.FieldIndexMap,
					Fields:        mapped_fields,
				} }
			case Lambda:
				var input_t = N.Input
				var output_t = N.Output
				return &NestedType { Lambda {
					Input:  TypeOpMap(input_t, f),
					Output: TypeOpMap(output_t, f),
				} }
			default:
				panic("impossible branch")
			}
		} else {
			return t
		}
	}
}

func TypeOpEqual(t1 Type, t2 Type) bool {
	var _, t1_is_unknown = t1.(*UnknownType)
	var _, t2_is_unknown = t2.(*UnknownType)
	if t1_is_unknown || t2_is_unknown {
		return false
	} else if t1 == t2 {
		return true
	} else {
		var t1_nested, t1_is_nested = t1.(*NestedType)
		var t2_nested, t2_is_nested = t2.(*NestedType)
		if t1_is_nested && t2_is_nested {
			switch N1 := t1_nested.Content.(type) {
			case Ref:
				var N2, ok = t2_nested.Content.(Ref)
				if !(ok) { return false }
				if N1.Def == N2.Def {
					if len(N1.Args) != len(N2.Args) {
						panic("something went wrong")
					}
					return TypeOpEqualTypeVec(N1.Args, N2.Args)
				} else {
					return false
				}
			case Tuple:
				var N2, ok = t2_nested.Content.(Tuple)
				if !(ok) { return false }
				return TypeOpEqualTypeVec(N1.Elements, N2.Elements)
			case Record:
				var N2, ok = t2_nested.Content.(Record)
				if !(ok) { return false }
				return TypeOpEqualFieldVec(N1.Fields, N2.Fields)
			case Lambda:
				var N2, ok = t2_nested.Content.(Lambda)
				if !(ok) { return false }
				var io1 = [2] Type { N1.Input, N1.Output }
				var io2 = [2] Type { N2.Input, N2.Output }
				return TypeOpEqualTypeVec(io1[:], io2[:])
			default:
				panic("impossible branch")
			}
		} else {
			return false
		}
	}
}

func TypeOpEqualTypeVec(u ([] Type), v ([] Type)) bool {
	if len(u) == len(v) {
		var L = len(u)
		for i := 0; i < L; i += 1 {
			var a = u[i]
			var b = v[i]
			var eq = TypeOpEqual(a, b)
			if !(eq) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func TypeOpEqualFieldVec(u ([] Field), v ([] Field)) bool {
	if len(u) == len(v) {
		var L = len(u)
		for i := 0; i < L; i += 1 {
			var fa = u[i]
			var fb = v[i]
			var name_eq = (fa.Name == fb.Name)
			if !(name_eq) {
				return false
			}
			var type_eq = TypeOpEqual(fa.Type, fb.Type)
			if !(type_eq) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}


