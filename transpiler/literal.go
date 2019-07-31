package transpiler

import "fmt"
import "strings"
import "../parser"
import "../parser/syntax"


var LiteralMap = map[string]TransFunction {
    // literal = primitive | adv_literal
    "literal": TranspileFirstChild,
    // adv_literal = comp | type_literal | list | hash | brace
    "adv_literal": TranspileFirstChild,
    // brace = when | iife | struct
    "brace": TranspileFirstChild,
    // when = @when {! branch_list }!
    "when": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        return fmt.Sprintf (
            "((function(){ %v })())",
            BranchList (
                tree, children["branch_list"],
                CondBranch, G(WHEN_FAILED),
            ),
        )
    },
    // match = @match expr {! branch_list }!
    "match": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var declare_target = fmt.Sprintf (
            "let %v = %v;",
            MATCH_TARGET, Transpile(tree, children["expr"]),
        )
        return fmt.Sprintf (
            "((function(){ %v %v })())",
            declare_target,
            BranchList (
                tree, children["branch_list"],
                MatchBranch, G(MATCH_FAILED),
            ),
        )
    },
    // observer = @observer body
    "observer": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("observer"), []rune("()"), []rune("Object"))
        var f = Function(tree, body_ptr, F_Sync, desc, "[]", G(T_ANY))
        return fmt.Sprintf("%v(%v)", G(C_OBSERVER), f)
    },
    // tree = _tree tree_node! | @tree { tree_node! }!
    "tree": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        return Transpile(tree, children["tree_node"])
    },
    // tree_node = type { node_props node_children }!
    "tree_node": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var inflater = Transpile(tree, children["type"])
        var node_props = Transpile(tree, children["node_props"])
        var node_children = Transpile(tree, children["node_children"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v, %v, %v], %v, %v, %v)",
            G(CALL), G(C_TREE_NODE), inflater, node_props, node_children,
            file, row, col,
        )
    },
    // node_props? = node_prop node_props
    "node_props": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "{}" }
        var prop_ptrs = FlatSubTree(tree, ptr, "node_prop", "node_props")
        var occurred = make(map[string]bool)
        var buf strings.Builder
        buf.WriteString("{ ")
        for i, prop_ptr := range prop_ptrs {
            // node_prop = name = expr! | string = expr!
            var children = Children(tree, prop_ptr)
            var name_ptr, has_name = children["name"]
            var name string
            // note: see comments in LiteralMap["hash"]
            if has_name {
                name = Transpile(tree, name_ptr)
            } else {
                name = Transpile(tree, children["string"])
            }
            if occurred[name] {
                parser.Error (
                    tree, prop_ptr,
                    fmt.Sprintf (
                        "duplicate prop %v",
                        name,
                    ),
                )
            }
            occurred[name] = true
            var value = Transpile(tree, children["expr"])
            fmt.Fprintf(&buf, "%v: %v", name, value)
            if i != len(prop_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteString(" }")
        return buf.String()
    },
    // node_children? = node_child node_children
    "node_children": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "[]" }
        return TranspileSubTree(tree, ptr, "node_child", "node_children")
    },
    // node_child = tree_node | = expr
    "node_child": TranspileLastChild,
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
                parser.Error (
                    tree, item,
                    fmt.Sprintf (
                        "duplicate hash key %v",
                        name,
                    ),
                )
            }
            names[name] = true
            var expr string
            if has_expr {
                expr = Transpile(tree, expr_ptr)
            } else {
                expr = VarLookup(tree, name_ptr)
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
        } else {
            return TranspileSubTree(tree, ptr, "list_item", "list_tail")
        }
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
            f = G(ITER_COMP)
        } else {
            f = G(LIST_COMP)
        }
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), f, rule, file, row, col,
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
        var occurred = make(map[string]bool)
        var iterators = make([]string, 0, 10)
        for _, item_ptr := range item_ptrs {
            // in_item = name @in expr
            var item_children = Children(tree, item_ptr)
            var name = string(GetTokenContent(tree, item_children["name"]))
            var expr = Transpile(tree, item_children["expr"])
            if occurred[name] {
                parser.Error (
                    tree, item_ptr, fmt.Sprintf (
                        `duplicate parameter "%v"`,
                        name,
                    ),
                )
            }
            occurred[name] = true
            names = append(names, name)
            iterators = append(iterators, expr)
        }
        var parameters = UntypedParameters(names)
        var proto = fmt.Sprintf (
            "{ parameters: %v, value_type: %v }",
            parameters, G(T_ANY),
        )
        var val_raw = BareFunction(fmt.Sprintf("return %v;", val_expr))
        var val_desc = EscapeRawString([]rune("comprehension.value_function"))
        var val = fmt.Sprintf (
            "%v(%v, %v, %v, %v)",
            L_WRAP, proto, "null", val_desc, val_raw,
        )
        var iterator_list = fmt.Sprintf (
            "[%v]", strings.Join(iterators, ", "),
        )
        var filter_expr = Transpile(tree, children["opt_filters"])
        var filter_raw = BareFunction(fmt.Sprintf("return %v;", filter_expr))
        var filter_desc = EscapeRawString([]rune("comprehension.filter"))
        var filter = fmt.Sprintf (
            "%v(%v, %v, %v, %v)",
            L_WRAP, proto, "null", filter_desc, filter_raw,
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
            "%v(%v, [%v, %v], %v, %v, %v)",
            G(CALL), G(C_STRUCT), type_, hash, file, row, col,
        )
    },
    // struct_hash = { } | { struct_hash_item struct_hash_tail }!
    "struct_hash": func (tree Tree, ptr int) string {
        if tree.Nodes[ptr].Length == 2 {
            return "{}"
        }
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
                parser.Error (
                    tree, item_ptr, fmt.Sprintf (
                        "duplicate struct field %v",
                        name,
                    ),
                )
            }
            names[name] = true
            var expr string
            if has_expr {
                expr = Transpile(tree, expr_ptr)
            } else {
                expr = VarLookup(tree, name_ptr)
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
