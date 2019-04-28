package transpiler

import "strings"
import "../syntax"


func TranspileOperationSequence (tree Tree, ptr int) [][]string {
    if tree.Nodes[ptr].Part.Id != syntax.Name2Id["operand_tail"] {
        panic("invalid usage of TranspileOperationSequence()")
    }
    var file = GetFileName(tree)
    var operations = make([][]string, 0, 20)
    for tree.Nodes[ptr].Length > 0 {
        // operand_tail? = get operand_tail | call operand_tail
        var operation_ptr = tree.Nodes[ptr].Children[0]
        var next_ptr = tree.Nodes[ptr].Children[1]
        var op_node = &tree.Nodes[operation_ptr]
        var row, col = GetRowColInfo(tree, operation_ptr)
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
                file, row, col,
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
                    file, row, col,
                })
            } else {
                operations = append(operations, []string {
                    "c", args, file, row, col,
                })
            }
        }
        ptr = next_ptr
    }
    return operations
}


var Expressions = map[string]TransFunction {
    // expr = operand expr_tail
    "expr": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
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
            }
            var reduced_index = -(operand_index) - 1
            var sub_expr [3]int = reduced[reduced_index]
            var operand1 = do_transpile(sub_expr[0])
            var operand2 = do_transpile(sub_expr[1])
            var operator_ptr = operator_ptrs[sub_expr[2]]
            var operator = Transpile(tree, operator_ptr)
            var operator_info = operators[sub_expr[2]]
            var lazy_eval = operator_info.LazyEval
            var row, col = GetRowColInfo(tree, operator_ptr)
            var buf strings.Builder
            buf.WriteString("c")
            buf.WriteRune('(')
            buf.WriteString(operator)
            buf.WriteString(", ")
            buf.WriteRune('[')
            buf.WriteString(operand1)
            buf.WriteString(", ")
            if lazy_eval {
                buf.WriteString(LazyValueWrapper(operand2))
            } else {
                buf.WriteString(operand2)
            }
            buf.WriteRune(']')
            buf.WriteString(", ")
            WriteList(&buf, []string {
                file, row, col,
            })
            buf.WriteRune(')')
            return buf.String()
        }
        return do_transpile(-len(reduced))
    },
    // operand = unary operand_base operand_tail
    "operand": func (tree Tree, ptr int) string {
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
                buf.WriteString(", ")
                buf.WriteString(op[j])
            }
            buf.WriteRune(')')
        }
        var has_unary = NotEmpty(tree, unary_ptr)
        if has_unary {
            var unary = Transpile(tree, unary_ptr)
            buf.WriteString("c")
            buf.WriteRune('(')
            buf.WriteString(unary)
            buf.WriteString(", ")
            buf.WriteRune('[')
        }
        reduce(len(operations)-1)
        if has_unary {
            buf.WriteRune(']')
            var file = GetFileName(tree)
            var row, col = GetRowColInfo(tree, unary_ptr)
            buf.WriteString(", ")
            WriteList(&buf, []string {
                file, row, col,
            })
            buf.WriteRune(')')
        }
        return buf.String()
    },
    // unary? = @not opt_call | - | _exc | ~ | @expose opt_call
    "unary": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        var child_node = &tree.Nodes[child_ptr]
        switch name := syntax.Id2Name[child_node.Part.Id]; name {
        case "@not":
            return `o("not")`
        case "~":
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
    // operand_base = wrapped | lambda | literal | dot_para | identifier
    "operand_base": TranspileFirstChild,
    "wrapped": TranspileChild("expr"),
    // operator = op_group1 | op_compare | op_logic | op_arith
    "operator": func (tree Tree, ptr int) string {
        var info = GetOperatorInfo(tree, ptr)
        var buf strings.Builder
        if info.CanOverload {
            buf.WriteString("o")
            buf.WriteRune('(')
            var name = strings.TrimPrefix(info.Match, "@")
            buf.WriteString(EscapeRawString([]rune(name)))
            buf.WriteRune(')')
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
    // nil_flag? = ?
    "nil_flag": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return "true"
        } else {
            return "false"
        }
    },
    // method_args = Call args | extra_arg
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
    // args = ( arglist )! extra_arg | < typelist >
    "args": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var typelist, exists = children["typelist"]
        if exists { return Transpile(tree, typelist) }
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
                buf.WriteString(", ")
            }
            buf.WriteString(Transpile(tree, extra_ptr))
        }
        buf.WriteRune(']')
        return buf.String()
    },
    // extra_arg? = -> lambda | -> adv_literal
    "extra_arg": TranspileLastChild,
    // arglist? = exprlist
    "arglist": TranspileFirstChild,
    // exprlist = expr exprlist_tail
    "exprlist": func (tree Tree, ptr int) string {
        var ptrs = FlatSubTree(tree, ptr, "expr", "exprlist_tail")
        var buf strings.Builder
        for i, item_ptr := range ptrs {
            buf.WriteString(Transpile(tree, item_ptr))
            if i < len(ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        return buf.String()
    },
}
