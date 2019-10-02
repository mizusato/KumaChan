package object

import "fmt"
import "sort"
import "unsafe"
import "strings"
import ."kumachan/interpreter/assertion"

/**
 *  There are 6 kinds of generic type that can be defined by user:
 *      1. Union Type (flat distinct union)
 *          e.g. union Maybe[T] = T | Nil
 *      2. Trait Type (intersection of classes/interfaces)
 *          e.g. trait GetterSetter[K,V] = Getter[K,V] & Setter[K,V]
 *      3. Schema Type (definition of struct) 
 *          e.g. struct Complex { real: Float, imag: Float }
 *               * this example has no type parameter,
 *                 but is also a generic type. (regarded as Complex[])
 *      4. Class Type
 *          e.g. class Array[T] implements Getter[Size, T], ...
 *      5. Interface Type
 *          e.g. interface Getter[K,V] { get(key: K) -> V }
 *      6. Function Type (along with function/method declaration)
 *          e.g. function map[F,T](i:Iterable[F], f:[F->T]) -> Iterator[T] ...
 *   A generic type is also called a template.
 *   Instantiating a generic type is also called inflating it.
 *   To get a concrete type (*TypeInfo), inflating is required.
 *   Before inflating a generic type, its type declaration must be
 *       validated. See `generics_validate.go` for more information.
 *   It is impossible to compare two generic types without inflating them,
 *       in other words, only concrete types can be compared.
 */

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
    GT_NativeClass
)

type GenericFunctionType struct {
    __GenericType  GenericType
    __Signature    *TypeExpr
}

type GenericUnionType struct {
    __GenericType  GenericType
    __Elements     [] *TypeExpr
}

type GenericTraitType struct {
    __GenericType  GenericType
    __Constraints  [] *TypeExpr
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

type GenericNativeClassType struct {
    __GenericType   GenericType
    __MethodList    [] GenericInterfaceMethod
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
            items[i] = FunctionTypeItem {
                __Parameters: params,
                __ReturnValue: retval,
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
            item_names[i] = fmt.Sprintf (
                "%v -> %v",
                strings.Join(param_names, ","),
                retval_name,
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



func (G *GenericType) Inflate(ctx *ObjectContext, args []int) int {
    Assert(G.__Validated, "Generics: inflating non-validated generic type")
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
    case GT_NativeClass:
        var g_native = (*GenericNativeClassType)(unsafe.Pointer(G))
        var native = &T_NativeClass {
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
        register(unsafe.Pointer(native))
        var method_types = make(map[Identifier] int)
        for _, method := range g_native.__MethodList {
            var name = method.__Name
            var _, exists = method_types[name]
            Assert(!exists, "Generics: duplicate native class method")
            method_types[name] = method.__Type.Evaluate(ctx, args)
        }
        native.__MethodTypes = method_types
        native.__TypeInfo.__Initialized = true
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
