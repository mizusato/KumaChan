package checker

import (
	"kumachan/interpreter/parser/ast"
	. "kumachan/standalone/util/error"
)


type MaybePattern interface { CheckerMaybePattern() }
func (impl Pattern) CheckerMaybePattern() {}
type Pattern struct {
	Point     ErrorPoint
	Concrete  ConcretePattern
}
type ConcretePattern interface { CheckerPattern() }
func (impl NullPattern) CheckerPattern() {}
type NullPattern struct {}
func (impl TrivialPattern) CheckerPattern() {}
type TrivialPattern struct {
	ValueName  string
	ValueType  Type
	Point      ErrorPoint
}
func (impl TuplePattern) CheckerPattern() {}
type TuplePattern struct {
	Items  [] PatternItem
}
func (impl RecordPattern) CheckerPattern() {}
type RecordPattern struct {
	Items  [] PatternItem
}

type PatternItem struct {
	Name   string
	Index  uint
	Type   Type
	Point  ErrorPoint
}


func PatternFrom (
	p_node  ast.VariousPattern,
	input   Type,
	ctx     ExprContext,
) (Pattern, *ExprError) {
	var err_result = func(e ConcreteExprError) (Pattern, *ExprError) {
		return Pattern{}, &ExprError {
			Point:    ErrorPointFrom(p_node.Node),
			Concrete: e,
		}
	}
	switch p := p_node.Pattern.(type) {
	case ast.PatternTrivial:
		return Pattern {
			Point:    ErrorPointFrom(p_node.Node),
			Concrete: TrivialPattern {
				ValueName: ast.Id2String(p.Name),
				ValueType: input,
				Point:     ErrorPointFrom(p.Name.Node),
			},
		}, nil
	case ast.PatternTuple:
		if len(p.Names) == 0 {
			return Pattern {
				Point:    ErrorPointFrom(p_node.Node),
				Concrete: TrivialPattern {
					ValueName: IgnoreMark,
					ValueType: input,
					Point:     ErrorPointFrom(p_node.Node),
				},
			}, nil
		}
		if len(p.Names) == 1 {
			// no single-element tuple
			return Pattern {
				Point:    ErrorPointFrom(p_node.Node),
				Concrete: TrivialPattern {
					ValueName: ast.Id2String(p.Names[0]),
					ValueType: input,
					Point:     ErrorPointFrom(p.Names[0].Node),
				},
			}, nil
		}
		switch tuple_ := UnboxTuple(input, ctx).(type) {
		case TR_Tuple:
			var tuple = tuple_.Tuple
			var required = len(p.Names)
			var given = len(tuple.Elements)
			if given != required {
				return err_result(E_TupleSizeNotMatching {
					Required:  required,
					Given:     given,
					GivenType: ctx.DescribeCertainType(&AnonymousType { tuple }),
				})
			} else {
				var occurred = make(map[string] bool)
				var ignored = 0
				var items = make([] PatternItem, 0)
				for i, identifier := range p.Names {
					var name = ast.Id2String(identifier)
					if name == IgnoreMark {
						ignored += 1
					} else {
						var _, duplicate = occurred[name]
						if duplicate {
							return Pattern{}, &ExprError {
								Point:    ErrorPointFrom(identifier.Node),
								Concrete: E_DuplicateBinding { name },
							}
						}
						occurred[name] = true
						var t = tuple.Elements[i]
						items = append(items, PatternItem {
							Name:  ast.Id2String(identifier),
							Index: uint(i),
							Type:  t,
							Point: ErrorPointFrom(identifier.Node),
						})
					}
				}
				if ignored == len(p.Names) {
					return err_result(E_EntireValueIgnored {})
				} else {
					return Pattern {
						Point:    ErrorPointFrom(p_node.Node),
						Concrete: TuplePattern { items },
					}, nil
				}
			}
		case TR_NonTuple:
			return err_result(E_MatchingNonTupleType {})
		case TR_TupleButOpaque:
			return err_result(E_MatchingOpaqueTupleType {})
		default:
			panic("impossible branch")
		}
	case ast.PatternRecord:
		switch record_ := UnboxRecord(input, ctx).(type) {
		case BR_Record:
			var record = record_.Record
			var occurred = make(map[string]  bool)
			var items = make([] PatternItem, len(p.FieldMaps))
			for i, field_map := range p.FieldMaps {
				var field_name = ast.Id2String(field_map.FieldName)
				var value_name = ast.Id2String(field_map.ValueName)
				var field, exists = record.Fields[field_name]
				if exists && field_name == IgnoreMark {
					// field should not be named using IgnoreMark;
					// IgnoreMark used in record pattern considered
					//   as "field does not exist" error
					panic("something went wrong")
				}
				if !exists {
					return Pattern{}, &ExprError {
						Point:    ErrorPointFrom(field_map.FieldName.Node),
						Concrete: E_FieldDoesNotExist {
							Field:  field_name,
							Target: ctx.DescribeCertainType(input),
						},
					}
				}
				var _, duplicate = occurred[value_name]
				if duplicate {
					return Pattern{}, &ExprError {
						Point:    ErrorPointFrom(field_map.ValueName.Node),
						Concrete: E_DuplicateBinding { value_name },
					}
				}
				occurred[value_name] = true
				var t = field.Type
				items[i] = PatternItem {
					Name:  value_name,
					Index: field.Index,
					Type:  t,
					Point: ErrorPointFrom(field_map.Node),
				}
			}
			return Pattern {
				Point:    ErrorPointFrom(p_node.Node),
				Concrete: RecordPattern { items },
			}, nil
		case BR_NonRecord:
			return err_result(E_MatchingNonRecordType {})
		case BR_RecordButOpaque:
			return err_result(E_MatchingOpaqueRecordType {})
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}


func (ctx ExprContext) WithPatternMatching(p Pattern, storage (map[string] Type)) ExprContext {
	var added = make(map[string] Type)
	switch P := p.Concrete.(type) {
	case TrivialPattern:
		var reg = ctx.ModuleInfo.Types
		added[P.ValueName] = UnboxWeak(P.ValueType, reg)
	case TuplePattern:
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
	case RecordPattern:
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
	default:
		panic("impossible branch")
	}
	if storage != nil {
		for k, v := range added {
			storage[k] = v
		}
	}
	var new_ctx = ctx.WithAddedLocalValues(added)
	return new_ctx
}

