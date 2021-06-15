package typsys


func Unbox(t Type, mod string) (Type, bool) {
	var nested, is_nested = t.(*NestedType)
	if !(is_nested) { return nil, false }
	var ref, is_ref = nested.Content.(Ref)
	if !(is_ref) { return nil, false }
	var boxed, is_boxed = ref.Def.Content.(*Boxed)
	if !(is_boxed) { return nil, false }
	if boxed.BoxKind == Protected || boxed.BoxKind == Opaque {
		if mod != ref.Def.Name.ModuleName {
			return nil, false
		}
	}
	return boxed.inflatedInnerType(ref), true
}

func (boxed *Boxed) inflatedInnerType(ref Ref) Type {
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

