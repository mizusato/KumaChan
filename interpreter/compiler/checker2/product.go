package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


type ProductPatternInfo ([] ProductPatternItemInfo)
type ProductPatternItemInfo struct {
	Binding  *LocalBinding
	Index1   uint  // 0 = whole, 1 = .0
}

func patternMatchTuple(pattern ast.PatternTuple, in typsys.Type, mod string, lm localBindingMap) (ProductPatternInfo, *source.Error) {
	var tuple, ok = unboxTuple(in, mod)
	if !(ok) {
		return nil, source.MakeError(pattern.Location,
			E_CannotMatchTuple {
				TypeName: typsys.DescribeType(in, nil),
			})
	}
	var L = len(tuple.Elements)
	var L_required = len(pattern.Names)
	if L != L_required {
		return nil, source.MakeError(pattern.Location,
			E_TupleSizeNotMatching {
				Required: uint(L_required),
				Given:    uint(L),
			})
	}
	var occurred = make(map[string] struct{})
	var info = make(ProductPatternInfo, L)
	for i := 0; i < L; i += 1 {
		var id = pattern.Names[i]
		var loc = id.Location
		var name = ast.Id2String(id)
		if name == Discarded {
			continue
		}
		var _, exists = occurred[name]
		occurred[name] = struct{}{}
		if exists {
			return nil, source.MakeError(loc,
				E_DuplicateBinding {
					BindingName: name,
				})
		}
		var t = tuple.Elements[i]
		var binding = &LocalBinding {
			Name:     name,
			Type:     t,
			Location: loc,
		}
		info[i] = ProductPatternItemInfo {
			Binding: binding,
			Index1:  uint(i + 1),
		}
	}
	for _, item := range info {
		lm.add(item.Binding)
	}
	return info, nil
}

func patternMatchRecord(pattern ast.PatternRecord, in typsys.Type, mod string, lm localBindingMap) (ProductPatternInfo, *source.Error) {
	var record, ok = unboxRecord(in, mod)
	if !(ok) {
		return nil, source.MakeError(pattern.Location,
			E_CannotMatchRecord {
				TypeName: typsys.DescribeType(in, nil),
			})
	}
	var occurred = make(map[string] struct{})
	var info = make(ProductPatternInfo, len(pattern.FieldMaps))
	for i, m := range pattern.FieldMaps {
		var binding_name = ast.Id2String(m.ValueName)
		var binding_loc = m.ValueName.Location
		var field_name = ast.Id2String(m.FieldName)
		var field_loc = m.FieldName.Location
		var _, binding_exists = occurred[binding_name]
		occurred[binding_name] = struct{}{}
		if binding_exists {
			return nil, source.MakeError(binding_loc,
				E_DuplicateBinding {
					BindingName: binding_name,
				})
		}
		var field_index, field_exists = record.FieldIndexMap[field_name]
		if field_exists {
			var field = record.Fields[field_index]
			var t = field.Type
			var binding = &LocalBinding {
				Name:     binding_name,
				Type:     t,
				Location: binding_loc,
			}
			info[i] = ProductPatternItemInfo {
				Binding: binding,
				Index1:  (1 + field_index),
			}
		} else {
			return nil, source.MakeError(field_loc,
				E_FieldNotFound {
					FieldName: field_name,
					TypeName:  typsys.DescribeType(in, nil),
				})
		}
	}
	return info, nil
}

func getTuple(t typsys.Type) (typsys.Tuple, bool) {
	var nested, is_nested = t.(*typsys.NestedType)
	if !(is_nested) { return typsys.Tuple {}, false }
	var tuple, is_tuple = nested.Content.(typsys.Tuple)
	return tuple, is_tuple
}
func getRecord(t typsys.Type) (typsys.Record, bool) {
	var nested, is_nested = t.(*typsys.NestedType)
	if !(is_nested) { return typsys.Record {}, false }
	var record, is_tuple = nested.Content.(typsys.Record)
	return record, is_tuple
}
func unboxTuple(t typsys.Type, mod string) (typsys.Tuple, bool) {
	var tuple, is_tuple = getTuple(t)
	if is_tuple {
		return tuple, true
	} else {
		var inner, exists = typsys.Unbox(t, mod)
		if exists {
			return unboxTuple(inner, mod)
		} else {
			return typsys.Tuple {}, false
		}
	}
}
func unboxRecord(t typsys.Type, mod string) (typsys.Record, bool) {
	var record, is_record = getRecord(t)
	if is_record {
		return record, true
	} else {
		var inner, exists = typsys.Unbox(t, mod)
		if exists {
			return unboxRecord(inner, mod)
		} else {
			return typsys.Record {}, false
		}
	}
}


func checkTuple(T ast.Tuple) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(T.Location, &s, ctx, nil)
		if expected == nil {
			var L = len(T.Elements)
			var elements = make([] *checked.Expr, L)
			var types = make([] typsys.Type, L)
			for i := 0; i < L; i += 1 {
				var el, err = cc.checkExpr(nil, T.Elements[i])
				if err != nil { return nil, nil, err }
				elements[i] = el
				types[i] = el.Type
			}
			var tuple_t = &typsys.NestedType {
				Content: typsys.Tuple { Elements: types },
			}
			return cc.ok(tuple_t, checked.Tuple { Elements: elements })
		} else {
			var tuple, is_tuple = getTuple(expected)
			if !(is_tuple) {
				return cc.error(
					E_TupleAssignedToIncompatible {
						TypeName: typsys.DescribeType(expected, s),
					})
			}
			var L = len(T.Elements)
			var L_required = len(tuple.Elements)
			if L != L_required {
				return cc.error(
					E_TupleSizeNotMatching {
						Required: uint(L_required),
						Given:    uint(L),
					})
			}
			var elements = make([] *checked.Expr, L)
			for i := 0; i < L; i += 1 {
				var el, err = cc.checkExpr(tuple.Elements[i], T.Elements[i])
				if err != nil { return nil, nil, err }
				elements[i] = el
			}
			var tuple_t = &typsys.NestedType { Content: tuple }
			return cc.ok(tuple_t, checked.Tuple { Elements: elements })
		}
	})
}


