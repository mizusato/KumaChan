package transpiler

import "strings"
import "../syntax"


func TranspileOperationSequence (tree Tree, ptr int) [][]string {
    if tree.Nodes[ptr].Part.Id != syntax.Name2Id["operand_tail"] {
        panic("invalid usage of TranspileOperationSequence()")
    }
    var operations = make([][]string, 0, 20)
    for tree.Nodes[ptr].Length > 0 {
        // operand_tail? = get operand_tail | call operand_tail
        var operation_ptr = tree.Nodes[ptr].Children[0]
        var next_ptr = tree.Nodes[ptr].Children[1]
        var op_node = &tree.Nodes[operation_ptr]
        if op_node.Part.Id == syntax.Name2Id["get"] {
            // get = get_expr | get_name
            var params = Children(tree, op_node.Children[0])
            // get_expr = Get [ expr! ]! nil_flag
            // get_name = Get . name! nil_flag
            var key string
            var _, is_get_expr = params["expr"]
            if is_get_expr {
                key = Transpile(tree, params["expr"])
            } else {
                key = Transpile(tree, params["name"])
            }
            operations = append(operations, []string {
                "g", key, Transpile(tree, params["nil_flag"]),
            })
        } else {
            // call = call_self | call_method
            var child_ptr = op_node.Children[0]
            var params = Children(tree, child_ptr)
            // call_self = Call args
            // call_method = -> name method_args
            var args = TranspileLastChild(tree, child_ptr)
            var name_ptr, is_method_call = params["name"]
            if is_method_call {
                operations = append(operations, []string {
                    "m", Transpile(tree, name_ptr), args,
                })
            } else {
                operations = append(operations, []string {
                    "c", args,
                })
            }
        }
        ptr = next_ptr
    }
    return operations
}


var TransMapByName = map[string]TransFunction {

    "program": TranspileFirstChild,

    "command": TranspileFirstChild,
    "cmd_group1": TranspileFirstChild,
    "cmd_group2": TranspileFirstChild,
    "cmd_group3": TranspileFirstChild,
    "cmd_exec": TranspileChild("expr"),

    "expr": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var operand_ptrs = FlatSubTree(tree, ptr, "operand", "expr_tail")
        var tail = children["expr_tail"]
        var operator_ptrs = FlatSubTree(tree, tail, "operator", "expr_tail")
        var operators = make([]syntax.Operator, len(operator_ptrs))
        for i, operator_ptr := range operator_ptrs {
            operators[i] = GetOperatorInfo(tree, operator_ptr)
        }
        var reduced = ReduceExpression(operators)
        var do_transpile func(int) string
        do_transpile = func (operand_index int) string {
            if operand_index >= 0 {
                return Transpile(tree, operand_ptrs[operand_index])
            } else {
                var reduced_index = -(operand_index) - 1
                var sub_expr [3]int = reduced[reduced_index]
                var operand1 = do_transpile(sub_expr[0])
                var operand2 = do_transpile(sub_expr[1])
                var operator = Transpile(tree, operator_ptrs[sub_expr[2]])
                var operator_info = operators[sub_expr[2]]
                var lazy_eval = operator_info.LazyEval
                var buf strings.Builder
                buf.WriteString(operator)
                buf.WriteRune('(')
                if lazy_eval {
                    buf.WriteString(LazyValueWrapper(operand1))
                } else {
                    buf.WriteString(operand1)
                }
                buf.WriteRune(',')
                if lazy_eval {
                    buf.WriteString(LazyValueWrapper(operand2))
                } else {
                    buf.WriteString(operand2)
                }
                buf.WriteRune(')')
                return buf.String()
            }
        }
        return do_transpile(-len(reduced))
    },
    "operand": func (tree Tree, ptr int) string {
        // operand = unary operand_base operand_tail
        var children = Children(tree, ptr)
        var unary_ptr = children["unary"]
        var base_ptr = children["operand_base"]
        var tail_ptr = children["operand_tail"]
        var base = Transpile(tree, base_ptr)
        var operations = TranspileOperationSequence(tree, tail_ptr)
        var buf strings.Builder
        var reduce func (int)
        reduce = func (i int) {
            if i == -1 {
                buf.WriteString(base)
                return
            }
            var op = operations[i]
            buf.WriteString(op[0])
            buf.WriteRune('(')
            reduce(i-1)
            for j := 1; j < len(op); j++ {
                buf.WriteRune(',')
                buf.WriteString(op[j])
            }
            buf.WriteRune(')')
        }
        var has_unary = NotEmpty(tree, unary_ptr)
        if has_unary {
            buf.WriteString(Transpile(tree, unary_ptr))
            buf.WriteRune('(')
        }
        reduce(len(operations)-1)
        if has_unary {
            buf.WriteRune(')')
        }
        return buf.String()
    },
    "unary": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        var child_node = &tree.Nodes[child_ptr]
        switch name := syntax.Id2Name[child_node.Part.Id]; name {
        case "@not", "~":
            return `o("~")`
        case "-":
            return `o("-")`
        case "!":
            return `o("!")`
        case "@expose":
            return "expose"
        default:
            panic("cannot transpile unknown unary operator " + name)
        }
    },
    "operand_base": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var wrapped_expr, exists = children["expr"]
        if exists {
            return Transpile(tree, wrapped_expr)
        } else {
            return TranspileFirstChild(tree, ptr)
        }
    },
    "operator": func (tree Tree, ptr int) string {
        var info = GetOperatorInfo(tree, ptr)
        var buf strings.Builder
        if info.CanOverload {
            buf.WriteString("o")
            buf.WriteRune('(')
            buf.WriteString(EscapeRawString([]rune(info.Match)))
            buf.WriteRune(')')
            // buf.WriteString(VarLookup([]rune("operator_" + name)))
        } else {
            var name = info.Name
            if name == "is" {
                buf.WriteString("is")
            } else {
                panic("unknown non-overloadable operator: " + name)
            }
        }
        return buf.String()
    },

    "name": func (tree Tree, ptr int) string {
        var node = &tree.Nodes[ptr]
        var child = &tree.Nodes[node.Children[0]]
        if child.Part.Id == syntax.Name2Id["String"] {
            return TransMap[syntax.Name2Id["string"]](tree, ptr)
        } else {
            return EscapeRawString(GetTokenContent(tree, ptr))
        }
    },
    "nil_flag": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return "true"
        } else {
            return "false"
        }
    },

    "method_args": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var extra_ptr, is_only_extra = children["extra_arg"]
        if is_only_extra {
            var buf strings.Builder
            buf.WriteRune('[')
            if tree.Nodes[extra_ptr].Length > 0 {
                buf.WriteString(Transpile(tree, extra_ptr))
            }
            buf.WriteRune(']')
            return buf.String()
        } else {
            var args_ptr, exists = children["args"]
            if !exists { panic("transpiler: expect args in methods_args") }
            return Transpile(tree, args_ptr)
        }
    },
    "args": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var buf strings.Builder
        buf.WriteRune('[')
        var arglist_ptr = children["arglist"]
        var has_arglist = NotEmpty(tree, arglist_ptr)
        if has_arglist {
            buf.WriteString(Transpile(tree, arglist_ptr))
        }
        var extra_ptr = children["extra_arg"]
        var has_extra = NotEmpty(tree, extra_ptr)
        if has_extra {
            if has_arglist {
                buf.WriteRune(',')
            }
            buf.WriteString(Transpile(tree, extra_ptr))
        }
        buf.WriteRune(']')
        return buf.String()
    },
    "extra_arg": TranspileLastChild,
    "arglist": TranspileFirstChild,
    "exprlist": func (tree Tree, ptr int) string {
        var ptrs = FlatSubTree(tree, ptr, "expr", "exprlist_tail")
        var buf strings.Builder
        for i, item_ptr := range ptrs {
            buf.WriteString(Transpile(tree, item_ptr))
            if i < len(ptrs)-1 {
                buf.WriteRune(',')
            }
        }
        return buf.String()
    },

    "identifier": func (tree Tree, ptr int) string {
        return VarLookup(GetTokenContent(tree, ptr))
    },

    "literal": TranspileFirstChild,
    /* Primitive Values */
    "primitive": TranspileFirstChild,
    "string": func (tree Tree, ptr int) string {
        var content = GetTokenContent(tree, ptr)
        var trimed = content[1:len(content)-1]
        return EscapeRawString(trimed)
    },
    "number": func (tree Tree, ptr int) string {
        return string(GetTokenContent(tree, ptr))
    },
    "bool": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        if tree.Nodes[child_ptr].Part.Id == syntax.Name2Id["@true"] {
            return "true"
        } else {
            return "false"
        }
    },

}