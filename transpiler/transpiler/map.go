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
        var operands = FlatSubTree(tree, ptr, "operand", "expr_tail")
        var tail = children["expr_tail"]
        var operators = FlatSubTree(tree, tail, "operator", "expr_tail")
        var output strings.Builder
        for i, operand := range operands {
            output.WriteString(Transpile(tree, operand))
            if i < len(operators) {
                output.WriteString(" (")
                output.WriteString(Transpile(tree, operators[i]))
                output.WriteString(") ")
            }
        }
        return output.String()
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
        var group_ptr = tree.Nodes[ptr].Children[0]
        var group = &tree.Nodes[group_ptr]
        var token_node = &tree.Nodes[group.Children[0]]
        var op_id = token_node.Part.Id
        return syntax.Id2Operator[op_id].Name
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
