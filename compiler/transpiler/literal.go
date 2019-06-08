package transpiler

import "fmt"
import "strings"
import "../syntax"


var LiteralMap = map[string]TransFunction {
    // literal = primitive | adv_literal
    "literal": TranspileFirstChild,
    // adv_literal = comp | type_literal | list | hash | brace_literal
    "adv_literal": TranspileFirstChild,
    // brace_literal = when | iife | struct
    "brace_literal": TranspileFirstChild,
    // when = @when { when_list }!
    "when": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var list = Transpile(tree, children["when_list"])
        return fmt.Sprintf("((function(){ %v })())", list)
    },
    // when_list = when_item when_list_tail
    "when_list": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var item_ptrs = FlatSubTree(tree, ptr, "when_item", "when_list_tail")
        var buf strings.Builder
        buf.WriteString("if (false) { __.v } ")
        for _, item_ptr := range item_ptrs {
            // when_item = expr : expr
            var row, col = GetRowColInfo(tree, item_ptr)
            var condition = TranspileFirstChild(tree, item_ptr)
            var value = TranspileLastChild(tree, item_ptr)
            var condition_bool = fmt.Sprintf (
                "__.c(__.rb, [%v], %v, %v, %v)",
                condition, file, row, col,
            )
            fmt.Fprintf (
                &buf, "else if (%v) { return %v }",
                condition_bool, value,
            )
        }
        fmt.Fprintf (
            &buf, "else { __.c(__.wf, [], %v, %v, %v) }",
            file, row, col,
        )
        return buf.String()
    },
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
    // comp = .[ comp_rule! ]! | [ comp_rule ]!
    "comp": func (tree Tree, ptr int) string {
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
    // struct = type struct_hash
    "struct": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var type_ = Transpile(tree, children["type"])
        var hash = Transpile(tree, children["struct_hash"])
        return fmt.Sprintf (
            "__.c(__.ns, [%v, %v], %v, %v, %v)",
            type_, hash, file, row, col,
        )
    },
    // struct_hash = { struct_hash_item struct_hash_tail }!
    "struct_hash": func (tree Tree, ptr int) string {
        var item_ptrs = FlatSubTree (
            tree, ptr, "struct_hash_item", "struct_hash_tail",
        )
        var names = make(map[string]bool)
        var buf strings.Builder
        buf.WriteString("{ ")
        for i, item_ptr := range item_ptrs {
            // struct_hash_item = name : expr! | :: name!
            var children = Children(tree, item_ptr)
            var name_ptr = children["name"]
            var name = Transpile(tree, name_ptr)
            var expr_ptr, has_expr = children["expr"]
            var _, exists = names[name]
            if exists {
                panic("duplicate Structure field " + name)
            }
            names[name] = true
            var expr string
            if has_expr {
                expr = Transpile(tree, expr_ptr)
            } else {
                expr = VarLookup(GetTokenContent(tree, name_ptr))
            }
            fmt.Fprintf(&buf, "%v: %v", name, expr)
            if i != len(item_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteString(" }")
        return buf.String()
    },
    // primitive = string | number | bool
    "primitive": TranspileFirstChild,
    // string = String
    "string": func (tree Tree, ptr int) string {
        var MulId = syntax.Name2Id["MulStr"]
        var child = tree.Nodes[tree.Nodes[ptr].Children[0]]
        var content = GetTokenContent(tree, ptr)
        var trimed []rune
        if child.Part.Id == MulId {
            trimed = content[3:len(content)-3]
        } else {
            trimed = content[1:len(content)-1]
        }
        return EscapeRawString(trimed)
    },
    // number = Hex | Exp | Dec | Int
    "number": func (tree Tree, ptr int) string {
        return string(GetTokenContent(tree, ptr))
    },
    // bool = @true | @false
    "bool": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        if tree.Nodes[child_ptr].Part.Id == syntax.Name2Id["@true"] {
            return "true"
        } else {
            return "false"
        }
    },
}
