package transpiler

import "fmt"
import "strings"


var TypeMap = map[string]TransFunction {
    // type = fun_sig | type_expr | ( expr )!
    "type": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr, is_expr = children["expr"]
        if is_expr {
            return Transpile(tree, expr)
        } else {
            return TranspileFirstChild(tree, ptr)
        }
    },
    // fun_sig = @$ < opt_typelist >! <! type! >!
    "fun_sig": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var l_ptr = children["opt_typelist"]
        var args string
        if Empty(tree, l_ptr) {
            args = "[]"
        } else {
            // opt_typelist? = typelist
            args = TranspileFirstChild(tree, l_ptr)
        }
        var ret = Transpile(tree, children["type"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "__.c(__.cfs, [%v, %v], %v, %v, %v)",
            args, ret, file, row, col,
        )
    },
    // typelist = type typelist_tail
    "typelist": func (tree Tree, ptr int) string {
        return TranspileSubTree(tree, ptr, "type", "typelist_tail")
    },
    // type_expr = identifier type_gets type_args
    "type_expr": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var children = Children(tree, ptr)
        var id = Transpile(tree, children["identifier"])
        var gets_ptr = children["type_gets"]
        var gets = FlatSubTree(tree, gets_ptr, "type_get", "type_gets")
        var t = id
        for _, get := range gets {
            // type_get = . name
            var key = TranspileLastChild(tree, get)
            var row, col = GetRowColInfo(tree, get)
            var buf strings.Builder
            buf.WriteString("__.g")
            buf.WriteRune('(')
            WriteList(&buf, []string {
                t, key, "false", file, row, col,
            })
            buf.WriteRune(')')
            t = buf.String()
        }
        var args_ptr = children["type_args"]
        if NotEmpty(tree, args_ptr) {
            var args = Transpile(tree, args_ptr)
            var row, col = GetRowColInfo(tree, args_ptr)
            var buf strings.Builder
            buf.WriteString("__.c")
            buf.WriteRune('(')
            WriteList(&buf, []string {
                t, args, file, row, col,
            })
            buf.WriteRune(')')
            return buf.String()
        } else {
            return t
        }
    },
    // type_args? = Call < type_arglist! >!
    "type_args": TranspileChild("type_arglist"),
    // type_arglist = type_arg type_arglist_tail
    "type_arglist": func (tree Tree, ptr int) string {
        return TranspileSubTree(tree, ptr, "type_arg", "type_arglist_tail")
    },
    // type_arg = type | primitive
    "type_arg": TranspileFirstChild,
    // type_literal = simple_type_literal | finite_literal
    "type_literal": TranspileFirstChild,
    // simple_type_literal = { name _bar1 filters! }!
    "simple_type_literal": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = GetTokenContent(tree, children["name"])
        var p_name = Transpile(tree, children["name"])
        var parameters = fmt.Sprintf("[{ name: %v, type: __.a }]", p_name)
        var proto = fmt.Sprintf (
            "{ parameters: %v, value_type: __.b }",
            parameters,
        )
        var desc = Desc (
            []rune("lambda.type_checker"),
            []rune(fmt.Sprintf("(%v: Any)", string(name))),
            []rune("Bool"),
        )
        var checker_expr = Transpile(tree, children["filters"])
        var raw = BareFunction(fmt.Sprintf("return %v;", checker_expr))
        var checker = fmt.Sprintf (
            "w(%v, %v, %v, %v)",
            proto, "null", desc, raw,
        )
        return fmt.Sprintf("__.ct(%v)", checker)
    },
    // filters = exprlist
    "filters": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        return Filters(tree, children["exprlist"])
    },
    // finite_literal = @one @of { exprlist }! | { exprlist }
    "finite_literal": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "__.c(__.cft, %v, %v, %v, %v)",
            TranspileSubTree (
                tree, children["exprlist"], "expr", "exprlist_tail",
            ),
            file, row, col,
        )
    },
    // enum = @enum name {! namelist! }!
    "enum": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var e_name = Transpile(tree, children["name"])
        var l_ptr = children["namelist"]
        var name_ptrs = FlatSubTree(tree, l_ptr, "name", "namelist_tail")
        var names = make([]string, 0, 16)
        for _, name_ptr := range name_ptrs {
            names = append(names, Transpile(tree, name_ptr))
        }
        var names_str = strings.Join(names, ", ")
        var enum = fmt.Sprintf("__.ce(%v, [%v])", e_name, names_str)
        return fmt.Sprintf (
            "__.c(dl, [%v, %v, true, __.t], %v, %v, %v)",
            e_name, enum, file, row, col,
        )
    },
    // schema = @struct name generic_params { field_list }! schema_config
    "schema": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var gp_ptr = children["generic_params"]
        var name = Transpile(tree, name_ptr)
        var config = Transpile(tree, children["schema_config"])
        var table, defaults, contains = FieldList(tree, children["field_list"])
        var schema = fmt.Sprintf (
            "__.c(__.cs, [%v, %v, %v, %v, %v], %v, %v, %v)",
            name, table, defaults, contains, config, file, row, col,
        )
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, schema)
        } else {
            value = schema
        }
        return fmt.Sprintf (
            "__.c(dl, [%v, %v, true, __.t], %v, %v, %v)",
            name, value, file, row, col,
        )
    },
    // schema_config? = @config { struct_guard operator_defs }!
    "schema_config": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "{ guard: null, ops: {} }" }
        var children = Children(tree, ptr)
        return fmt.Sprintf (
            "{ guard: %v, ops: %v }",
            Transpile(tree, children["struct_guard"]),
            Transpile(tree, children["operator_defs"]),
        )
    },
    // struct_guard? = @guard body!
    "struct_guard": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "null" }
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var parameters = "[{ name: 'fields', type: __.h }]"
        var desc = Desc (
            []rune("struct_guard"),
            []rune("(fields: Hash)"),
            []rune("Void"),
        )
        return Function(tree, body_ptr, F_Sync, desc, parameters, "__.v")
    },
}
