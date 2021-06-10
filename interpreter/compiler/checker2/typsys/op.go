package typsys


func TypeOpMap(t Type, f func(t Type)(Type)) Type {
	var mapped = f(t)
	if mapped != nil {
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


