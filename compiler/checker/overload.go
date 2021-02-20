package checker

import (
	"fmt"
	"strings"
	. "kumachan/util/error"
)


type AvailableCall struct {
	Expr       Expr
	Function   *GenericFunction
	Inferring  TypeArgsInferringContext
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
		return GenerateCallResult (
			name, call_info, available, unavailable,
			false, TypeArgsInferringContext {}, ctx,
		)
	}
}

func AssignUndecidedTo(expected Type, call UndecidedCall, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var name = call.FuncName
	var available = make([] AvailableCall, 0)
	var unavailable = make([] UnavailableCall, 0)
	for _, opt := range call.Calls {
		var this_call_ctx = ctx.WithInferringStateCloned()
		var expr, err = TypedAssignTo(expected, opt.Expr, this_call_ctx)
		if err == nil {
			available = append(available, AvailableCall {
				Expr:      expr,
				Function:  opt.Function,
				Inferring: this_call_ctx.Inferring,
			})
		} else {
			unavailable = append(unavailable, UnavailableCall {
				Function: opt.Function,
				Name:     name,
				Error:    err,
			})
		}
	}
	semi, err := GenerateCallResult (
		name, info, available, unavailable,
		true, ctx.Inferring, ctx,
	)
	if err != nil { return Expr{}, err }
	return Expr(semi.Value.(TypedExpr)), nil
}

func GenerateCallResult (
	name         string,
	info         ExprInfo,
	available    [] AvailableCall,
	unavailable  [] UnavailableCall,
	assigned     bool,
	inferring    TypeArgsInferringContext,
	ctx          ExprContext,
) (SemiExpr, *ExprError) {
	var mod_name = ctx.GetModuleName()
	if len(available) == 0 {
		var unavailable_info = make([] UnavailableFuncInfo, len(unavailable))
		for i, item := range unavailable {
			unavailable_info[i].FuncDesc = DescribeFunction(item.Function, name, mod_name)
			unavailable_info[i].Error = item.Error
		}
		return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_NoneOfFunctionsCallable { unavailable_info },
		}
	} else if len(available) == 1 {
		var opt = available[0]
		if assigned {
			inferring.MergeArgsFrom(opt.Inferring)
		}
		return LiftTyped(opt.Expr), nil
	} else {
		if assigned {
			var min_type Type = &AnyType {}
			var min_index = ^uint(0)
			var min_not_found = false
			for i, item := range available {
				var item_t = item.Expr.Type
				if TypeEqual(min_type, item_t, ctx.GetTypeRegistry()) {
					min_not_found = true
					break
				}
				var _, ok = AssignType(min_type, item_t, ToInferred, ctx)
				if ok {
					min_type = item_t
					min_index = uint(i)
				} else {
					var _, ok = AssignType(item_t, min_type, ToInferred, ctx)
					if !(ok) {
						min_not_found = true
						break
					}
				}
			}
			if !(min_not_found) {
				var opt = available[min_index]
				inferring.MergeArgsFrom(opt.Inferring)
				return LiftTyped(opt.Expr), nil
			}
			var available_desc = make([] string, len(available))
			for i, item := range available {
				available_desc[i] = DescribeFunction(item.Function, name, mod_name)
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
	var mod_name = ctx.GetModuleName()
	if len(functions) == 0 { panic("something went wrong") }
	if len(functions) == 1 {
		var f = functions[0]
		return GenericFunctionAssignTo (
			expected, name, 0, f, type_args, info, ctx,
		)
	} else {
		var candidates = make([] UnavailableFuncInfo, 0)
		var ok_expr Expr
		var ok_inferring TypeArgsInferringContext
		var available = make([] *GenericFunction, 0)
		for i, f := range functions {
			var this_f_ctx = ctx.WithInferringStateCloned()
			var index = uint(i)
			var expr, err = GenericFunctionAssignTo (
				expected, name, index, f, type_args, info, this_f_ctx,
			)
			if err != nil {
				candidates = append(candidates, UnavailableFuncInfo {
					FuncDesc: DescribeFunction(f, name, mod_name),
					Error:    err,
				})
			} else {
				available = append(available, f)
				ok_expr = expr
				ok_inferring = this_f_ctx.Inferring
			}
		}
		if len(available) == 0 {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_NoneOfFunctionsAssignable {
					To:         ctx.DescribeInferredType(expected),
					Candidates: candidates,
				},
			}
		} else if len(available) == 1 {
			ctx.Inferring.MergeArgsFrom(ok_inferring)
			return ok_expr, nil
		} else {
			var ok_candidates = make([] string, len(available))
			for i, f := range available {
				ok_candidates[i] = DescribeFunction(f, name, mod_name)
			}
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_AmbiguousFunctionAssign {
					Candidates: ok_candidates,
				},
			}
		}
	}
}

func DescribeFunction(f *GenericFunction, name string, mod string) string {
	var params = TypeParamsNames(f.TypeParams)
	return fmt.Sprintf (
		"%s[%s]: %s",
		name,
		strings.Join(params, ","),
		DescribeTypeWithParams (
			&AnonymousType { f.DeclaredType },
			params,
			mod,
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
	for _, existing := range functions {
		var existing_p = existing.Function.TypeParams
		var added_p = added_params
		var existing_t = &AnonymousType { existing.Function.DeclaredType }
		var added_t = &AnonymousType { added_type }
		var existing_f = existing.Function.Implicit
		var added_f = added_fields
		if len(existing_p) != len(added_p) {
			continue
		}
		var L = uint(len(existing_p))
		var args = make([]Type, L)
		for i := uint(0); i < L; i += 1 {
			args[i] = &ParameterType { Index: i }
		}
		var t1 = FillTypeArgs(existing_t, args)
		var t2 = FillTypeArgs(added_t, args)
		if TypeEqual(t1, t2, reg) {
			var f1 = FillTypeArgs(&AnonymousType { Bundle { existing_f } }, args)
			var f2 = FillTypeArgs(&AnonymousType { Bundle { added_f } }, args)
			if TypeEqual(f1, f2, reg) {
				return &FunctionError {
					Point: err_point,
					Concrete: E_InvalidOverload {
						BetweenLocal: !(existing.IsImported),
						AddedName:    added_name,
						AddedModule:  existing.ModuleName,
					},
				}
			}
		}
	}
	return nil
}

