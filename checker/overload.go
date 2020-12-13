package checker

import (
	"fmt"
	"strings"
	. "kumachan/error"
)


type AvailableCall struct {
	Expr      Expr
	Function  *GenericFunction
}
type UnavailableCall struct {
	Function  *GenericFunction
	Name      string
	Error     *ExprError
}
func (impl UndecidedCall) SemiExprVal() {}
type UndecidedCall struct {
	FuncName  string
	Calls     [] AvailableCall
}

func OverloadedCall (
	functions  [] *GenericFunction,
	name       string,
	type_args  [] Type,
	arg        SemiExpr,
	f_info     ExprInfo,
	call_info  ExprInfo,
	ctx        ExprContext,
) (SemiExpr, *ExprError) {
	if len(functions) == 0 { panic("something went wrong") }
	if len(functions) == 1 {
		var f = functions[0]
		call, err := GenericFunctionCall (
			f, name, 0, type_args,
			arg, f_info, call_info, ctx,
		)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(call), nil
	} else {
		var available = make([] AvailableCall, 0)
		var unavailable = make([] UnavailableCall, 0)
		for i, f := range functions {
			var index = uint(i)
			var expr, err = GenericFunctionCall (
				f, name, index, type_args,
				arg, f_info, call_info, ctx,
			)
			if err != nil {
				unavailable = append(unavailable, UnavailableCall {
					Function: f,
					Name:     name,
					Error:    err,
				})
			} else {
				available = append(available, AvailableCall {
					Expr:     expr,
					Function: f,
				})
			}
		}
		return GenerateCallResult(name, call_info, available, unavailable, false)
	}
}

func AssignUndecidedTo(expected Type, call UndecidedCall, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	err := RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	var name = call.FuncName
	var available = make([] AvailableCall, 0)
	var unavailable = make([] UnavailableCall, 0)
	for _, opt := range call.Calls {
		var expr, err = TypedAssignTo(expected, opt.Expr, ctx)
		if err == nil {
			available = append(available, AvailableCall {
				Expr:     expr,
				Function: opt.Function,
			})
		} else {
			unavailable = append(unavailable, UnavailableCall{
				Function: opt.Function,
				Name:     name,
				Error:    err,
			})
		}
	}
	semi, err := GenerateCallResult(name, info, available, unavailable, true)
	if err != nil { return Expr{}, err }
	return Expr(semi.Value.(TypedExpr)), nil
}

func GenerateCallResult (
	name         string,
	info         ExprInfo,
	available    [] AvailableCall,
	unavailable  [] UnavailableCall,
	assigned     bool,
) (SemiExpr, *ExprError) {
	if len(available) == 0 {
		var unavailable_info = make([] UnavailableFuncInfo, len(unavailable))
		for i, item := range unavailable {
			unavailable_info[i].FuncDesc = DescribeFunction(item.Function, name)
			unavailable_info[i].Error = item.Error
		}
		return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_NoneOfFunctionsCallable { unavailable_info },
		}
	} else if len(available) == 1 {
		return LiftTyped(available[0].Expr), nil
	} else {
		if assigned {
			var available_desc = make([] string, len(available))
			for i, item := range available {
				available_desc[i] = DescribeFunction(item.Function, name)
			}
			return SemiExpr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: &E_AmbiguousCall { available_desc },
			}
		} else {
			return SemiExpr {
				Value: UndecidedCall {
					FuncName: name,
					Calls:    available,
				},
				Info:  info,
			}, nil
		}
	}
}

func OverloadedAssignTo (
	expected   Type,
	functions  [] *GenericFunction,
	name       string,
	type_args  [] Type,
	info       ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	if len(functions) == 0 { panic("something went wrong") }
	if len(functions) == 1 {
		var f = functions[0]
		return GenericFunctionAssignTo (
			expected, name, 0, f, type_args, info, ctx,
		)
	} else {
		var candidates = make([] UnavailableFuncInfo, 0)
		for i, f := range functions {
			var index = uint(i)
			var expr, err = GenericFunctionAssignTo (
				expected, name, index, f, type_args, info, ctx,
			)
			if err != nil {
				candidates = append(candidates, UnavailableFuncInfo {
					FuncDesc: DescribeFunction(f, name),
					Error:    err,
				})
			} else {
				return expr, nil
			}
		}
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		} else {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_NoneOfFunctionsAssignable {
					To:         ctx.DescribeExpectedType(expected),
					Candidates: candidates,
				},
			}
		}
	}
}

func DescribeFunction(f *GenericFunction, name string) string {
	var params = TypeParamsNames(f.TypeParams)
	return fmt.Sprintf (
		"%s[%s]: %s",
		name,
		strings.Join(params, ","),
		DescribeTypeWithParams (
			&AnonymousType { f.DeclaredType },
			params,
		),
	)
}

func ValidateOverload (
	functions     [] FunctionReference,
	added_type    Func,
	added_fields  map[string] Field,
	added_name    string,
	added_params  [] string,
	reg           TypeRegistry,
	err_point     ErrorPoint,
) *FunctionError {
	// TODO: fix bug here
	for _, existing := range functions {
		var existing_t = &AnonymousType { existing.Function.DeclaredType }
		var added_t = &AnonymousType { added_type }
		var existing_fields = existing.Function.Implicit
		var cannot_overload = false
		if AreTypesConflict(existing_t, added_t, reg) {
			if len(existing_fields) == len(added_fields) {
				var same_keys = true
				for key, _ := range added_fields {
					var _, exists = existing_fields[key]
					if !exists { same_keys = false }
				}
				if same_keys {
					for key, added_field := range added_fields {
						var existing_field = existing_fields[key]
						if AreTypesConflict(added_field.Type, existing_field.Type, reg) {
							cannot_overload = true
							break
						}
					}
				}
			} else {
				var less map[string] Field
				var more map[string] Field
				if len(existing_fields) < len(added_fields) {
					less = existing_fields
					more = added_fields
				} else if len(existing_fields) > len(added_fields) {
					less = added_fields
					more = existing_fields
				} else {
					panic("impossible branch")
				}
				var conflict = true
				for key, less_field := range less {
					var more_field, exists = more[key]
					if exists {
						if !(AreTypesConflict(less_field.Type, more_field.Type, reg)) {
							conflict = false
							break
						}
					} else {
						conflict = false
						break
					}
				}
				if conflict {
					cannot_overload = true
				}
			}
		}
		if cannot_overload {
			var existing_params = TypeParamsNames(existing.Function.TypeParams)
			var sig_desc = func (
				params  ([] string),
				t       Type,
				fields  map[string] Field,
			) string {
				var t_desc = DescribeTypeWithParams(t, params)
				var fields_desc = ""
				if len(fields) > 0 {
					var fields_t = &AnonymousType { Bundle { fields } }
					var desc = DescribeTypeWithParams(fields_t, params)
					fields_desc = fmt.Sprintf(" (implicit: %s)", desc)
				}
				return (t_desc + fields_desc)
			}
			var added_desc = sig_desc(added_params, added_t, added_fields)
			var existing_desc = sig_desc(existing_params, existing_t, existing_fields)
			return &FunctionError {
				Point: err_point,
				Concrete: E_InvalidOverload {
					BetweenLocal: !(existing.IsImported),
					AddedName:    added_name,
					AddedModule:  existing.ModuleName,
					AddedSig:     added_desc,
					ExistingSig:  existing_desc,
				},
			}
		}
	}
	return nil
}

func AreTypesConflict(type1 Type, type2 Type, reg TypeRegistry) bool {
	// Does type1 conflict with type2 (when overloading functions)
	switch t1 := type1.(type) {
	case *WildcardRhsType:
		return true
	case *ParameterType:
		return true  // rough comparison
	case *NamedType:
		switch t2 := type2.(type) {
		case *NamedType:
			var check_args = func() bool {
				var L1 = len(t1.Args)
				var L2 = len(t2.Args)
				if L1 != L2 { panic("type registration went wrong") }
				var L = L1
				for i := 0; i < L; i += 1 {
					var a1, a2 = t1.Args[i], t2.Args[i]
					if !(AreTypesConflict(a1, a2, reg)) {
						return false
					}
				}
				return true
			}
			if t1.Name == t2.Name {
				return check_args()
			} else {
				return false
			}
		default:
			return false
		}
	case *AnonymousType:
		switch t2 := type2.(type) {
		case *AnonymousType:
			switch r1 := t1.Repr.(type) {
			case Unit:
				switch t2.Repr.(type) {
				case Unit:
					return true
				default:
					return false
				}
			case Tuple:
				switch r2 := t2.Repr.(type) {
				case Tuple:
					var L1 = len(r1.Elements)
					var L2 = len(r2.Elements)
					if L1 == L2 {
						var L = L1
						for i := 0; i < L; i += 1 {
							var e1, e2 = r1.Elements[i], r2.Elements[i]
							if !(AreTypesConflict(e1, e2, reg)) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Bundle:
				switch r2 := t2.Repr.(type) {
				case Bundle:
					var L1 = len(r1.Fields)
					var L2 = len(r2.Fields)
					if L1 == L2 {
						for name, f1 := range r1.Fields {
							var f2, exists = r2.Fields[name]
							if !exists { return false }
							var t1, t2 = f1.Type, f2.Type
							if !(AreTypesConflict(t1, t2, reg)) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Func:
				switch r2 := t2.Repr.(type) {
				case Func:
					if !(AreTypesConflict(r1.Input, r2.Input, reg)) {
						return false
					}
					if !(AreTypesConflict(r1.Output, r2.Output, reg)) {
						return false
					}
					return true
				default:
					return true
				}
			default:
				panic("impossible branch")
			}
		default:
			return false
		}
	default:
		panic("impossible branch")
	}
}

