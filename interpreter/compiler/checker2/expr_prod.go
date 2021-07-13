package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


type ProductPatternMatching func
	(in typsys.Type, mod string, lm LocalBindingMap) (
	checked.ProductPatternInfo, *source.Error)

func productPatternMatch(pattern ast.VariousPattern) ProductPatternMatching {
	switch P := pattern.Pattern.(type) {
	case ast.PatternTrivial:
		return patternMatchTrivial(P)
	case ast.PatternTuple:
		return patternMatchTuple(P)
	case ast.PatternRecord:
		return patternMatchRecord(P)
	default:
		panic("impossible branch")
	}
}

func patternMatchTrivial(pattern ast.PatternTrivial) ProductPatternMatching {
	return ProductPatternMatching(func(in typsys.Type, mod string, lm LocalBindingMap) (checked.ProductPatternInfo, *source.Error) {
		in = unboxWeak(in, mod)
		var binding = &checked.LocalBinding {
			Name:     ast.Id2String(pattern.Name),
			Type:     in,
			Location: pattern.Location,
		}
		lm.add(binding)
		return checked.ProductPatternInfo([] checked.ProductPatternItemInfo { {
			Binding: binding,
			Index1:  0,
		}}), nil
	})
}

func patternMatchTuple(pattern ast.PatternTuple) ProductPatternMatching {
	return ProductPatternMatching(func(in typsys.Type, mod string, lm LocalBindingMap) (checked.ProductPatternInfo, *source.Error) {
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
		var info = make(checked.ProductPatternInfo, L)
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
			var binding = &checked.LocalBinding {
				Name:     name,
				Type:     t,
				Location: loc,
			}
			info[i] = checked.ProductPatternItemInfo {
				Binding: binding,
				Index1:  uint(i + 1),
			}
		}
		for _, item := range info {
			lm.add(item.Binding)
		}
		return info, nil
	})
}

func patternMatchRecord(pattern ast.PatternRecord) ProductPatternMatching {
	return ProductPatternMatching(func(in typsys.Type, mod string, lm LocalBindingMap) (checked.ProductPatternInfo, *source.Error) {
		var record, ok = unboxRecord(in, mod)
		if !(ok) {
			return nil, source.MakeError(pattern.Location,
				E_CannotMatchRecord {
					TypeName: typsys.DescribeType(in, nil),
				})
		}
		var occurred = make(map[string] struct{})
		var info = make(checked.ProductPatternInfo, len(pattern.FieldMaps))
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
				var binding = &checked.LocalBinding {
					Name:     binding_name,
					Type:     t,
					Location: binding_loc,
				}
				info[i] = checked.ProductPatternItemInfo {
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
		for _, item := range info {
			lm.add(item.Binding)
		}
		return info, nil
	})
}

func checkTuple(T ast.Tuple) ExprChecker {
	return makeExprChecker(T.Location, func(cc *checkContext) checkResult {
		var L = len(T.Elements)
		if L == 0 {
			return cc.assign(typsys.UnitType {}, checked.UnitValue {})
		}
		if L == 1 {
			return cc.forwardToChildExpr(T.Elements[0])
		}
		if L > MaxTupleSize {
			return cc.error(
				E_TooManyTupleElements { SizeLimitError {
					Given: uint(L),
					Limit: MaxTupleSize,
				}})
		}
		if cc.expected == nil {
			var elements = make([] *checked.Expr, L)
			var types = make([] typsys.Type, L)
			for i := 0; i < L; i += 1 {
				var el, err = cc.checkChildExpr(nil, T.Elements[i])
				if err != nil { return cc.propagate(err) }
				elements[i] = el
				types[i] = el.Type
			}
			var tuple_t = &typsys.NestedType {
				Content: typsys.Tuple { Elements: types },
			}
			return cc.assign(tuple_t, checked.Tuple { Elements: elements })
		} else {
			var tuple, is_tuple = getTuple(cc.expected)
			if !(is_tuple) {
				return cc.error(
					E_TupleAssignedToIncompatible {
						TypeName: cc.describeType(cc.expected),
					})
			}
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
				var el, err = cc.checkChildExpr(tuple.Elements[i], T.Elements[i])
				if err != nil { return cc.propagate(err) }
				elements[i] = el
			}
			var tuple_t = &typsys.NestedType { Content: tuple }
			return cc.assign(tuple_t, checked.Tuple { Elements: elements })
		}
	})
}

func checkRecord(R ast.Record) ExprChecker {
	return makeExprChecker(R.Location, func(cc *checkContext) checkResult {
		var num_fields = uint(len(R.Values))
		if num_fields > MaxRecordSize {
			return cc.error(
				E_TooManyRecordFields { SizeLimitError {
					Given: num_fields,
					Limit: MaxRecordSize,
				}})
		}
		var update, has_update = R.Update.(ast.Update)
		var mapping = make(map[string] uint)
		var fields = make([] typsys.Field, num_fields)
		var values = make([] *checked.Expr, num_fields)
		for i, item := range R.Values {
			var key = item.Key
			var k = ast.Id2String(key)
			var _, duplicate = mapping[k]
			mapping[k] = uint(i)
			if duplicate {
				return cc.error(
					E_DuplicateField {
						FieldName: k,
					})
			}
			var value = getFieldValue(item)
			var value_expr, err = cc.checkChildExpr(nil, value)
			if err != nil { return cc.propagate(err) }
			values[i] = value_expr
			fields[i] = typsys.Field {
				Attr: attr.FieldAttrs {
					Attrs: attr.Attrs { Location: item.Location },
				},
				Name: k,
				Type: value_expr.Type,
			}
		}
		if has_update {
			var base, err = cc.checkChildExpr(nil, update.Base)
			if err != nil { return cc.propagate(err) }
			var base_record, ok = cc.unboxRecord(base)
			if !(ok) {
				return cc.error(
					E_UpdateOnNonRecord {
						TypeName: cc.describeTypeOf(base),
					})
			}
			var replaced = make([] checked.TupleUpdateElement, num_fields)
			for i := uint(0); i < num_fields; i += 1 {
				var field = fields[i]
				var k = field.Name
				var base_index, exists = base_record.FieldIndexMap[k]
				if !(exists) {
					return cc.error(
						E_FieldNotFound {
							FieldName: k,
							TypeName:  cc.describeTypeOf(base),
						})
				}
				var base_field = base_record.Fields[base_index]
				var err = cc.assignType(base_field.Type, field.Type)
				if err != nil { return cc.propagate(err) }
				replaced[i] = checked.TupleUpdateElement {
					Index: base_index,
					Value: values[i],
				}
			}
			var record_t = &typsys.NestedType { Content: base_record }
			var record_update = checked.TupleUpdate {
				Base:     base,
				Replaced: replaced,
			}
			return cc.assign(record_t, record_update)
		} else {
			var record_t = &typsys.NestedType {
				Content: typsys.Record {
					FieldIndexMap: mapping,
					Fields:        fields,
				},
			}
			var record = checked.Tuple {
				Elements: values,
			}
			return cc.assign(record_t, record)
		}
	})
}

func getFieldValue(item ast.FieldValue) ast.Expr {
	var given_value, given = item.Value.(ast.Expr)
	if given {
		return given_value
	} else {
		return desugarOmittedFieldValue(item.Key)
	}
}
func desugarOmittedFieldValue(key ast.Identifier) ast.Expr {
	return ast.Expr {
		Node: key.Node,
		Term: ast.VariousTerm {
			Node: key.Node,
			Term: ast.InlineRef {
				Node: key.Node,
				Item: key,
			},
		},
	}
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
func unboxWeak(t typsys.Type, mod string) typsys.Type {
	var inner, box, exists = typsys.Unbox(t, mod)
	if exists && box.WeakWrapping {
		return unboxWeak(inner, mod)
	} else {
		return t
	}
}
func unboxTuple(t typsys.Type, mod string) (typsys.Tuple, bool) {
	var tuple, is_tuple = getTuple(t)
	if is_tuple {
		return tuple, true
	} else {
		var inner, _, exists = typsys.Unbox(t, mod)
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
		var inner, _, exists = typsys.Unbox(t, mod)
		if exists {
			return unboxRecord(inner, mod)
		} else {
			return typsys.Record {}, false
		}
	}
}


