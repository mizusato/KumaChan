package transpiler

import "fmt"
import "strings"


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
}
