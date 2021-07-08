package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


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

func productPatternMatch(pattern ast.VariousPattern, in typsys.Type, mod string, lm localBindingMap) (checked.ProductPatternInfo, *source.Error) {
	switch P := pattern.Pattern.(type) {
	case ast.PatternTrivial:
		return patternMatchTrivial(P, in, mod, lm)
	case ast.PatternTuple:
		return patternMatchTuple(P, in, mod, lm)
	case ast.PatternRecord:
		return patternMatchRecord(P, in, mod, lm)
	default:
		panic("impossible branch")
	}
}

// TODO: extract common part of function signatures
func patternMatchTrivial(pattern ast.PatternTrivial, in typsys.Type, mod string, lm localBindingMap) (checked.ProductPatternInfo, *source.Error) {
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
}

func patternMatchTuple(pattern ast.PatternTuple, in typsys.Type, mod string, lm localBindingMap) (checked.ProductPatternInfo, *source.Error) {
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
}

func patternMatchRecord(pattern ast.PatternRecord, in typsys.Type, mod string, lm localBindingMap) (checked.ProductPatternInfo, *source.Error) {
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


func checkTuple(T ast.Tuple) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(T.Location, &s, ctx, nil)
		if expected == nil {
			var L = len(T.Elements)
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
				var el, err = cc.checkChildExpr(tuple.Elements[i], T.Elements[i])
				if err != nil { return cc.propagate(err) }
				elements[i] = el
			}
			var tuple_t = &typsys.NestedType { Content: tuple }
			return cc.ok(tuple_t, checked.Tuple { Elements: elements })
		}
	})
}

func checkRecord(R ast.Record) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(R.Location, &s, ctx, nil)
		var num_fields = uint(len(R.Values))
		if num_fields > MaxRecordSize {
			return nil, nil, source.MakeError(R.Location,
				E_TooManyRecordFields { SizeLimitError {
					Given: num_fields,
					Limit: MaxRecordSize,
				}})
		}
		if expected == nil {
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
					return cc.error(E_DuplicateField { FieldName: k })
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
				var base_record, ok = unboxRecord(base.Type, ctx.ModName)
				if !(ok) {
					return cc.error(E_UpdateOnNonRecord {
						TypeName: typsys.DescribeType(base.Type, nil),
					})
				}
				var replaced = make([] checked.TupleUpdateElement, num_fields)
				for i := uint(0); i < num_fields; i += 1 {
					var field = fields[i]
					var k = field.Name
					var base_index, exists = base_record.FieldIndexMap[k]
					if !(exists) {
						return cc.error(E_FieldNotFound {
							FieldName: k,
							TypeName:  typsys.DescribeType(base.Type, nil),
						})
					}
					var base_field = base_record.Fields[base_index]
					if !(cc.assignType(base_field.Type, field.Type)) {
						return cc.error(E_NotAssignable{
							From: typsys.DescribeType(base_field.Type, nil),
							To:   typsys.DescribeType(field.Type, nil),
						})
					}
					replaced[i] = checked.TupleUpdateElement {
						Index: base_index,
						Value: values[i],
					}
				}
				var record_t = &typsys.NestedType { Content: base_record }
				var record = checked.TupleUpdate {
					Base:     base,
					Replaced: replaced,
				}
				return cc.ok(record_t, record)
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
				return cc.ok(record_t, record)
			}
		} else {
			var _, has_update = R.Update.(ast.Update)
			if has_update {
				var R_term = ast.VariousTerm {
					Node: R.Node,
					Term: R,
				}
				var expr, err = cc.checkChildTerm(nil, R_term)
				if err != nil { return cc.propagate(err) }
				return cc.assign(expected, expr.Type, expr.Content)
			} else {
				var expected_record, ok = getRecord(expected)
				if !(ok) {
					return cc.error(E_RecordAssignedToIncompatible {
						TypeName: typsys.DescribeType(expected, s),
					})
				}
				var required_num_fields = uint(len(expected_record.Fields))
				if num_fields != required_num_fields {
					return cc.error(E_RecordSizeNotMatching {
						Given:    num_fields,
						Required: required_num_fields,
					})
				}
				var occurred = make(map[string] struct{})
				var values = make([] *checked.Expr, num_fields)
				for i, item := range R.Values {
					var k = ast.Id2String(R.Values[i].Key)
					var _, duplicate = occurred[k]
					occurred[k] = struct{}{}
					if duplicate {
						return cc.error(E_DuplicateField { FieldName: k })
					}
					var e_index, exists = expected_record.FieldIndexMap[k]
					if !(exists) {
						return cc.error(E_FieldNotFound {
							FieldName: k,
							TypeName:  typsys.DescribeType(expected, s),
						})
					}
					var value = getFieldValue(item)
					var value_expr, err = cc.checkChildExpr(nil, value)
					if err != nil { return cc.propagate(err) }
					values[e_index] = value_expr
				}
				var record_t = &typsys.NestedType {
					Content: expected_record,
				}
				var record = checked.Tuple {
					Elements: values,
				}
				return cc.ok(record_t, record)
			}
		}
	})
}


