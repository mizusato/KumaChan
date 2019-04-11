package transpiler

import "strings"
import "../syntax"


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
        var children = Children(tree, ptr)
        var base = children["operand_base"]
        return Transpile(tree, base)
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
                buf.WriteString(Runtime)
                buf.WriteString("is")
            } else {
                panic("unknown non-overloadable operator: " + name)
            }
        }
        return buf.String()
    },

    "identifier": func (tree Tree, ptr int) string {
        // node.pos is also token.pos because identifier = Name
        var token_pos = tree.Nodes[ptr].Pos
        var token = &tree.Tokens[token_pos]
        return VarLookup(token.Content)
    },

    "literal": TranspileFirstChild,
    /* Primitive Values */
    "primitive": TranspileFirstChild,
    "string": func (tree Tree, ptr int) string {
        // node.pos is also token.pos because string = String
        var token_pos = tree.Nodes[ptr].Pos
        var token = &tree.Tokens[token_pos]
        var content = token.Content
        var trimed = content[1:len(content)-1]
        return EscapeRawString(trimed)
    },
    "number": func (tree Tree, ptr int) string {
        // node.pos is also token.pos because number = Hex | Exp | Dec | Int
        var token_pos = tree.Nodes[ptr].Pos
        var token = &tree.Tokens[token_pos]
        return string(token.Content)
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
