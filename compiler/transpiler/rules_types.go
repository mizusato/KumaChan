package transpiler

import "fmt"
import "strings"
import "../syntax"


var Types = map[string]TransFunction {
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
    // finite_literal = @one @of { exprlist_opt }! | { exprlist }
    "finite_literal": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var exprlist_ptr, ok = children["exprlist"]
        if !ok {
            // exprlist_opt? = exprlist
            var opt_ptr = children["exprlist_opt"]
            if NotEmpty(tree, opt_ptr) {
                exprlist_ptr = tree.Nodes[opt_ptr].Children[0]
            } else {
                return "__.cf()"
            }
        }
        var expr_ptrs = FlatSubTree(tree, exprlist_ptr, "expr", "exprlist_tail")
        var buf strings.Builder
        buf.WriteString("__.cf")
        buf.WriteRune('(')
        for i, expr_ptr := range expr_ptrs {
            var expr = Transpile(tree, expr_ptr)
            buf.WriteString(expr)
            if i != len(expr_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(')')
        return buf.String()
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
    // schema = @struct name generic_params { field_list schema_config }!
    "schema": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var gp_ptr = children["generic_params"]
        var name = Transpile(tree, name_ptr)
        var config = Transpile(tree, children["schema_config"])
        var table, defaults = FieldList(tree, children["field_list"])
        var schema = fmt.Sprintf (
            "__.cs(%v, %v, %v, %v)",
            name, table, defaults, config,
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
    // schema_config? = , @config { schema_req schema_op_defs }!
    "schema_config": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            return fmt.Sprintf (
                "{ req: %v, ops: %v }",
                Transpile(tree, children["schema_req"]),
                Transpile(tree, children["schema_op_defs"]),
            )
        } else {
            return "{ req: null, ops: {} }"
        }
    },
    // schema_req? = @require (! name! )! opt_arrow body!
    "schema_req": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var name_ptr = children["name"]
            var body_ptr = children["body"]
            var name = Transpile(tree, name_ptr)
            var name_raw = GetTokenContent(tree, name_ptr)
            var desc = Desc (
                []rune("schema_requirement"), name_raw, []rune("Bool"),
            )
            var parameters = fmt.Sprintf("[{ name: %v, type: __.a }]", name)
            return Function(tree, body_ptr, F_Sync, desc, parameters, "__.b")
        } else {
            return "null"
        }
    },
    // schema_op_defs? = schema_op_def schema_op_defs
    "schema_op_defs": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var ds = FlatSubTree(tree, ptr, "schema_op_def", "schema_op_defs")
            var buf strings.Builder
            buf.WriteString("{ ")
            for i, def_ptr := range ds {
                // schema_op_def = @operator schema_op schema_op_fun
                // schema_op = @str | < | + | - | * | / | %
                var children = Children(tree, def_ptr)
                var op_ptr = tree.Nodes[children["schema_op"]].Children[0]
                var op_match = syntax.Id2Name[tree.Nodes[op_ptr].Part.Id]
                var op_name = strings.TrimPrefix(op_match, "@")
                var op_fun = Transpile(tree, children["schema_op_fun"])
                var op_escaped = EscapeRawString([]rune(op_name))
                fmt.Fprintf(&buf, "%v: %v", op_escaped, op_fun)
                if i != len(ds)-1 {
                    buf.WriteString(", ")
                }
            }
            buf.WriteString(" }")
            return buf.String()
        } else {
            return "{}"
        }
    },
    // schema_op_fun = (! namelist! )! opt_arrow body!
    "schema_op_fun": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var namelist_ptr = children["namelist"]
        var body_ptr = children["body"]
        var parameters = UntypedParameterList(tree, namelist_ptr)
        var desc = Desc (
            []rune("schema_operator"),
            GetWholeContent(tree, namelist_ptr),
            []rune("Object"),
        )
        return Function(tree, body_ptr, F_Sync, desc, parameters, "__.a")
    },
}
