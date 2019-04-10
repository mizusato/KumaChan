package transpiler


var TransMapByName = map[string]TransFunction {
    "program": TranspileFirstChild,

    "command": TranspileFirstChild,
    "cmd_group1": TranspileFirstChild,
    "cmd_group2": TranspileFirstChild,
    "cmd_group3": TranspileFirstChild,
    "cmd_exec": TranspileChild("expr"),
    "expr": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var operand1 = children["operand"]
        return Transpile(tree, operand1)
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
    "literal": TranspileFirstChild,
    "primitive": TranspileFirstChild,
    "string": func (tree Tree, ptr int) string {
        var token_pos = tree.Nodes[ptr].Pos
        var token = &tree.Tokens[token_pos]
        var content = token.Content
        var trimed = content[1:len(content)-1]
        return EscapeRawString(trimed)
    },
}
