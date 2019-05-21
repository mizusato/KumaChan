package transpiler

import "fmt"
import "strings"


var Containers = map[string]TransFunction {
    // map = @map { } | @map { map_item! map_tail }
    "map": func (tree Tree, ptr int) string {
        if tree.Nodes[ptr].Length == 3 {
            return "(new Map())"
        }
        var items = FlatSubTree(tree, ptr, "map_item", "map_tail")
        var buf strings.Builder
        buf.WriteRune('(')
        buf.WriteString("new Map")
        buf.WriteString("([")
        for i, item := range items {
            buf.WriteString(Transpile(tree, item))
            if i != len(items)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteString("])")
        buf.WriteRune(')')
        return buf.String()
    },
    // map_item = map_key :! expr!
    "map_item": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        // map_key = expr
        var key = TranspileFirstChild(tree, children["map_key"])
        var value = Transpile(tree, children["expr"])
        return fmt.Sprintf("[%v, %v]", key, value)
    },
    // hash = { } | { hash_item! hash_tail }!
    "hash": func (tree Tree, ptr int) string {
        if tree.Nodes[ptr].Length == 2 {
            return "{}"
        }
        var items = FlatSubTree(tree, ptr, "hash_item", "hash_tail")
        var buf strings.Builder
        buf.WriteRune('{')
        for i, item := range items {
            buf.WriteString(Transpile(tree, item))
            if i != len(items)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune('}')
        return buf.String()
    },
    // hash_item = name :! expr! | :: name!
    "hash_item": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = children["name"]
        var expr, has_expr = children["expr"]
        var buf strings.Builder
        buf.WriteString(Transpile(tree, name))
        buf.WriteString(": ")
        if has_expr {
            buf.WriteString(Transpile(tree, expr))
        } else {
            buf.WriteString(VarLookup(GetTokenContent(tree, name)))
        }
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
