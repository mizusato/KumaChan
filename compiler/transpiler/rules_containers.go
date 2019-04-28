package transpiler

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
        var buf strings.Builder
        buf.WriteRune('[')
        buf.WriteString(key)
        buf.WriteString(", ")
        buf.WriteString(value)
        buf.WriteRune(']')
        return buf.String()
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
}
