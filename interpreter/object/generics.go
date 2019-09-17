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
    __HasUpperBound   bool
    __UpperBound      *TypeExpr
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
}

type GenericTraitType struct {
    __GenericType  GenericType
    __Constraints  [] *TypeExpr
}

type GenericSchemaType struct {
    __GenericType     GenericType
    __Immutable       bool
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

func (e *TypeExpr) IsValidBound() (bool, *TypeExpr, *GenericType) {
    switch e.__Kind {
    case TE_Final:
        return true, nil, nil
    case TE_Argument:
        return true, nil, nil
    case TE_Function:
        var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
        var ok = true
        var te *TypeExpr
        var gt *GenericType
        for _, f := range f_expr.__Items {
            for _, p := range f.__Parameters {
                ok, te, gt = p.IsValidBound()
                if !ok { break }
            }
            ok, te, gt = f.__ReturnValue.IsValidBound()
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
                ok, te, gt = arg_expr.IsValidBound()
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
            fingerprint = append(fingerprint, retval)
            fingerprint = append(fingerprint, -1)
            var name = fmt.Sprintf (
                "%v -> %v",
                strings.Join(param_names, ", "), retval_name,
            )
            items[i] = T_Function_Item {
                __Parameters: params,
                __ReturnValue: retval,
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
        if parameter.__HasUpperBound {
            ok, te, gt = parameter.__UpperBound.IsValidBound()
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
        }
        ctx.__RegisterType((*TypeInfo)(unsafe.Pointer(T_arg)))
        args[i] = T_arg.__TypeInfo.__Id
        T_args[i] = T_arg
    }
    for i, T_arg := range T_args {
        var parameter = G.__Parameters[i]
        if parameter.__HasUpperBound {
            T_arg.__UpperBound = parameter.__UpperBound.Evaluate(ctx, args)
            T_arg.__TypeInfo.__Initialized = true
        }
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
        T.__Initialized = true
    case GT_Trait:
        var g_trait = (*GenericTraitType)(unsafe.Pointer(G))
        var trait = &T_Trait {
            __TypeInfo: TypeInfo {
                __Kind: TK_Trait,
                __Name: name,
                __Initialized: false,
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
        T.__Initialized = true
    case GT_Schema:
        var g_schema = (*GenericSchemaType)(unsafe.Pointer(G))
        var schema = &T_Schema {
            __TypeInfo: TypeInfo {
                __Kind: TK_Schema,
                __Name: name,
                __Initialized: false,
            },
            __Immutable: g_schema.__Immutable,
            // Bases, Supers, Fields, OffsetTable = nil
        }
        register(unsafe.Pointer(schema))
        var bases = make([]int, len(g_schema.__BaseList))
        for i, base_expr := range g_schema.__BaseList {
            bases[i] = base_expr.Evaluate(ctx, args)
        }
        sort.Ints(bases)
        var supers = []int { T.__Id }
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
            fields = append(fields, field.Evaluate(ctx, args, T.__Id))
        }
        for _, super := range supers {
            if super != T.__Id {
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
        T.__Initialized = true
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
