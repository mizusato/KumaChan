package object

import "fmt"
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
    __DefaultValue   Object
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
    var t, exists = ctx.GetInflatedType(G.__Id, args)
    if exists {
        return t
    }
    var T *TypeInfo = nil
    var name = G.GetInflatedName(ctx, args)
    switch G.__Kind {
    case GT_Union:
        var union = (*GenericUnionType)(unsafe.Pointer(G))
        var elements = make([]int, len(union.__Elements))
        for i, element_expr := range union.__Elements {
            elements[i] = element_expr.Evaluate(ctx)
        }
        T = (*TypeInfo)(unsafe.Pointer(&T_Union {
            __TypeInfo: TypeInfo {
                __Kind: TK_Union,
                __Name: name,
                __Initialized: true,
            },
            __Elements: elements,
        }))
    case GT_Trait:
        var trait = (*GenericTraitType)(unsafe.Pointer(G))
        var constraints = make([]int, len(trait.__Constraints))
        for i, constraint_expr := range trait.__Constraints {
            constraints[i] = constraint_expr.Evaluate(ctx)
        }
        T = (*TypeInfo)(unsafe.Pointer(&T_Trait {
            __TypeInfo: TypeInfo {
                __Kind: TK_Trait,
                __Name: name,
                __Initialized: true,
            },
            __Constraints: constraints,
        }))
    case GT_Schema:
        // var schema = (*GenericSchemaType)(unsafe.Pointer(G))
        // TODO
    }
    ctx.__RegisterInflatedType(T, G.__Id, args)
    return T.__Id
}
