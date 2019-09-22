package object

import "unsafe"
import ."kumachan/interpreter/assertion"

/**
 *	The process of validating generic types:
 *     1. go through all type/function declarations, for each declaration:
 *        (1) for each bound of it
 *            e.g. `B[C]` and `T` in `class A[T < B[C], U < T] { .. }`
 *            check if the bound is valid
 *            using `GenericType::IsArgBoundsValid()`
 *        (*) "valid" means any type parameter cannot be used as
 *            a type argument in the bound.
 *            e.g. `B[C]` is a valid bound of T, but `B[T]` is NOT
 *                 `T` is a valid bound of U, but `B[T]` is NOT
 *     2. go through all type/function declarations, for each declaration:
 *        (1) if it is a schema declaration
 *            i. for each declared (generic) base schema,
 *                get all base schemas of it.
 *            ii. do (i) recursively to get all super schemas
 *            iii. for each super schema, check
 *                - if it is a type expr of form `G[T1, ...]` or `G`
 *                - if it is a GenericSchemaType
 *                - if it is extensible
 *                - if is has correct mutability
 *                    * mutable can only inherit mutable, and vice versa
 *                - if it is not duplicate or circular
 *            iv. collect all fields of all super schemas
 *                - check if there is a conflict field
 *        (2) if it is a class declaration
 *            i. for each declared (generic) base class,
 *                get all base classes of it.
 *            ii. do (i) recursively to get all super classes
 *            iii. for each super class, check
 *                - if it is a type expr of form `G[T1, ...]` or `G`
 *                - if it is a GenericClassType
 *                - if it is extensible
 *                - if it is not duplicate or circular
 *            iv. collect all methods of all super schemas
 *                - check if there is a conflict method
 *            v. for each declared base interface,
 *                - check if the interface is implemented
 *                    * this check is performed by doing structural comparison
 *                      between class and interface
 *                      on each type expr of each method
 *                      using `TypeExprEqual()`
 *        (3) if it is a union type
 *            i. for each inner type expr in the declaration
 *               check if all used generic type is already validated
 *               using `TypeExpr::AreAllTemplatesValidated()`
 *        (4) if it is a trait type
 *            i. for each constraint of the trait,
 *               check if it is an extensible class or an interface
 *        (5) create placeholder types for its type parameters
 *            using `GenericType::GetArgPlaceholders()`
 *        (6) for each inner type expression of it
 *            validate the type expression using `TypeExpr::Check()`
 *        (7) mark it validated
 */

type ValidationError struct {
	__Kind      ValidationErrorKind
	__Template  *GenericType
}

type ValidationErrorKind int
const (
	VEK_SuperTypeCircular ValidationErrorKind = iota
	VEK_SuperTypeDuplicate
	VEK_SuperTypeInvalid
	VEK_SuperTypeNotExtensible
	VEK_SuperTypeWrongMutability
	VEK_FieldConflict
	VEK_MethodConflict
	VEK_InflationError
	VEK_InterfaceDuplicate
	VEK_InterfaceNotImplemented
	VEK_BadUnion
	VEK_InvalidTraitConstraint
)

type VE_SuperTypeCircular struct {
	__ValidationError  ValidationError
	__CircularType     *GenericType
}

type VE_SuperTypeDuplicate struct {
	__ValidationError  ValidationError
	__DuplicateType    *GenericType
}

type VE_SuperTypeInvalid struct {
	__ValidationError  ValidationError
	__InvalidSuper     *TypeExpr
}

type VE_SuperTypeNotExtensible struct {
	__ValidationError    ValidationError
	__NonExtensibleType  *GenericType
}

type VE_SuperTypeWrongMutability struct {
	__ValidationError  ValidationError
	__WrongSchema      *GenericType
	__NeedImmutable    bool
}

type VE_FieldConflict struct {
	__ValidationError  ValidationError
	__Schema1          *GenericType
	__Schema2          *GenericType
	__FieldName        Identifier
}

type VE_MethodConflict struct {
	__ValidationError  ValidationError
	__Class1           *GenericType
	__Class2           *GenericType
	__MethodName       Identifier
}

type VE_InflationError struct {
	__ValidationError  ValidationError
	__InflationError   *InflationError
}

type VE_InterfaceNotImplemented struct {
	__ValidationError  ValidationError
	__Interface        *GenericType
	__BadMethod        Identifier
	__IsMissing        bool
}

type VE_InterfaceDuplicate struct {
	__ValidationError  ValidationError
	__Interface        *GenericType
}

type VE_BadUnion struct {
	__ValidationError  ValidationError
	__BadExpr          *TypeExpr
}

type VE_InvalidTraitConstraint struct {
	__ValidationError  ValidationError
	__Constraint       *TypeExpr
}

type InflationError struct {
	__Kind      InflationErrorKind
	__Expr      *TypeExpr
	__Template  *GenericType
}

type InflationErrorKind int
const (
	IEK_TemplateNotValidated InflationErrorKind = iota
	IEK_WrongArgQuantity
	IEK_InvalidArg
)

type IE_WrongArgQuantity struct {
	__InflationError  InflationError
	__Given           int
	__Required        int
}

type IE_InvalidArg struct {
	__InflationError  InflationError
	__Arg             int
	__Bound           int
	__IsUpper         bool
}


func (G *GenericType) Validate(ctx *ObjectContext) *ValidationError {
	var super_invalid = func (e *TypeExpr) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_SuperTypeInvalid {
				__ValidationError: ValidationError {
					__Kind:     VEK_SuperTypeInvalid,
					__Template: G,
				},
				__InvalidSuper: e,
			}),
		)
	}
	var super_duplicate = func (super int) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_SuperTypeDuplicate{
				__ValidationError: ValidationError {
					__Kind:     VEK_SuperTypeDuplicate,
					__Template: G,
				},
				__DuplicateType: ctx.GetGenericType(super),
			}),
		)
	}
	var super_circular = func (super int) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_SuperTypeCircular{
				__ValidationError: ValidationError {
					__Kind:     VEK_SuperTypeCircular,
					__Template: G,
				},
				__CircularType: ctx.GetGenericType(super),
			}),
		)
	}
	var not_extensible = func (S *GenericType) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_SuperTypeNotExtensible {
				__ValidationError: ValidationError {
					__Kind:     VEK_SuperTypeNotExtensible,
					__Template: G,
				},
				__NonExtensibleType: S,
			}),
		)
	}
	var wrong_mutability = func (S *GenericType, im bool) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_SuperTypeWrongMutability {
				__ValidationError: ValidationError {
					__Kind:     VEK_SuperTypeWrongMutability,
					__Template: G,
				},
				__WrongSchema: S,
				__NeedImmutable: im,
			}),
		)
	}
	var field_conflict = func (
		S1 *GenericType,
		S2 *GenericType,
		name Identifier,
	) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_FieldConflict {
				__ValidationError: ValidationError {
					__Kind: VEK_FieldConflict,
					__Template: G,
				},
				__Schema1: S1,
				__Schema2: S2,
				__FieldName: name,
			}),
		)
	}
	var method_conflict = func (
		C1 *GenericType,
		C2 *GenericType,
		name Identifier,
	) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_MethodConflict {
				__ValidationError: ValidationError {
					__Kind: VEK_MethodConflict,
					__Template: G,
				},
				__Class1: C1,
				__Class2: C2,
				__MethodName: name,
			}),
		)
	}
	var interface_duplicate = func (I *GenericType) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_InterfaceDuplicate {
				__ValidationError: ValidationError {
					__Kind: VEK_InterfaceDuplicate,
					__Template: G,
				},
				__Interface: I,
			}),
		)
	}
	var not_implemented = func (
		I *GenericType,
		method Identifier,
		missing bool,
	) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_InterfaceNotImplemented {
				__ValidationError: ValidationError {
					__Kind: VEK_InterfaceNotImplemented,
					__Template: G,
				},
				__Interface: I,
				__BadMethod: method,
				__IsMissing: missing,
			}),
		)
	}
	var bad_union = func (e *TypeExpr) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_BadUnion {
				__ValidationError: ValidationError {
					__Kind: VEK_BadUnion,
					__Template: G,
				},
				__BadExpr: e,
			}),
		)
	}
	var invalid_constraint = func (e *TypeExpr) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_InvalidTraitConstraint{
				__ValidationError: ValidationError {
					__Kind:     VEK_InvalidTraitConstraint,
					__Template: G,
				},
				__Constraint: e,
			}),
		)
	}
	var inflation_error = func (err *InflationError) *ValidationError {
		return (*ValidationError)(
			unsafe.Pointer(&VE_InflationError {
				__ValidationError: ValidationError {
					__Kind: VEK_InflationError,
					__Template: G,
				},
				__InflationError: err,
			}),
		)
	}
	// Check for Hierarchy Consistency
	if G.__Kind == GT_Schema {
		var G_as_Schema = (*GenericSchemaType)(unsafe.Pointer(G))
		var supers = map[int]int { G.__Id: 0 }
		var add_base_to_supers func(*TypeExpr, int) *ValidationError
		add_base_to_supers = func (base *TypeExpr, depth int) *ValidationError {
			if base.__Kind != TE_Inflation {
				return super_invalid(base)
			}
			var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
			var B = base_as_inf_expr.__Template
			if B.__Kind != GT_Schema {
				return super_invalid(base)
			}
			var B_as_Schema = (*GenericSchemaType)(unsafe.Pointer(B))
			if !(B_as_Schema.__Extensible) {
				return not_extensible(B)
			}
			var required_mutability = G_as_Schema.__Immutable
			var given_mutability = B_as_Schema.__Immutable
			if required_mutability != given_mutability {
				return wrong_mutability(B, required_mutability)
			}
			var current = B.__Id
			var existing_depth, exists = supers[current]
			if exists {
				if existing_depth == depth {
					return super_duplicate(current)
				} else {
					return super_circular(current)
				}
			}
			supers[current] = depth
			for _, base_of_base := range B_as_Schema.__BaseList {
				var err = add_base_to_supers(base_of_base, depth+1)
				if err != nil {
					return err
				}
			}
			return nil
		}
		for _, base := range G_as_Schema.__BaseList {
			var err = add_base_to_supers(base, 1)
			if err != nil {
				return err
			}
		}
		var occurred_fields = make(map[Identifier] *GenericType)
		for super, _ := range supers {
			var S = ctx.GetGenericType(super)
			Assert(S.__Kind == GT_Schema, "Generics: invalid generic schema")
			var S_as_Schema = (*GenericSchemaType)(unsafe.Pointer(S))
			for _, field := range S_as_Schema.__OwnFieldList {
				var name = field.__Name
				var existing, exists = occurred_fields[name]
				if exists {
					return field_conflict(existing, S, name)
				} else {
					occurred_fields[name] = S
				}
			}
		}
	} else if G.__Kind == GT_Class {
		var G_as_Class = (*GenericClassType)(unsafe.Pointer(G))
		var supers = map[int]int { G.__Id:  0 }
		var add_base_to_supers func(*TypeExpr, int) *ValidationError
		add_base_to_supers = func (base *TypeExpr, depth int) *ValidationError {
			if base.__Kind != TE_Inflation {
				return super_invalid(base)
			}
			var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
			var B = base_as_inf_expr.__Template
			if B.__Kind != GT_Class {
				return super_invalid(base)
			}
			var B_as_Class = (*GenericClassType)(unsafe.Pointer(B))
			if !(B_as_Class.__Extensible) {
				return not_extensible(B)
			}
			var current = B.__Id
			var existing_depth, exists = supers[current]
			if exists {
				if existing_depth == depth {
					return super_duplicate(current)
				} else {
					return super_circular(current)
				}
			}
			supers[current] = depth
			for _, base_of_base := range B_as_Class.__BaseClassList {
				var err = add_base_to_supers(base_of_base, depth+1)
				if err != nil {
					return err
				}
			}
			return nil
		}
		for _, base := range G_as_Class.__BaseClassList {
			var err = add_base_to_supers(base, 1)
			if err != nil {
				return err
			}
		}
		var occurred_methods = make(map[Identifier] *GenericType)
		var own_method_types = make(map[Identifier] *TypeExpr)
		for super, _ := range supers {
			var S = ctx.GetGenericType(super)
			Assert(S.__Kind == GT_Class, "Generics: invalid generic class")
			var S_as_Class = (*GenericClassType)(unsafe.Pointer(S))
			for _, method := range S_as_Class.__OwnMethodList {
				var name = method.__Name
				var existing, exists = occurred_methods[name]
				if exists {
					return method_conflict(existing, S, name)
				} else {
					occurred_methods[name] = S
					if super == G.__Id {
						own_method_types[name] = method.__Type
					}
				}
			}
		}
		// Check if all interfaces are implemented by TypeExprEqual()
		var I_supers = make(map[int]*GenericType)
		for _, base := range G_as_Class.__BaseInterfaceList {
			if base.__Kind != TE_Inflation {
				return super_invalid(base)
			}
			var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
			var B = base_as_inf_expr.__Template
			if B.__Kind != GT_Interface {
				return super_invalid(base)
			}
			var _, exists = I_supers[B.__Id]
			if exists {
				return interface_duplicate(B)
			}
			I_supers[B.__Id] = B
			var B_as_Interface = (*GenericInterfaceType)(unsafe.Pointer(B))
			Assert (
				len(B_as_Interface.__MethodList) > 0,
				"Generics: empty interface is not valid",
			)
			for _, method := range B_as_Interface.__MethodList {
				var name = method.__Name
				var given_T, exists = own_method_types[name]
				if !exists {
					return not_implemented(B, name, true)
				} else {
					var required_T = method.__Type
					var ok = TypeExprEqual (
						given_T, required_T,
						base_as_inf_expr.__Arguments,
					)
					if !ok {
						return not_implemented(B, name, false)
					}
				}
			}
		}
	}
	// Check unions and traits
	if G.__Kind == GT_Union {
		var G_as_Union = (*GenericUnionType)(unsafe.Pointer(G))
		for _, element := range G_as_Union.__Elements {
			var ok, bad_expr = element.AreAllTemplatesValidated()
			if !ok {
				return bad_union(bad_expr)
			}
		}
	} else if G.__Kind == GT_Trait {
		var G_as_Trait = (*GenericTraitType)(unsafe.Pointer(G))
		for _, constraint := range G_as_Trait.__Constraints {
			if constraint.__Kind != TE_Inflation {
				return invalid_constraint(constraint)
			}
			var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(constraint))
			var template = inf_expr.__Template
			var k = template.__Kind
			if k != GT_Interface && k != GT_Class {
				return invalid_constraint(constraint)
			} else if k == GT_Class {
				var class = (*GenericClassType)(unsafe.Pointer(template))
				if !(class.__Extensible) {
					return invalid_constraint(constraint)
				}
			}
		}
	}
	// Evaluate bounds and generate placeholders by GetArgPlaceholders()
	// note: bounds of all types must be pre-validated by IsArgBoundsValid()
	var args, err = G.GetArgPlaceholders(ctx)
	if err != nil {
		return inflation_error(err)
	}
	// Validate each inner type expression in the template
	switch G.__Kind {
	case GT_Function:
		var G_as_Function = (*GenericFunctionType)(unsafe.Pointer(G))
		var err = G_as_Function.__Signature.Check(ctx, args)
		if err != nil {
			return inflation_error(err)
		}
	case GT_Union:
		var G_as_Union = (*GenericUnionType)(unsafe.Pointer(G))
		for _, element := range G_as_Union.__Elements {
			var err = element.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
	case GT_Trait:
		var G_as_Trait = (*GenericTraitType)(unsafe.Pointer(G))
		for _, constraint := range G_as_Trait.__Constraints {
			var err = constraint.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
	case GT_Schema:
		var G_as_Schema = (*GenericSchemaType)(unsafe.Pointer(G))
		for _, base := range G_as_Schema.__BaseList {
			var err = base.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
		for _, field := range G_as_Schema.__OwnFieldList {
			var err = field.__Type.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
	case GT_Class:
		var G_as_Class = (*GenericClassType)(unsafe.Pointer(G))
		for _, base_class := range G_as_Class.__BaseClassList {
			var err = base_class.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
		for _, base_interface := range G_as_Class.__BaseInterfaceList {
			var err = base_interface.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
		for _, method := range G_as_Class.__OwnMethodList {
			var err = method.__Type.Check(ctx, args)
			if err != nil {
				return inflation_error(err)
			}
		}
	}
	// Mark validated
	G.__Validated = true
	return nil
}

func (e *TypeExpr) IsValidBound(depth int) (bool, *TypeExpr) {
	switch e.__Kind {
	case TE_Final:
		return true, nil
	case TE_Argument:
		// should forbid [T < G[T]] and [T, U < G[T]]
		if depth == 0 {
			return true, nil
		} else {
			return false, e
		}
	case TE_Function:
		var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
		var ok = true
		var te *TypeExpr
		for _, f := range f_expr.__Items {
			for _, p := range f.__Parameters {
				ok, te = p.IsValidBound(depth+1)
				if !ok { break }
			}
			ok, te = f.__ReturnValue.IsValidBound(depth+1)
			if !ok { break }
			ok, te = f.__Exception.IsValidBound(depth+1)
			if !ok { break }
		}
		return ok, te
	case TE_Inflation:
		var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
		var ok = true
		var te *TypeExpr
		for _, arg_expr := range inf_expr.__Arguments {
			ok, te = arg_expr.IsValidBound(depth+1)
			if !ok { break }
		}
		return ok, te
	default:
		panic("impossible branch")
	}
}

func (G *GenericType) IsArgBoundsValid() (bool, *TypeExpr) {
	var ok = true
	var te *TypeExpr
	for _, parameter := range G.__Parameters {
		if parameter.__UpperBound != nil {
			ok, te = parameter.__UpperBound.IsValidBound(0)
			if !ok { break }
		}
		if parameter.__LowerBound != nil {
			ok, te = parameter.__LowerBound.IsValidBound(0)
			if !ok { break }
		}
	}
	return ok, te
}

func TypeExprEqual (e1 *TypeExpr, e2 *TypeExpr, args []*TypeExpr) bool {
	if e2.__Kind == TE_Argument {
		var e2_as_arg = (*ArgumentTypeExpr)(unsafe.Pointer(e2))
		e2 = args[e2_as_arg.__Index]
	}
	if e1.__Kind != e2.__Kind {
		return false
	} else {
		var kind = e1.__Kind
		switch kind {
		case TE_Final:
			var final1 = (*FinalTypeExpr)(unsafe.Pointer(e1))
			var final2 = (*FinalTypeExpr)(unsafe.Pointer(e2))
			return final1.__Type == final2.__Type
		case TE_Argument:
			var a1 = (*ArgumentTypeExpr)(unsafe.Pointer(e1))
			var a2 = (*ArgumentTypeExpr)(unsafe.Pointer(e2))
			return a1.__Index == a2.__Index
		case TE_Function:
			var f1 = (*FunctionTypeExpr)(unsafe.Pointer(e1))
			var f2 = (*FunctionTypeExpr)(unsafe.Pointer(e2))
			if len(f1.__Items) == len(f2.__Items) {
				var LI = len(f1.__Items)
				for i := 0; i < LI; i += 1 {
					var I1 = f1.__Items[i]
					var I2 = f2.__Items[i]
					if len(I1.__Parameters) == len(I2.__Parameters) {
						var LP = len(I1.__Parameters)
						for j := 0; j < LP; j += 1 {
							var P1 = I1.__Parameters[i]
							var P2 = I2.__Parameters[i]
							if !(TypeExprEqual(P1, P2, args)) {
								return false
							}
						}
						var R1 = I1.__ReturnValue
						var R2 = I2.__ReturnValue
						if !(TypeExprEqual(R1, R2, args)) {
							return false
						}
						var E1 = I1.__Exception
						var E2 = I2.__Exception
						if !(TypeExprEqual(E1, E2, args)) {
							return false
						}
					} else {
						return false
					}
				}
				return true
			} else {
				return false
			}
		case TE_Inflation:
			var inf1 = (*InflationTypeExpr)(unsafe.Pointer(e1))
			var inf2 = (*InflationTypeExpr)(unsafe.Pointer(e2))
			if inf1.__Template.__Id == inf2.__Template.__Id {
				if len(inf1.__Arguments) == len(inf2.__Arguments) {
					var L = len(inf1.__Arguments)
					for i := 0; i < L; i += 1 {
						var a1 = inf1.__Arguments[i]
						var a2 = inf2.__Arguments[i]
						if !(TypeExprEqual(a1, a2, args)) {
							return false
						}
					}
					return true
				} else {
					return false
				}
			} else {
				return false
			}
		default:
			panic("impossible branch")
		}
	}
}

func (e *TypeExpr) AreAllTemplatesValidated() (bool, *TypeExpr) {
	switch e.__Kind {
	case TE_Final:
		return true, nil
	case TE_Argument:
		return true, nil
	case TE_Function:
		var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
		var ok = true
		var te *TypeExpr
		for _, f := range f_expr.__Items {
			for _, p := range f.__Parameters {
				ok, te = p.AreAllTemplatesValidated()
				if !ok { break }
			}
			ok, te = f.__ReturnValue.AreAllTemplatesValidated()
			if !ok { break }
			ok, te = f.__Exception.AreAllTemplatesValidated()
			if !ok { break }
		}
		return ok, te
	case TE_Inflation:
		var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
		var ok = true
		var te *TypeExpr
		for _, arg_expr := range inf_expr.__Arguments {
			ok, te = arg_expr.AreAllTemplatesValidated()
			if !ok { break }
		}
		return ok, te
	default:
		panic("impossible branch")
	}
}

func (G *GenericType) GetArgPlaceholders(ctx *ObjectContext) (
	[]int, *InflationError,
) {
	var args = make([]int, len(G.__Parameters))
	var T_args = make([]*T_Placeholder, len(args))
	for i, parameter := range G.__Parameters {
		var T_arg = &T_Placeholder {
			__TypeInfo: TypeInfo {
				__Kind: TK_Placeholder,
				__Name: parameter.__Name,
				__Initialized: false,
			},
			__UpperBound: -1,
			__LowerBound: -1,
		}
		ctx.__RegisterType((*TypeInfo)(unsafe.Pointer(T_arg)))
		args[i] = T_arg.__TypeInfo.__Id
		T_args[i] = T_arg
	}
	for i, _ := range T_args {
		var parameter = G.__Parameters[i]
		if parameter.__UpperBound != nil {
			var B, err = parameter.__UpperBound.Try2Evaluate(ctx, args)
			if err != nil {
				return nil, err
			}
			if B != T_args[i].__TypeInfo.__Id {
				T_args[i].__UpperBound = B
			}
		}
		if parameter.__LowerBound != nil {
			var B, err = parameter.__LowerBound.Try2Evaluate(ctx, args)
			if err != nil {
				return nil, err
			}
			if B != T_args[i].__TypeInfo.__Id {
				T_args[i].__LowerBound = B
			}
		}
		T_args[i].__TypeInfo.__Initialized = true
	}
	return args, nil
}

func (e *TypeExpr) Check(ctx *ObjectContext, args []int) *InflationError {
	switch e.__Kind {
	case TE_Final:
		return nil
	case TE_Argument:
		return nil
	case TE_Function:
		var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
		for _, f := range f_expr.__Items {
			for _, ei := range f.__Parameters {
				var err = ei.Check(ctx, args)
				if err != nil {
					return err
				}
			}
			var err1 = f.__ReturnValue.Check(ctx, args)
			if err1 != nil {
				return err1
			}
			var err2 = f.__Exception.Check(ctx, args)
			if err2 != nil {
				return err2
			}
		}
		return nil
	case TE_Inflation:
		var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
		var to_be_checked = inf_expr.__Arguments
		var template = inf_expr.__Template
		if !(template.__Validated) {
			return &InflationError {
				__Kind: IEK_TemplateNotValidated,
				__Expr: e,
				__Template: template,
			}
		} else {
			return CheckArgs(ctx, args, template, e, to_be_checked)
		}
	default:
		panic("impossible branch")
	}
}

func CheckArgs (
	ctx    *ObjectContext,
	args   [] int,
	G      *GenericType,
	expr   *TypeExpr,
	chk    [] *TypeExpr,
) *InflationError {
	if len(chk) != len(G.__Parameters) {
		return (*InflationError)(unsafe.Pointer(&IE_WrongArgQuantity {
			__InflationError: InflationError {
				__Kind: IEK_WrongArgQuantity,
				__Template: G,
				__Expr: expr,
			},
			__Given: len(chk),
			__Required: len(G.__Parameters),
		}))
	}
	var need_eval = make([]bool, len(chk))
	var has_bound = make([]bool, len(chk))
	var used_by_other = make([]bool, len(chk))
	for i, parameter := range G.__Parameters {
		var L = parameter.__LowerBound
		var U = parameter.__UpperBound
		if L != nil {
			if L.__Kind == TE_Argument {
				var L_as_Arg = (*ArgumentTypeExpr)(unsafe.Pointer(L))
				used_by_other[L_as_Arg.__Index] = true
			}
			has_bound[i] = true
		}
		if U != nil {
			if U.__Kind == TE_Argument {
				var U_as_Arg = (*ArgumentTypeExpr)(unsafe.Pointer(U))
				used_by_other[U_as_Arg.__Index] = true
			}
			has_bound[i] = true
		}
	}
	for i := 0; i < len(chk); i += 1 {
		need_eval[i] = has_bound[i] || used_by_other[i]
	}
	var chk_evaluated = make([]int, len(chk))
	for i, e := range chk {
		if need_eval[i] {
			var id, err = e.Try2Evaluate(ctx, args)
			if err != nil {
				return err
			}
			chk_evaluated[i] = id
		} else {
			chk_evaluated[i] = -1
		}
	}
	var check_bound = func(bound int, checked int, upper bool) *InflationError {
		var B = ctx.GetType(bound)
		var A = ctx.GetType(checked)
		var ok bool
		if upper {
			ok = (A.IsSubTypeOf(B, ctx) == True)
		} else {
			ok = (B.IsSubTypeOf(A, ctx) == True)
		}
		if ok {
			return nil
		} else {
			return (*InflationError)(unsafe.Pointer(&IE_InvalidArg {
				__InflationError: InflationError {
					__Kind: IEK_InvalidArg,
					__Template: G,
					__Expr: expr,
				},
				__Bound: bound,
				__Arg: checked,
				__IsUpper: upper,
			}))
		}
	}
	for i, parameter := range G.__Parameters {
		var L = parameter.__LowerBound
		var U = parameter.__UpperBound
		if L != nil {
			var bound, err = L.Try2Evaluate(ctx, chk_evaluated)
			if err != nil {
				return err
			}
			err = check_bound(bound, chk_evaluated[i], false)
			if err != nil {
				return err
			}
		}
		if U != nil {
			var bound, err = U.Try2Evaluate(ctx, chk_evaluated)
			if err != nil {
				return err
			}
			err = check_bound(bound, chk_evaluated[i], true)
			if err != nil {
				return err
			}
		}
		if !need_eval[i] {
			var chk_i = chk[i]
			if chk_i.__Kind == TE_Inflation {
				var ie = (*InflationTypeExpr)(unsafe.Pointer(chk_i))
				var G_inner = ie.__Template
				var chk_inner = ie.__Arguments
				var err = CheckArgs(ctx, args, G_inner, chk_i, chk_inner)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}