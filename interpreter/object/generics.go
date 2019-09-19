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
    __Index  int
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
    __Checked      bool
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
    // TODO: element should be singleton or final schema/class
}

type GenericTraitType struct {
    __GenericType  GenericType
    __Constraints  [] *TypeExpr
    // TODO: element should be non-final schema or interface
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

func (e *TypeExpr) IsValidBound(depth int) (bool, *TypeExpr, *GenericType) {
    switch e.__Kind {
    case TE_Final:
        return true, nil, nil
    case TE_Argument:
        // should forbid [T < G[T]] and [T, U < G[T]]
        if depth == 0 {
            return true, nil, nil
        } else {
            return false, e, nil
        }
    case TE_Function:
        var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
        var ok = true
        var te *TypeExpr
        var gt *GenericType
        for _, f := range f_expr.__Items {
            for _, p := range f.__Parameters {
                ok, te, gt = p.IsValidBound(depth+1)
                if !ok { break }
            }
            ok, te, gt = f.__ReturnValue.IsValidBound(depth+1)
            if !ok { break }
            ok, te, gt = f.__Exception.IsValidBound(depth+1)
            if !ok { break }
        }
        return ok, te, gt
    case TE_Inflation:
        var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
        if inf_expr.__Template.__Checked {
            var ok = true
            var te *TypeExpr
            var gt *GenericType
            for _, arg_expr := range inf_expr.__Arguments {
                ok, te, gt = arg_expr.IsValidBound(depth+1)
                if !ok { break }
            }
            return ok, te, gt
        } else {
            return false, e, inf_expr.__Template
        }
    default:
        panic("impossible branch")
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
        var items = make([]T_Function_Item, len(f_expr.__Items))
        var item_names = make([]string, len(items))
        var fingerprint = make([]int, 0)
        for i, f := range f_expr.__Items {
            var params = make([]int, len(f.__Parameters))
            var param_names = make([]string, len(params))
            for j, ei := range f.__Parameters {
                params[j] = ei.Evaluate(ctx, args)
                param_names[j] = ctx.GetTypeName(params[j])
                fingerprint = append(fingerprint, params[j])
            }
            var retval = f.__ReturnValue.Evaluate(ctx, args)
            var retval_name = ctx.GetTypeName(retval)
            var exception = f.__Exception.Evaluate(ctx, args)
            var exception_name = ctx.GetTypeName(exception)
            fingerprint = append(fingerprint, retval)
            fingerprint = append(fingerprint, exception)
            fingerprint = append(fingerprint, -1)
            var name = fmt.Sprintf (
                "%v -> %v(%v)",
                strings.Join(param_names, ", "),
                retval_name,
                exception_name,
            )
            items[i] = T_Function_Item {
                __Parameters: params,
                __ReturnValue: retval,
                __Exception: exception,
            }
            item_names[i] = name
        }
        var cache, exists = ctx.GetInflatedType(-1, fingerprint)
        if exists {
            return cache
        } else {
            var name = fmt.Sprintf("[%v]", strings.Join(item_names, " | "))
            var T = (*TypeInfo)(unsafe.Pointer(&T_Function {
                __TypeInfo: TypeInfo {
                    __Kind: TK_Function,
                    __Name: name,
                },
                __Items: items,
            }))
            ctx.__RegisterInflatedType(T, -1, fingerprint)
            return T.__Id
        }
    case TE_Inflation:
        var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
        var args = make([]int, len(inf_expr.__Arguments))
        for i, arg_expr := range inf_expr.__Arguments {
            args[i] = arg_expr.Evaluate(ctx, args)
        }
        var template = inf_expr.__Template
        return template.Inflate(ctx, args)
    default:
        panic("impossible branch")
    }
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

func (G *GenericType) IsArgBoundsValid() (bool, *TypeExpr, *GenericType) {
    var ok = true
    var te *TypeExpr
    var gt *GenericType
    for _, parameter := range G.__Parameters {
        if parameter.__UpperBound != nil {
            ok, te, gt = parameter.__UpperBound.IsValidBound(0)
            if !ok { break }
        }
        if parameter.__LowerBound != nil {
            ok, te, gt = parameter.__LowerBound.IsValidBound(0)
            if !ok { break }
        }
    }
    return ok, te, gt
}

func (G *GenericType) GetArgPlaceholders(ctx *ObjectContext) []int {
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
            var B = parameter.__UpperBound.Evaluate(ctx, args)
            if B != T_args[i].__TypeInfo.__Id {
                T_args[i].__UpperBound = B
            }
        }
        if parameter.__LowerBound != nil {
            var B = parameter.__LowerBound.Evaluate(ctx, args)
            if B != T_args[i].__TypeInfo.__Id {
                T_args[i].__LowerBound = B
            }
        }
        T_args[i].__TypeInfo.__Initialized = true
    }
    return args
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
