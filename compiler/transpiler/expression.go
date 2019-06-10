package transpiler

import "fmt"
import "strings"
import "../syntax"


func TranspileOperationSequence (tree Tree, ptr int) [][]string {
    if tree.Nodes[ptr].Part.Id != syntax.Name2Id["operand_tail"] {
        panic("invalid usage of TranspileOperationSequence()")
    }
    var file = GetFileName(tree)
    var operations = make([][]string, 0, 20)
    for NotEmpty(tree, ptr) {
        // operand_tail? = get operand_tail | call operand_tail
        var operation_ptr = tree.Nodes[ptr].Children[0]
        var next_ptr = tree.Nodes[ptr].Children[1]
        var op_node = &tree.Nodes[operation_ptr]
        var row, col = GetRowColInfo(tree, operation_ptr)
        if op_node.Part.Id == syntax.Name2Id["get"] {
            // get_expr = Get [ expr! ]! nil_flag
            // get_name = Get . name! nil_flag
            var key, nil_flag = GetKey(tree, operation_ptr)
            operations = append(operations, []string {
                "__.g", key, nil_flag,
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
                    "__.c", args, file, row, col,
                })
            }
        }
        ptr = next_ptr
    }
    return operations
}


var ExpressionMap = map[string]TransFunction {
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
            var lazy_eval = operator_info.Lazy
            var row, col = GetRowColInfo(tree, operator_ptr)
            var real_operand2 string
            if lazy_eval {
                real_operand2 = LazyValueWrapper(operand2)
            } else {
                real_operand2 = operand2
            }
            return fmt.Sprintf(
                "__.c(%v, [%v, %v], %v, %v, %v)",
                operator, operand1, real_operand2,
                file, row, col,
            )
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
            buf.WriteString("__.c")
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
    // unary? = unary_group1 | unary_group2 | unary_group3
    "unary": func (tree Tree, ptr int) string {
        var group_ptr = tree.Nodes[ptr].Children[0]
        var real_child_ptr = tree.Nodes[group_ptr].Children[0]
        var child_node = &tree.Nodes[real_child_ptr]
        var name = syntax.Id2Name[child_node.Part.Id]
        if name == "@expose" {
            return "expose"
        } else if name == "-" {
            return `__.o("negate")`
        } else {
            var t = strings.TrimPrefix(name, "@")
            return fmt.Sprintf("__.o(%v)", EscapeRawString([]rune(t)))
        }
    },
    // operand_base = wrapped | lambda | literal | dot_para | identifier
    "operand_base": TranspileFirstChild,
    // wrapped = ( expr! )!
    "wrapped": TranspileChild("expr"),
    // dot_para = . Name
    "dot_para": func (tree Tree, ptr int) string {
        var last_ptr = tree.Nodes[ptr].Children[tree.Nodes[ptr].Length-1]
        return VarLookup(GetTokenContent(tree, last_ptr))
    },
    // identifier = Name
    "identifier": func (tree Tree, ptr int) string {
        // depended by Command["reset"]
        var content = GetTokenContent(tree, ptr)
        var content_string = string(content)
        if strings.HasPrefix(content_string, "__") && content_string != "__" {
            // ECMAScript Identifier
            var trimed = strings.TrimPrefix(content_string, "__")
            return fmt.Sprintf("(%v)", trimed)
        } else {
            return VarLookup(GetTokenContent(tree, ptr))
        }
    },
    // operator = op_group1 | op_compare | op_logic | op_arith
    "operator": func (tree Tree, ptr int) string {
        // depended by CommandsMap["set_op"]
        var info = GetOperatorInfo(tree, ptr)
        var buf strings.Builder
        buf.WriteString("__.o")
        buf.WriteRune('(')
        var name = strings.TrimPrefix(info.Match, "@")
        buf.WriteString(EscapeRawString([]rune(name)))
        buf.WriteRune(')')
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
        var typelist, exists = children["type_arglist"]
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
