package transpiler

import "fmt"
import "strings"


var Containers = map[string]TransFunction {
    // hash = { } | { hash_item! hash_tail }!
    "hash": func (tree Tree, ptr int) string {
        if tree.Nodes[ptr].Length == 2 {
            return "{}"
        }
        var items = FlatSubTree(tree, ptr, "hash_item", "hash_tail")
        var names = make(map[string]bool)
        var buf strings.Builder
        buf.WriteString("{ ")
        for i, item := range items {
            // hash_item = name :! expr! | string :! expr! | :: name!
            var children = Children(tree, item)
            var name_ptr, has_name = children["name"]
            var expr_ptr, has_expr = children["expr"]
            var name string
            if has_name {
                name = Transpile(tree, name_ptr)
            } else {
                name = Transpile(tree, children["string"])
            }
            var _, exists = names[name]
            if exists {
                // this check requires `Transpile(tree, name_ptr)`
                // and `Transpile(tree, children["string"])`
                // using the same quotes, which is promised by EscapeRawString()
                panic("duplicate hash key " + name)
            }
            names[name] = true
            var expr string
            if has_expr {
                expr = Transpile(tree, expr_ptr)
            } else {
                expr = VarLookup(GetTokenContent(tree, name_ptr))
            }
            fmt.Fprintf(&buf, "%v: %v", name, expr)
            if i != len(items)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteString(" }")
        return buf.String()
    },
    // list = [ ] | [ list_item! list_tail ]!
    "list": func (tree Tree, ptr int) string {
        if tree.Nodes[ptr].Length == 2 {
            return "[]"
        }
        var items = FlatSubTree(tree, ptr, "list_item", "list_tail")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, item := range items {
            buf.WriteString(Transpile(tree, item))
            if i != len(items)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
    },
    // list_item = expr
    "list_item": TranspileFirstChild,
    // comprehension = .[ comp_rule! ]! | [ comp_rule ]!
    "comprehension": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var _, is_iterator = children[".["]
        var rule = Transpile(tree, children["comp_rule"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var f string
        if is_iterator {
            f = "__.ic"
        } else {
            f = "__.lc"
        }
        return fmt.Sprintf (
            "__.c(%v, [%v], %v, %v, %v)",
            f, rule, file, row, col,
        )
    },
    // comp_rule = expr _bar1 in_list! opt_filters
    "comp_rule": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var val_expr = Transpile(tree, children["expr"])
        var list_ptr = children["in_list"]
        // in_list = in_item in_list_tail
        var item_ptrs = FlatSubTree(tree, list_ptr, "in_item", "in_list_tail")
        var names = make([]string, 0, 10)
        var iterators = make([]string, 0, 10)
        for _, item_ptr := range item_ptrs {
            // in_item = name @in expr
            var item_children = Children(tree, item_ptr)
            var name = GetTokenContent(tree, item_children["name"])
            var expr = Transpile(tree, item_children["expr"])
            names = append(names, string(name))
            iterators = append(iterators, expr)
        }
        var parameters = UntypedParameters(names)
        var proto = fmt.Sprintf (
            "{ parameters: %v, value_type: __.a }",
            parameters,
        )
        var val_raw = BareFunction(fmt.Sprintf("return %v;", val_expr))
        var val_desc = EscapeRawString([]rune("comprehension.value_function"))
        var val = fmt.Sprintf (
            "w(%v, %v, %v, %v)",
            proto, "null", val_desc, val_raw,
        )
        var iterator_list = fmt.Sprintf (
            "[%v]", strings.Join(iterators, ", "),
        )
        var filter_expr = Transpile(tree, children["opt_filters"])
        var filter_raw = BareFunction(fmt.Sprintf("return %v;", filter_expr))
        var filter_desc = EscapeRawString([]rune("comprehension.filter"))
        var filter = fmt.Sprintf (
            "w(%v, %v, %v, %v)",
            proto, "null", filter_desc, filter_raw,
        )
        return fmt.Sprintf("%v, %v, %v", val, iterator_list, filter)
    },
    // opt_filters? = , exprlist
    "opt_filters": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            return Filters(tree, children["exprlist"])
        } else {
            return "true"
        }
    },
}
