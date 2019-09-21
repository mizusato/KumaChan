package object

import "fmt"
import "sort"
import "unsafe"
import "strings"
import ."kumachan/interpreter/assertion"

type TypeExpr struct {
    __Kind       TypeExprKind
    __Position   int
}

type TypeExprKind int
const (
    TE_Final TypeExprKind = iota
    TE_Argument
    TE_Function
    TE_Inflation
)

type FinalTypeExpr struct {
    __TypeExpr  TypeExpr
    __Type      int
}

type ArgumentTypeExpr struct {
    __TypeExpr   TypeExpr
    __Index      int
}

type FunctionTypeExpr struct {
    __TypeExpr   TypeExpr
    __Items      [] FunctionTypeExprItem
}

type FunctionTypeExprItem struct {
    __Parameters   [] *TypeExpr
    __ReturnValue  *TypeExpr
    __Exception    *TypeExpr
}

type InflationTypeExpr struct {
    __TypeExpr    TypeExpr
    __Arguments   [] *TypeExpr
    __Template    *GenericType
}

type GenericType struct {
    __Kind         GenericTypeKind
    __Id           int
    __Name         string
    __Parameters   [] GenericTypeParameter
    __Validated    bool
    __Position     int
}

type GenericTypeParameter struct {
    __Name            string
    __UpperBound      *TypeExpr
    __LowerBound      *TypeExpr
}

type GenericTypeKind int
const (
    GT_Function GenericTypeKind = iota
    GT_Union
    GT_Trait
    GT_Schema
    GT_Class
    GT_Interface
)

type GenericFunctionType struct {
    __GenericType  GenericType
    __Signature    *TypeExpr
}

type GenericUnionType struct {
    __GenericType  GenericType
    __Elements     [] *TypeExpr
    // TODO: a union should not reference itself
}

type GenericTraitType struct {
    __GenericType  GenericType
    __Constraints  [] *TypeExpr
    // TODO: element should be non-final class or interface
}

type GenericSchemaType struct {
    __GenericType     GenericType
    __Immutable       bool
    __Extensible      bool
    __BaseList        [] *TypeExpr
    __OwnFieldList    [] GenericSchemaField
}

type GenericSchemaField struct {
    __Name           Identifier
    __Type           *TypeExpr
    __HasDefault     bool
    __DefaultValue   *Object
}

type GenericClassType struct {
    __GenericType         GenericType
    __Extensible          bool
    __BaseClassList       [] *TypeExpr
    __BaseInterfaceList   [] *TypeExpr
    __OwnMethodList       [] GenericClassMethod
}

type GenericClassMethod struct {
    __Name      Identifier
    __Type      *TypeExpr
    __FunInfo   int
}

type GenericInterfaceType struct {
    __GenericType  GenericType
    __MethodList   [] GenericInterfaceMethod
}

type GenericInterfaceMethod struct {
    __Name   Identifier
    __Type   *TypeExpr
}

type ValidationError struct {
    __Kind      ValidationErrorKind
    __Template  *GenericType
}

type ValidationErrorKind int
const (
    VEK_CircularInheritance ValidationErrorKind = iota
    VEK_DuplicateSuperType
    VEK_InvalidSuperType
    VEK_DeriveNonExtensible
    VEK_DeriveWrongMutability
    VEK_FieldConflict
    VEK_MethodConflict
    VEK_InflationError
    VEK_InterfaceNotImplemented
)

type VE_CircularInheritance struct {
    __ValidationError  ValidationError
    __CircularType     *GenericType
}

type VE_DuplicateSuperType struct {
    __ValidationError  ValidationError
    __DuplicateType    *GenericType
}

type VE_InvalidSuperType struct {
    __ValidationError  ValidationError
    __InvalidSuper     *TypeExpr
}

type VE_DeriveNonExtensible struct {
    __ValidationError    ValidationError
    __NonExtensibleType  *GenericType
}

type VE_DeriveWrongMutability struct {
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

func (e *TypeExpr) Try2Evaluate(ctx *ObjectContext, args []int) (
    int, *InflationError,
) {
    var err = e.Check(ctx, args)
    if err != nil {
        return -1, err
    } else {
        return e.Evaluate(ctx, args), nil
    }
}

func (e *TypeExpr) Evaluate(ctx *ObjectContext, args []int) int {
    switch e.__Kind {
    case TE_Final:
        return (*FinalTypeExpr)(unsafe.Pointer(e)).__Type
    case TE_Argument:
        return args[(*ArgumentTypeExpr)(unsafe.Pointer(e)).__Index]
    case TE_Function:
        var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
        var items = make([]FunctionTypeItem, len(f_expr.__Items))
        for i, f := range f_expr.__Items {
            var params = make([]int, len(f.__Parameters))
            for j, ei := range f.__Parameters {
                params[j] = ei.Evaluate(ctx, args)
            }
            var retval = f.__ReturnValue.Evaluate(ctx, args)
            var exception = f.__Exception.Evaluate(ctx, args)
            items[i] = FunctionTypeItem {
                __Parameters: params,
                __ReturnValue: retval,
                __Exception: exception,
            }
        }
        return GetFunctionType(ctx, items)
    case TE_Inflation:
        var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
        var template = inf_expr.__Template
        Assert (
            template.__Validated,
            "Generics: unable to inflate non-validated generic type",
        )
        var args = make([]int, len(inf_expr.__Arguments))
        for i, arg_expr := range inf_expr.__Arguments {
            args[i] = arg_expr.Evaluate(ctx, args)
        }
        return template.Inflate(ctx, args)
    default:
        panic("impossible branch")
    }
}


func GetFunctionType(ctx *ObjectContext, items []FunctionTypeItem) int {
    var fingerprint = make([]int, 0)
    for _, item := range items {
        for _, parameter := range item.__Parameters {
            fingerprint = append(fingerprint, parameter)
        }
        fingerprint = append(fingerprint, item.__ReturnValue)
        fingerprint = append(fingerprint, item.__Exception)
        fingerprint = append(fingerprint, -1)
    }
    var cached, exists = ctx.GetInflatedType(-1, fingerprint)
    if exists {
        return cached
    } else {
        var item_names = make([]string, len(items))
        for i, item := range items {
            var param_names = make([]string, len(item.__Parameters))
            for j, parameter := range item.__Parameters {
                param_names[j] = ctx.GetTypeName(parameter)
            }
            var retval_name = ctx.GetTypeName(item.__ReturnValue)
            var exception_name = ctx.GetTypeName(item.__Exception)
            item_names[i] = fmt.Sprintf (
                "%v -> %v(%v)",
                strings.Join(param_names, ","),
                retval_name,
                exception_name,
            )
        }
        var name = fmt.Sprintf("[ %v ]", strings.Join(item_names, " | "))
        var T = (*TypeInfo)(unsafe.Pointer(&T_Function {
            __TypeInfo: TypeInfo {
                __Kind: TK_Function,
                __Name: name,
                __Initialized: true,
            },
            __Items: items,
        }))
        ctx.__RegisterInflatedType(T, -1, fingerprint)
        return T.__Id
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

func (G *GenericType) Validate(ctx *ObjectContext) *ValidationError {
    var invalid_super = func (e *TypeExpr) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_InvalidSuperType {
            __ValidationError: ValidationError {
                __Kind: VEK_InvalidSuperType,
                __Template: G,
            },
            __InvalidSuper: e,
        }))
    }
    var duplicate_super = func (super int) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_DuplicateSuperType {
            __ValidationError: ValidationError {
                __Kind: VEK_DuplicateSuperType,
                __Template: G,
            },
            __DuplicateType: ctx.GetGenericType(super),
        }))
    }
    var circular_inheritance = func (super int) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_CircularInheritance {
            __ValidationError: ValidationError {
                __Kind: VEK_CircularInheritance,
                __Template: G,
            },
            __CircularType: ctx.GetGenericType(super),
        }))
    }
    var not_extensible = func (S *GenericType) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_DeriveNonExtensible {
            __ValidationError: ValidationError {
                __Kind: VEK_DeriveNonExtensible,
                __Template: G,
            },
            __NonExtensibleType: S,
        }))
    }
    var wrong_mutability = func (S *GenericType, im bool) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_DeriveWrongMutability {
            __ValidationError: ValidationError {
                __Kind: VEK_DeriveWrongMutability,
                __Template: G,
            },
            __WrongSchema: S,
            __NeedImmutable: im,
        }))
    }
    var field_conflict = func (
        S1 *GenericType, S2 *GenericType, name Identifier,
    ) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_FieldConflict {
            __ValidationError: ValidationError {
                __Kind: VEK_FieldConflict,
                __Template: G,
            },
            __Schema1: S1,
            __Schema2: S2,
            __FieldName: name,
        }))
    }
    var method_conflict = func (
        C1 *GenericType, C2 *GenericType, name Identifier,
    ) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_MethodConflict {
            __ValidationError: ValidationError {
                __Kind: VEK_MethodConflict,
                __Template: G,
            },
            __Class1: C1,
            __Class2: C2,
            __MethodName: name,
        }))
    }
    var interface_not_implemented = func (
        I *GenericType, method Identifier, missing bool,
    ) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(
            &VE_InterfaceNotImplemented {
                __ValidationError: ValidationError {
                    __Kind: VEK_InterfaceNotImplemented,
                    __Template: G,
                },
                __Interface: I,
                __BadMethod: method,
                __IsMissing: missing,
            },
        ))
    }
    var inflation_error = func (err *InflationError) *ValidationError {
        return (*ValidationError)(unsafe.Pointer(&VE_InflationError {
            __ValidationError: ValidationError {
                __Kind: VEK_InflationError,
                __Template: G,
            },
            __InflationError: err,
        }))
    }
    // Check for Hierarchy Consistency
    if G.__Kind == GT_Schema {
        var G_as_Schema = (*GenericSchemaType)(unsafe.Pointer(G))
        var supers = [][2]int { [2]int { G.__Id, 0 } }
        var add_base_to_supers func(*TypeExpr, int) *ValidationError
        add_base_to_supers = func (base *TypeExpr, depth int) *ValidationError {
            if base.__Kind != TE_Inflation {
                return invalid_super(base)
            }
            var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
            var B = base_as_inf_expr.__Template
            if B.__Kind != GT_Schema {
                return invalid_super(base)
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
            for _, super := range supers {
                var existing = super[0]
                if existing == current {
                    var existing_depth = super[1]
                    if existing_depth == depth {
                        return duplicate_super(existing)
                    } else {
                        return circular_inheritance(existing)
                    }
                }
            }
            supers = append(supers, [2]int{ current, depth })
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
        for _, super := range supers {
            var S = ctx.GetGenericType(super[0])
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
        var supers = [][2]int { [2]int { G.__Id, 0 } }
        var add_base_to_supers func(*TypeExpr, int) *ValidationError
        add_base_to_supers = func (base *TypeExpr, depth int) *ValidationError {
            if base.__Kind != TE_Inflation {
                return invalid_super(base)
            }
            var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
            var B = base_as_inf_expr.__Template
            if B.__Kind != GT_Class {
                return invalid_super(base)
            }
            var B_as_Class = (*GenericClassType)(unsafe.Pointer(B))
            if !(B_as_Class.__Extensible) {
                return not_extensible(B)
            }
            var current = B.__Id
            for _, super := range supers {
                var existing = super[0]
                if existing == current {
                    var existing_depth = super[1]
                    if existing_depth == depth {
                        return duplicate_super(existing)
                    } else {
                        return circular_inheritance(existing)
                    }
                }
            }
            supers = append(supers, [2]int{ current, depth })
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
        for _, super := range supers {
            var S = ctx.GetGenericType(super[0])
            Assert(S.__Kind == GT_Class, "Generics: invalid generic class")
            var S_as_Class = (*GenericClassType)(unsafe.Pointer(S))
            for _, method := range S_as_Class.__OwnMethodList {
                var name = method.__Name
                var existing, exists = occurred_methods[name]
                if exists {
                    return method_conflict(existing, S, name)
                } else {
                    occurred_methods[name] = S
                    if super[0] == G.__Id {
                        own_method_types[name] = method.__Type
                    }
                }
            }
        }
        // Check if all interfaces are implemented by TypeExprEqual()
        for _, base := range G_as_Class.__BaseInterfaceList {
            if base.__Kind != TE_Inflation {
                return invalid_super(base)
            }
            var base_as_inf_expr = (*InflationTypeExpr)(unsafe.Pointer(base))
            var B = base_as_inf_expr.__Template
            if B.__Kind != GT_Interface {
                return invalid_super(base)
            }
            var B_as_Interface = (*GenericInterfaceType)(unsafe.Pointer(B))
            Assert (
                len(B_as_Interface.__MethodList) > 0,
                "Generics: empty interface is not valid",
            )
            for _, method := range B_as_Interface.__MethodList {
                var name = method.__Name
                var given_T, exists = own_method_types[name]
                if !exists {
                    return interface_not_implemented(B, name, true)
                } else {
                    var required_T = method.__Type
                    var ok = TypeExprEqual (
                        given_T, required_T,
                        base_as_inf_expr.__Arguments,
                    )
                    if !ok {
                        return interface_not_implemented(B, name, false)
                    }
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
    var inner_types = make([]*TypeExpr, 0)
    switch G.__Kind {
        // TODO
    }
    for _, e := range inner_types {
        var err = e.Check(ctx, args)
        if err != nil {
            return inflation_error(err)
        }
    }
    panic("unimplemented")
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

func (G *GenericType) Inflate(ctx *ObjectContext, args []int) int {
    // validating types is not the responsibility of this method
    var t, exists = ctx.GetInflatedType(G.__Id, args)
    if exists {
        return t
    }
    var T *TypeInfo = nil
    var name = G.GetInflatedName(ctx, args)
    var register = func (T_concrete unsafe.Pointer) {
        T = (*TypeInfo)(T_concrete)
        ctx.__RegisterInflatedType(T, G.__Id, args)
    }
    switch G.__Kind {
    case GT_Function:
        // note: this branch has its own return statement
        var g_function = (*GenericFunctionType)(unsafe.Pointer(G))
        var signature = g_function.__Signature.Evaluate(ctx, args)
        var F = ctx.GetType(signature)
        Assert(F.__Kind == TK_Function, "Generics: bad signature")
        return signature
    case GT_Union:
        var g_union = (*GenericUnionType)(unsafe.Pointer(G))
        var union = &T_Union {
            __TypeInfo: TypeInfo {
                __Kind: TK_Union,
                __Name: name,
                __Initialized: false,
                __FromGeneric: true,
                __GenericId: G.__Id,
                __GenericArgs: args,
            },
            // __Elements: nil
        }
        register(unsafe.Pointer(union))
        var elements = make([]int, len(g_union.__Elements))
        for i, element_expr := range g_union.__Elements {
            elements[i] = element_expr.Evaluate(ctx, args)
        }
        sort.Ints(elements)
        union.__Elements = elements
        union.__TypeInfo.__Initialized = true
    case GT_Trait:
        var g_trait = (*GenericTraitType)(unsafe.Pointer(G))
        var trait = &T_Trait {
            __TypeInfo: TypeInfo {
                __Kind: TK_Trait,
                __Name: name,
                __Initialized: false,
                __FromGeneric: true,
                __GenericId: G.__Id,
                __GenericArgs: args,
            },
            // __Constraints: nil
        }
        register(unsafe.Pointer(trait))
        var constraints = make([]int, len(g_trait.__Constraints))
        for i, constraint_expr := range g_trait.__Constraints {
            constraints[i] = constraint_expr.Evaluate(ctx, args)
        }
        sort.Ints(constraints)
        trait.__Constraints = constraints
        trait.__TypeInfo.__Initialized = true
    case GT_Schema:
        var g_schema = (*GenericSchemaType)(unsafe.Pointer(G))
        var schema = &T_Schema {
            __TypeInfo: TypeInfo {
                __Kind: TK_Schema,
                __Name: name,
                __Initialized: false,
                __FromGeneric: true,
                __GenericId: G.__Id,
                __GenericArgs: args,
            },
            __Immutable: g_schema.__Immutable,
            __Extensible: g_schema.__Extensible,
            // Bases, Supers, Fields, OffsetTable = nil
        }
        register(unsafe.Pointer(schema))
        var this_id = schema.__TypeInfo.__Id
        var bases = make([]int, len(g_schema.__BaseList))
        for i, base_expr := range g_schema.__BaseList {
            bases[i] = base_expr.Evaluate(ctx, args)
        }
        sort.Ints(bases)
        var supers = []int { this_id }
        for _, base := range bases {
            var B = ctx.GetType(base)
            Assert(B.__Kind == TK_Schema, "Generics: bad base schema")
            Assert(B.__Initialized, "Generics: circular inheritance")
            var Base = (*T_Schema)(unsafe.Pointer(B))
            for _, super_of_base := range Base.__Supers {
                supers = append(supers, super_of_base)
            }
        }
        var offset_table = make(map[Identifier] int)
        var fields = make([]SchemaField, 0)
        for _, field := range g_schema.__OwnFieldList {
            var offset = len(fields)
            var name = field.__Name
            var _, exists = offset_table[name]
            Assert(!exists, "Generics: duplicate schema field")
            offset_table[name] = offset
            fields = append(fields, field.Evaluate(ctx, args, this_id))
        }
        for _, super := range supers {
            if super != this_id {
                var S = ctx.GetType(super)
                Assert(S.__Kind == TK_Schema, "Generics: bad super schema")
                Assert(S.__Initialized, "Generics: circular inheritance")
                var Super = (*T_Schema)(unsafe.Pointer(S))
                for _, field := range Super.__Fields {
                    var offset = len(fields)
                    var name = field.__Name
                    var _, exists = offset_table[name]
                    Assert(!exists, "Generics: duplicate schema field")
                    offset_table[name] = offset
                    fields = append(fields, field)
                }
            }
        }
        schema.__Bases = bases
        schema.__Supers = supers
        schema.__Fields = fields
        schema.__OffsetTable = offset_table
        schema.__TypeInfo.__Initialized = true
    case GT_Class:
        var g_class = (*GenericClassType)(unsafe.Pointer(G))
        var class = &T_Class {
            __TypeInfo: TypeInfo {
                __Kind: TK_Class,
                __Name: name,
                __Initialized: false,
                __FromGeneric: true,
                __GenericId: G.__Id,
                __GenericArgs: args,
            },
            __Extensible: g_class.__Extensible,
            // Methods, {Base,Super}{Classes,Interfaces} = nil
         }
         register(unsafe.Pointer(class))
         var this_id = class.__TypeInfo.__Id
         var base_classes = make([]int, len(g_class.__BaseClassList))
         for i, base_class_expr := range g_class.__BaseClassList {
             base_classes[i] = base_class_expr.Evaluate(ctx, args)
         }
         sort.Ints(base_classes)
         var base_interfaces = make([]int, len(g_class.__BaseInterfaceList))
         for i, base_interface_expr := range g_class.__BaseInterfaceList {
             base_interfaces[i] = base_interface_expr.Evaluate(ctx, args)
         }
         sort.Ints(base_interfaces)
         var super_classes = []int { this_id }
         var offset_table = []int { 0 }
         for i, base_class := range base_classes {
             var B = ctx.GetType(base_class)
             Assert(B.__Kind == TK_Class, "Generics: bad base class")
             Assert(B.__Initialized, "Generics: circular inheritance")
             var Base = (*T_Class)(unsafe.Pointer(B))
             for _, super_of_base := range Base.__SuperClasses {
                 offset_table = append(offset_table, i)
                 super_classes = append(super_classes, super_of_base)
             }
         }
         var methods = make(map[Identifier]MethodInfo)
         for _, own_method := range g_class.__OwnMethodList {
             var name = own_method.__Name
             var _, exists = methods[name]
             Assert(!exists, "Generics: duplicate method name")
             methods[name] = MethodInfo {
                 __Type: own_method.__Type.Evaluate(ctx, args),
                 __From: this_id,
                 __Offset: 0,
                 __FunInfo: own_method.__FunInfo,
             }
         }
         var super_interfaces = append([]int {}, base_interfaces...)
         for i, super_class := range super_classes {
             if super_class != this_id {
                 var S = ctx.GetType(super_class)
                 Assert(S.__Kind == TK_Class, "Generics: bad super class")
                 Assert(S.__Initialized, "Generics: circular inheritance")
                 var Super = (*T_Class)(unsafe.Pointer(S))
                 for _, sup_i := range Super.__SuperInterfaces {
                     super_interfaces = append(super_interfaces, sup_i)
                 }
                 for name, sup_m := range Super.__Methods {
                     var _, exists = methods[name]
                     Assert(!exists, "Generics: duplicate method name")
                     methods[name] = MethodInfo {
                         __Type: sup_m.__Type,
                         __From: sup_m.__From,
                         __Offset: offset_table[i],
                         __FunInfo: sup_m.__FunInfo,
                     }
                 }
             }
         }
         class.__Methods = methods
         class.__BaseClasses = base_classes
         class.__BaseInterfaces = base_interfaces
         class.__SuperClasses = super_classes
         class.__SuperInterfaces = super_interfaces
         class.__TypeInfo.__Initialized = true
    case GT_Interface:
        var g_interface = (*GenericInterfaceType)(unsafe.Pointer(G))
        var interface_ = &T_Interface {
            __TypeInfo: TypeInfo {
                __Kind: TK_Interface,
                __Name: name,
                __Initialized: false,
                __FromGeneric: true,
                __GenericId: G.__Id,
                __GenericArgs: args,
            },
            // MethodTypes = nil
        }
        register(unsafe.Pointer(interface_))
        var method_types = make(map[Identifier] int)
        for _, method := range g_interface.__MethodList {
            var name = method.__Name
            var _, exists = method_types[name]
            Assert(!exists, "Generics: duplicate interface method")
            method_types[name] = method.__Type.Evaluate(ctx, args)
        }
        interface_.__MethodTypes = method_types
        interface_.__TypeInfo.__Initialized = true
    default:
        panic("impossible branch")
    }
    return T.__Id
}

func (G *GenericType) GetInflatedName(ctx *ObjectContext, args []int) string {
    if len(args) == 0 {
        return G.__Name
    } else {
        var arg_names = make([]string, len(args))
        for i, arg := range args {
            arg_names[i] = ctx.GetTypeName(arg)
        }
        return fmt.Sprintf (
            "%v[%v]",
            G.__Name, strings.Join(arg_names, ", "),
        )
    }
}

func (field *GenericSchemaField) Evaluate (
    ctx *ObjectContext, args []int, from int,
) SchemaField {
    return SchemaField {
        __Name: field.__Name,
        __Type: field.__Type.Evaluate(ctx, args),
        __HasDefault: field.__HasDefault,
        __DefaultValue: field.__DefaultValue,
        __From: from,
    }
}
