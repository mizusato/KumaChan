package typsys


func Unbox(t Type, mod string) (Type, *Box, bool) {
	var nested, is_nested = t.(*NestedType)
	if !(is_nested) { return nil, nil, false }
	var ref, is_ref = nested.Content.(Ref)
	if !(is_ref) { return nil, nil, false }
	var box, is_box = ref.Def.Content.(*Box)
	if !(is_box) { return nil, nil, false }
	if box.BoxKind == Protected || box.BoxKind == Opaque {
		if mod != ref.Def.Name.ModuleName {
			return nil, nil, false
		}
	}
	return box.inflatedInnerType(ref), box, true
}

func (boxed *Box) inflatedInnerType(ref Ref) Type {
	if ref.Def.Content != boxed {
		panic("invalid argument")
	}
	if len(ref.Args) != len(ref.Def.Parameters) {
		panic("invalid argument")
	}
	return TypeOpMap(boxed.InnerType, func(t Type) (Type, bool) {
		var T, is_param = t.(ParameterType)
		if is_param {
			for i, arg := range ref.Args {
				var p = &(ref.Def.Parameters[i])
				if p == T.Parameter {
					return arg, true
				}
			}
		}
		return nil, false
	})
}


