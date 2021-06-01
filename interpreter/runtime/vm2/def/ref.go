package def


func BranchRef(enum EnumValue, idx ShortIndex) NativeFuncValue {
	return ValNativeFunc(func(arg Value, _ InteropContext) Value {
		var new_value, update = Unwrap(arg.(EnumValue))
		if update {
			return Tuple(&ValEnum {
				Index: idx,
				Value: new_value,
			}, arg)
		} else {
			if enum.Index == idx {
				return Tuple(enum, Some(enum.Value))
			} else {
				return Tuple(enum, None())
			}
		}
	})
}

func BranchRefFromCaseRef(base_ref Value, idx ShortIndex) NativeFuncValue {
	return ValNativeFunc(func(arg Value, h InteropContext) Value {
		var t = h.Call(base_ref, None())
		var pair = t.(TupleValue).Elements
		var base_enum = pair[0]
		var base_branch = pair[1]
		var value, has_value = Unwrap(base_branch.(EnumValue))
		var new_value, update = Unwrap(arg.(EnumValue))
		if has_value {
			if update {
				var u = h.Call(base_ref, Some(&ValEnum {
					Index: idx,
					Value: new_value,
				}))
				return Tuple(u.(TupleValue).Elements[0], arg)
			} else {
				var enum = value.(EnumValue)
				if enum.Index == idx {
					return Tuple(base_enum, Some(enum.Value))
				} else {
					return Tuple(base_enum, None())
				}
			}
		} else {
			return Tuple(base_enum, None())
		}
	})
}

func BranchRefFromProjRef(base_ref Value, idx ShortIndex) NativeFuncValue {
	return ValNativeFunc(func(arg Value, h InteropContext) Value {
		var new_value, update = Unwrap(arg.(EnumValue))
		if update {
			var t = h.Call(base_ref, Some(&ValEnum {
				Index: idx,
				Value: new_value,
			}))
			return Tuple(t.(TupleValue).Elements[0], arg)
		} else {
			var value = h.Call(base_ref, None())
			var pair = value.(TupleValue).Elements
			var base_tup = pair[0]
			var base_field = pair[1]
			var enum = base_field.(EnumValue)
			if enum.Index == idx {
				return Tuple(base_tup, Some(enum.Value))
			} else {
				return Tuple(base_tup, None())
			}
		}
	})
}

func FieldRef(tuple TupleValue, idx ShortIndex) NativeFuncValue {
	return ValNativeFunc(func(arg Value, _ InteropContext) Value {
		var new_value, update = Unwrap(arg.(EnumValue))
		if update {
			var draft =  make([] Value, len(tuple.Elements))
			copy(draft, tuple.Elements)
			draft[idx] = new_value
			return Tuple(TupleOf(draft), new_value)
		} else {
			return Tuple(tuple, tuple.Elements[idx])
		}
	})
}

func FieldRefFromProjRef(base_ref Value, idx ShortIndex) NativeFuncValue {
	return ValNativeFunc(func(arg Value, h InteropContext) Value {
		var t = h.Call(base_ref, None())
		var tup = t.(TupleValue).Elements[1].(TupleValue)
		var new_field_value, update = Unwrap(arg.(EnumValue))
		if update {
			var draft =  make([] Value, len(tup.Elements))
			copy(draft, tup.Elements)
			draft[idx] = new_field_value
			var new_tup_value = TupleOf(draft)
			var t = h.Call(base_ref, Some(new_tup_value))
			var new_base_value = t.(TupleValue).Elements[0]
			return Tuple(new_base_value, new_field_value)
		} else {
			return Tuple(tup, tup.Elements[idx])
		}
	})
}

