package object

import "fmt"
import "sort"
import "unsafe"
import "strings"

type TypeExpr struct {
    __Kind TypeExprKind
}

type TypeExprKind int
const (
    TE_Final TypeExprKind = iota
    TE_Function
    TE_Inflation
)

type FinalTypeExpr struct {
    __TypeExpr  TypeExpr
    __Type      int
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
}

type GenericTypeParameter struct {
    __Name            string
    __HasUpperBound   bool
    __UpperBound      int
}

type GenericTypeKind int
const (
    GT_Union GenericTypeKind = iota
    GT_Trait
    GT_Schema
    GT_Class
    GT_Interface
)

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
    __OwnMethodList       map[Identifier] *GenericClassMethod
}

type GenericClassMethod struct {
    __Type       *TypeExpr
    __Function   *Function
}

type GenericInterfaceType struct {
    __GenericType  GenericType
    __MethodList   [] *TypeExpr
}


func (e *TypeExpr) Evaluate(ctx *ObjectContext) int {
    switch e.__Kind {
    case TE_Final:
        return (*FinalTypeExpr)(unsafe.Pointer(e)).__Type
    case TE_Function:
        var f_expr = (*FunctionTypeExpr)(unsafe.Pointer(e))
        var items = make([]T_Function_Item, len(f_expr.__Items))
        var item_names = make([]string, len(items))
        for i, f := range f_expr.__Items {
            var params = make([]int, len(f.__Parameters))
            var param_names = make([]string, len(params))
            for j, ei := range f.__Parameters {
                params[j] = ei.Evaluate(ctx)
                param_names[j] = ctx.GetTypeName(params[j])
            }
            var retval = f.__ReturnValue.Evaluate(ctx)
            var retval_name = ctx.GetTypeName(retval)
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
        var name = fmt.Sprintf("[%v]", strings.Join(item_names, " | "))
        var T = (*TypeInfo)(unsafe.Pointer(&T_Function {
            __TypeInfo: TypeInfo {
                __Kind: TK_Function,
                __Name: name,
            },
            __Items: items,
        }))
        ctx.__RegisterType(T)
        return T.__Id
    case TE_Inflation:
        var inf_expr = (*InflationTypeExpr)(unsafe.Pointer(e))
        var args = make([]int, len(inf_expr.__Arguments))
        for i, arg_expr := range inf_expr.__Arguments {
            args[i] = arg_expr.Evaluate(ctx)
        }
        var template = inf_expr.__Template
        return template.Inflate(ctx, args)
    }
    return 0
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
            elements[i] = element_expr.Evaluate(ctx)
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
            constraints[i] = constraint_expr.Evaluate(ctx)
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
            // Bases, Supers, Fields = nil
        }
        register(unsafe.Pointer(schema))
        var bases = make([]int, len(g_schema.__BaseList))
        for i, base_expr := range g_schema.__BaseList {
            bases[i] = base_expr.Evaluate(ctx)
        }
        sort.Ints(bases)
        var supers = []int { T.__Id }
        var fields = make([]SchemaField, 0)
        for _, field := range g_schema.__OwnFieldList {
            fields = append(fields, field.Evaluate(ctx, T.__Id))
        }
        schema.__Bases = bases
        schema.__Supers = supers
        schema.__Fields = fields
        if len(bases) == 0 {
            T.__Initialized = true
        }
        // non-trivial supers and non-own fields will be added while validating
    }
    return T.__Id
}

func (field *GenericSchemaField) Evaluate (
    ctx *ObjectContext, from int,
) SchemaField {
    return SchemaField {
        __Name: field.__Name,
        __Type: field.__Type.Evaluate(ctx),
        __HasDefault: field.__HasDefault,
        __DefaultValue: field.__DefaultValue,
        __From: from,
    }
}
