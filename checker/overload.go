package checker

import (
	"fmt"
	"strings"
	. "kumachan/error"
)


type Candidate struct {
	Function  *GenericFunction
	Name      string
	Error     *ExprError
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
		var expr, err = GenericFunctionCall (
			f, name, 0, type_args, arg, f_info, call_info, ctx,
		)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(expr), nil
	} else {
		var options = make([]AvailableCall, 0)
		var candidates = make([]Candidate, 0)
		for i, f := range functions {
			var index = uint(i)
			var expr, err = GenericFunctionCall (
				f, name, index, type_args, arg, f_info, call_info, ctx,
			)
			if err != nil {
				candidates = append(candidates, Candidate {
					Function: f,
					Name:     name,
					Error:    err,
				})
			} else {
				options = append(options, AvailableCall {
					Expr:     expr,
					Function: f,
				})
			}
		}
		var available_count = len(options)
		if available_count == 0 {
			return SemiExpr{}, &ExprError {
				Point:    call_info.ErrorPoint,
				Concrete: E_NoneOfFunctionsCallable {
					Candidates: candidates,
				},
			}
		} else if available_count == 1 {
			return LiftTyped(options[0].Expr), nil
		} else {
			return SemiExpr {
				Value: UndecidedCall {
					Options:  options,
					FuncName: name,
				},
				Info: call_info,
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
		var candidates = make([]Candidate, 0)
		for i, f := range functions {
			var index = uint(i)
			var expr, err = GenericFunctionAssignTo (
				expected, name, index, f, type_args, info, ctx,
			)
			if err != nil {
				candidates = append(candidates, Candidate {
					Function: f,
					Name:     name,
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
					To:         ctx.DescribeType(expected),
					Candidates: candidates,
				},
			}
		}
	}
}

func DescribeCandidate(c Candidate) string {
	var name = c.Name
	var f = c.Function
	var params = TypeParamsNames(f.TypeParams)
	return fmt.Sprintf (
		"%s[%s]: %s",
		name,
		strings.Join(params, ","),
		DescribeTypeWithParams (
			AnonymousType { f.DeclaredType },
			params,
		),
	)
}

func CheckOverload (
	functions     [] FunctionReference,
	added_type    Func,
	added_name    string,
	added_params  [] string,
	reg           TypeRegistry,
	err_point     ErrorPoint,
) *FunctionError {
	for _, existing := range functions {
		var existing_t = AnonymousType { existing.Function.DeclaredType }
		var added_t = AnonymousType { added_type }
		var cannot_overload = AreTypesConflict(existing_t, added_t, reg)
		if cannot_overload {
			var existing_params = existing.Function.TypeParams
			return &FunctionError {
				Point: err_point,
				Concrete: E_InvalidOverload {
					BetweenLocal: !(existing.IsImported),
					AddedName:    added_name,
					AddedModule:  existing.ModuleName,
					AddedType: DescribeTypeWithParams (
						added_t, added_params,
					),
					ExistingType: DescribeTypeWithParams (
						existing_t, TypeParamsNames(existing_params),
					),
				},
			}
		}
	}
	return nil
}

func AreTypesConflict(type1 Type, type2 Type, reg TypeRegistry) bool {
	// Does type1 conflict with type2 (when overloading functions)
	switch t1 := type1.(type) {
	case WildcardRhsType:
		return true
	case ParameterType:
		return true  // rough comparison
	case NamedType:
		switch t2 := type2.(type) {
		case NamedType:
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
			var check_union = func(union Union, another NamedType) bool {
				var q = [] Union { union }
				for len(q) > 0 {
					var u = q[0]
					q = q[1:]
					for _, sub := range u.CaseTypes {
						if another.Name == sub.Name {
							return check_args()
						} else {
							var t = reg[sub.Name]
							var sub_union, sub_is_union = t.Value.(Union)
							if sub_is_union {
								q = append(q, sub_union)
							}
						}
					}
				}
				return false
			}
			if t1.Name == t2.Name {
				return check_args()
			} else {
				var T1 = reg[t1.Name]
				var T2 = reg[t2.Name]
				var t1_union, t1_is_union = T1.Value.(Union)
				if t1_is_union {
					if check_union(t1_union, t2) {
						return true
					}
				}
				var t2_union, t2_is_union = T2.Value.(Union)
				if t2_is_union {
					if check_union(t2_union, t1) {
						return true
					}
				}
				return false
			}
		default:
			return false
		}
	case AnonymousType:
		switch t2 := type2.(type) {
		case AnonymousType:
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

func IsLocalType(type_ Type, mod string) bool {
	switch t := type_.(type) {
	case WildcardRhsType:
		return false
	case ParameterType:
		return false
	case NamedType:
		if t.Name.ModuleName == mod {
			return true
		} else {
			for _, arg := range t.Args {
				if IsLocalType(arg, mod) {
					return true
				}
			}
			return false
		}
	case AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return false
		case Tuple:
			for _, el := range r.Elements {
				if IsLocalType(el, mod) {
					return true
				}
			}
			return false
		case Bundle:
			for _, f := range r.Fields {
				if IsLocalType(f.Type, mod) {
					return true
				}
			}
			return false
		case Func:
			if IsLocalType(r.Input, mod) {
				return true
			}
			if IsLocalType(r.Output, mod) {
				return true
			}
			return false
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

