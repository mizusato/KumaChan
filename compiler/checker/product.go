package checker

import (
	. "kumachan/misc/util/error"
	"kumachan/lang/parser/ast"
)


func (impl SemiTypedTuple) SemiExprVal() {}
type SemiTypedTuple struct {
	Values  [] SemiExpr
}

func (impl SemiTypedRecord) SemiExprVal() {}
type SemiTypedRecord struct {
	Index     map[string] uint
	Values    [] SemiExpr
	KeyNodes  [] ast.Node
}

func (impl Product) ExprVal() {}
type Product struct {
	Values  [] Expr
}

func (impl Get) ExprVal() {}
type Get struct {
	Product  Expr
	Index    uint
}

func (impl Reference) ExprVal() {}
type Reference struct {
	Base     Expr
	Index    uint
	Kind     ReferenceKind
	Operand  ReferenceOperand
}
type ReferenceKind int
const (
	RK_Field ReferenceKind = iota
	RK_Branch
)
type ReferenceOperand int
const (
	RO_Record ReferenceOperand = iota
	RO_Enum
	RO_ProjRef
	RO_CaseRef
)

func (impl Set) ExprVal() {}
type Set struct {
	Product   Expr
	Index     uint
	NewValue  Expr
}


func CheckTuple(tuple ast.Tuple, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(tuple.Node)
	var L = len(tuple.Elements)
	if L == 0 {
		return LiftTyped(Expr {
			Type:  &AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		}), nil
	} else if L == 1 {
		var expr, err = Check(tuple.Elements[0], ctx)
		if err != nil { return SemiExpr{}, err }
		return expr, nil
	} else {
		var el_exprs = make([] SemiExpr, L)
		for i, el := range tuple.Elements {
			var expr, err = Check(el, ctx)
			if err != nil { return SemiExpr{}, err }
			el_exprs[i] = expr
		}
		return SemiExpr {
			Value: SemiTypedTuple { el_exprs },
			Info: info,
		}, nil
	}
}

func CheckRecord(record ast.Record, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(record.Node)
	switch update := record.Update.(type) {
	case ast.Update:
		var base_semi, err = Check(update.Base, ctx)
		if err != nil { return SemiExpr{}, err }
		switch b := base_semi.Value.(type) {
		case TypedExpr:
			if IsRecordLiteral(Expr(b)) { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToLiteralRecord {},
			} }
			var L = len(record.Values)
			if !(L >= 1) { panic("something went wrong") }
			var base = Expr(b)
			switch target_ := UnboxRecord(base.Type, ctx).(type) {
			case BR_Record:
				var target = target_.Record
				var occurred_names = make(map[string] bool)
				var current_base = base
				for _, field := range record.Values {
					var name = ast.Id2String(field.Key)
					var target_field, exists = target.Fields[name]
					if !exists {
						return SemiExpr{}, &ExprError {
							Point: ErrorPointFrom(field.Key.Node),
							Concrete: E_FieldDoesNotExist {
								Field:  name,
								Target: ctx.DescribeCertainType(base.Type),
							},
						}
					}
					var _, duplicate = occurred_names[name]
					if duplicate {
						return SemiExpr{}, &ExprError {
							Point:    ErrorPointFrom(field.Key.Node),
							Concrete: E_ExprDuplicateField { name },
						}
					}
					occurred_names[name] = true
					var value_node = DesugarOmittedFieldValue(field)
					var value_semi, err1 = Check(value_node, ctx)
					if err1 != nil { return SemiExpr{}, err1 }
					var value, err2 = AssignTo(target_field.Type, value_semi, ctx)
					if err2 != nil { return SemiExpr{}, err2 }
					current_base = Expr {
						Type:  current_base.Type,
						Value: Set {
							Product:  current_base,
							Index:    target_field.Index,
							NewValue: value,
						},
						Info:  current_base.Info,
					}
				}
				var final = current_base
				return SemiExpr {
					Value: TypedExpr(final),
					Info:  info,
				}, nil
			case BR_RecordButOpaque:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToOpaqueRecord {},
				}
			case BR_NonRecord:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToNonRecord {},
				}
			default:
				panic("impossible branch")
			}
		case SemiTypedRecord:
			return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToLiteralRecord {},
			}
		default:
			return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToNonRecord {},
			}
		}
	default:
		var L = len(record.Values)
		var f_exprs = make([] SemiExpr, L)
		var f_index_map = make(map[string] uint, L)
		var f_key_nodes = make([] ast.Node, L)
		for i, field := range record.Values {
			var name = ast.Id2String(field.Key)
			var _, exists = f_index_map[name]
			if exists { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(field.Key.Node),
				Concrete: E_ExprDuplicateField { name },
			} }
			var value = DesugarOmittedFieldValue(field)
			var expr, err = Check(value, ctx)
			if err != nil { return SemiExpr{}, err }
			f_exprs[i] = expr
			f_index_map[name] = uint(i)
			f_key_nodes[i] = field.Key.Node
		}
		return SemiExpr {
			Value: SemiTypedRecord {
				Index:    f_index_map,
				Values:   f_exprs,
				KeyNodes: f_key_nodes,
			},
			Info: info,
		}, nil
	}
}

func CheckGet(base SemiExpr, key ast.Identifier, info ExprInfo, ctx ExprContext) (SemiExpr, *ExprError) {
	switch b := base.Value.(type) {
	case UntypedRef:
		// TODO: find record in output types of overloaded functions
		var expr, err = AssignTo(nil, base, ctx)
		if err == nil {
			return CheckGet(LiftTyped(expr), key, info, ctx)
		} else {
			return SemiExpr{}, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		}
	case TypedExpr:
		if IsRecordLiteral(Expr(b)) { return SemiExpr{}, &ExprError {
			Point:    base.Info.ErrorPoint,
			Concrete: E_GetFromLiteralRecord {},
		} }
		switch record_ := UnboxRecord(b.Type, ctx).(type) {
		case BR_Record:
			var record = record_.Record
			var key_string = ast.Id2String(key)
			var field, exists = record.Fields[key_string]
			if !exists { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(key.Node),
				Concrete: E_FieldDoesNotExist {
					Field:  key_string,
					Target: ctx.DescribeCertainType(&AnonymousType { record }),
				},
			} }
			var t = field.Type
			return LiftTyped(Expr {
				Type:  t,
				Value: Get {
					Product: Expr(b),
					Index:   field.Index,
				},
				Info:  info,
			}), nil
		case BR_RecordButOpaque:
			return SemiExpr{}, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromOpaqueRecord {},
			}
		case BR_NonRecord:
			return SemiExpr{}, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromNonRecord {},
			}
		default:
			panic("impossible branch")
		}
	case SemiTypedRecord:
		return SemiExpr{}, &ExprError {
			Point:    base.Info.ErrorPoint,
			Concrete: E_GetFromLiteralRecord {},
		}
	default:
		return SemiExpr{}, &ExprError {
			Point:    base.Info.ErrorPoint,
			Concrete: E_GetFromNonRecord {},
		}
	}
}

func CheckRefField(base SemiExpr, key ast.Identifier, info ExprInfo, ctx ExprContext) (SemiExpr, *ExprError) {
	var get_field_info = func(t Type) (Type, uint, *ExprError) {
		switch record_ := UnboxRecord(t, ctx).(type) {
		case BR_Record:
			var record = record_.Record
			var key_string = ast.Id2String(key)
			var field, exists = record.Fields[key_string]
			if !exists {
				return nil, BadIndex, &ExprError {
					Point:    ErrorPointFrom(key.Node),
					Concrete: E_FieldDoesNotExist {
						Field:  key_string,
						Target: ctx.DescribeCertainType(&AnonymousType{record}),
					},
				}
			}
			return field.Type, field.Index, nil
		case BR_RecordButOpaque:
			return nil, BadIndex, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromOpaqueRecord {},
			}
		case BR_NonRecord:
			return nil, BadIndex, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromNonRecord {},
			}
		default:
			panic("impossible branch")
		}
	}
	{
		var inf_ctx = ctx.WithInferringEnabled(__ProjRefParams, __NoBounds)
		var base_assigned, err = AssignTo(__ProjRefToBeInferred, base, inf_ctx)
		if err == nil {
			var args = inf_ctx.Inferring.GetPlainArgs()
			var ref_base_t = args[0]
			var ref_field_t = args[1]
			var next_field_type, next_field_index, err = get_field_info(ref_field_t)
			if err != nil { return SemiExpr{}, err }
			return LiftTyped(Expr {
				Type: ProjRef(ref_base_t, next_field_type),
				Value: Reference {
					Base:    base_assigned,
					Index:   next_field_index,
					Kind:    RK_Field,
					Operand: RO_ProjRef,
				},
				Info: info,
			}), nil
		}
	}
	{
		var inf_ctx = ctx.WithInferringEnabled(__CaseRefParams, __NoBounds)
		var base_assigned, err = AssignTo(__CaseRefToBeInferred, base, inf_ctx)
		if err == nil {
			var args = inf_ctx.Inferring.GetPlainArgs()
			var ref_base_t = args[0]
			var ref_case_t = args[1]
			var field_type, field_index, err = get_field_info(ref_case_t)
			if err != nil { return SemiExpr{}, err }
			return LiftTyped(Expr {
				Type:  CaseRef(ref_base_t, field_type),
				Value: Reference {
					Base:    base_assigned,
					Index:   field_index,
					Kind:    RK_Field,
					Operand: RO_CaseRef,
				},
				Info: info,
			}), nil
		}
	}
	{
		var base_typed, err = AssignTo(nil, base, ctx)
		if err != nil { return SemiExpr{}, err }
		var base_type = base_typed.Type
		switch record_ := UnboxRecord(base_type, ctx).(type) {
		case BR_Record:
			var record = record_.Record
			var key_string = ast.Id2String(key)
			var field, exists = record.Fields[key_string]
			if !exists { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(key.Node),
				Concrete: E_FieldDoesNotExist {
					Field:  key_string,
					Target: ctx.DescribeCertainType(&AnonymousType { record }),
				},
			} }
			return LiftTyped(Expr {
				Type:  ProjRef(base_type, field.Type),
				Value: Reference {
					Base:    Expr(base_typed),
					Index:   field.Index,
					Kind:    RK_Field,
					Operand: RO_Record,
				},
				Info:  info,
			}), nil
		case BR_RecordButOpaque:
			return SemiExpr{}, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromOpaqueRecord {},
			}
		case BR_NonRecord:
			return SemiExpr{}, &ExprError {
				Point:    base.Info.ErrorPoint,
				Concrete: E_GetFromNonRecord {},
			}
		default:
			panic("impossible branch")
		}
	}
}


func AssignTupleTo(expected Type, tuple SemiTypedTuple, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var non_nil_expected Type
	if expected == nil {
		non_nil_expected = &AnonymousType {
			Tuple {
				// Fill with nil
				Elements: make([] Type, len(tuple.Values)),
			},
		}
	} else {
		non_nil_expected = expected
	}
	switch E := non_nil_expected.(type) {
	default:
		var typed_exprs = make([] Expr, len(tuple.Values))
		for i, el := range tuple.Values {
			var typed, err = AssignTo(nil, el, ctx)
			if err != nil { return Expr{}, err }
			typed_exprs[i] = typed
		}
		var el_types = make([]Type, len(tuple.Values))
		for i, el := range typed_exprs {
			el_types[i] = el.Type
		}
		var final_t = &AnonymousType { Tuple { el_types } }
		var typed_tuple = Expr {
			Type:  final_t,
			Value: Product { typed_exprs },
			Info:  info,
		}
		return TypedAssignTo(expected, typed_tuple, ctx)
	case *AnonymousType:
		switch tuple_t := E.Repr.(type) {
		case Tuple:
			var required = len(tuple_t.Elements)
			var given = len(tuple.Values)
			if given != required {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_TupleSizeNotMatching {
						Required:  required,
						Given:     given,
						GivenType: ctx.DescribeInferredType(&AnonymousType {tuple_t }),
					},
				}
			}
			var typed_exprs = make([] Expr, given)
			for i, el := range tuple.Values {
				var el_expected = tuple_t.Elements[i]
				var typed, err = AssignTo(el_expected, el, ctx)
				if err != nil { return Expr{}, err }
				typed_exprs[i] = typed
			}
			var el_types = make([] Type, len(tuple.Values))
			for i, el := range typed_exprs {
				el_types[i] = el.Type
			}
			var final_t = &AnonymousType { Tuple { el_types } }
			return Expr {
				Type:  final_t,
				Info:  info,
				Value: Product { typed_exprs },
			}, nil
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_TupleAssignedToNonTupleType {
			NonTupleType: ctx.DescribeInferredType(non_nil_expected),
		},
	}
}

func AssignRecordTo(expected Type, record SemiTypedRecord, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	switch E := expected.(type) {
	case *AnonymousType:
		switch record_t := E.Repr.(type) {
		case Unit:
			if len(record.Values) == 0 {
				return Expr {
					Type:  &AnonymousType { Unit {} },
					Value: UnitValue {},
					Info:  info,
				}, nil
			}
		case Record:
			var values = make([] Expr, len(record_t.Fields))
			for field_name, field := range record_t.Fields {
				var given_index, exists = record.Index[field_name]
				if exists {
					var given_value = record.Values[given_index]
					var value, err = AssignTo(field.Type, given_value, ctx)
					if err != nil { return Expr{}, err }
					values[field.Index] = value
				} else {
					var node = info.ErrorPoint.Node
					var getter = CraftAstRefTerm(DefaultValueGetter, node)
					var unit = CraftAstTupleTerm(node)
					var getter_call = CraftAstCallExpr(getter, unit, node)
					var zero, err = AssignAstExprTo(field.Type, getter_call, ctx)
					if err != nil { return Expr{}, &ExprError {
						Point:    info.ErrorPoint,
						Concrete: E_MissingField {
							Field: field_name,
							Type:  ctx.DescribeInferredType(field.Type),
						},
					} }
					values[field.Index] = zero
				}
			}
			for given_field_name, index := range record.Index {
				var _, exists = record_t.Fields[given_field_name]
				if !exists {
					var key_node = record.KeyNodes[index]
					return Expr{}, &ExprError {
						Point:    ErrorPointFrom(key_node),
						Concrete: E_SuperfluousField { given_field_name },
					}
				}
			}
			var final_fields = make(map[string]Field)
			for field_name, field := range record_t.Fields {
				final_fields[field_name] = Field {
					Type:  values[field.Index].Type,
					Index: field.Index,
				}
			}
			var final_t = &AnonymousType { Record{ final_fields } }
			return Expr {
				Type:  final_t,
				Info:  info,
				Value: Product { values },
			}, nil
		}
	}
	return  Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_RecordAssignedToNonRecordType {
			NonRecordType: ctx.DescribeInferredType(expected),
		},
	}
}


func IsRecordLiteral(expr Expr) bool {
	switch expr.Value.(type) {
	case Product:
		switch t := expr.Type.(type) {
		case *AnonymousType:
			switch t.Repr.(type) {
			case Record:
				return true
			}
		}
	}
	return false
}

// TODO: CraftSemiTypedTuple

func CraftAstTupleTerm(node ast.Node, elements... ast.Expr) ast.VariousTerm {
	return ast.VariousTerm {
		Node: node,
		Term: ast.Tuple {
			Node:     node,
			Elements: elements,
		},
	}
}

func DesugarOmittedFieldValue(field ast.FieldValue) ast.Expr {
	switch val_expr := field.Value.(type) {
	case ast.Expr:
		return val_expr
	default:
		return ast.Expr {
			Node: field.Node,
			Term: ast.VariousTerm {
				Node: field.Node,
				Term: ast.InlineRef {
					Node:     field.Node,
					Module:   ast.Identifier {
						Node: field.Node,
						Name: []rune(""),
					},
					Id:       field.Key,
					TypeArgs: make([]ast.VariousType, 0),
				},
			},
			Pipeline: nil,
		}
	}
}
