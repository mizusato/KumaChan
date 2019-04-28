package transpiler

import "strings"
import "../syntax"


var TransMapByName = map[string]TransFunction {

    // program = module | eval
    "program": TranspileFirstChild,

    // eval = command
    "eval": func (tree Tree, ptr int) string {
        var command = TranspileFirstChild(tree, ptr)
        var buf strings.Builder
        buf.WriteString(BareFunction("return " + command))
        buf.WriteRune('(')
        buf.WriteString(Runtime)
        buf.WriteString("scope.Eval")
        buf.WriteRune(')')
        return buf.String()
    },

    // commands? = command commands
    "commands": func (tree Tree, ptr int) string {
        var commands = FlatSubTree(tree, ptr, "command", "commands")
        var ReturnId = syntax.Name2Id["cmd_return"]
        var has_return = false
        var buf strings.Builder
        for _, command := range commands {
            if !has_return && tree.Nodes[command].Part.Id == ReturnId {
                has_return = true
            }
            buf.WriteString(Transpile(tree, command))
            buf.WriteString("; ")
        }
        if !has_return {
            // return Void
            buf.WriteString("return v;")
        }
        return buf.String()
    },
    // command = cmd_group1 | cmd_group2 | cmd_group3
    "command": TranspileFirstChild,
    "cmd_group1": TranspileFirstChild,
    "cmd_group2": TranspileFirstChild,
    "cmd_group3": TranspileFirstChild,
    // cmd_exec = expr
    "cmd_exec": TranspileFirstChild,

    /* Rules About Expressions */
    "expr": Expr["expr"],
    "operand": Expr["operand"],
    "unary": Expr["operand_unary"],
    "operand_base": Expr["operand_base"],
    "operator": Expr["operator"],
    "nil_flag": Expr["nil_flag"],
    "method_args": Expr["method_args"],
    "args": Expr["args"],
    "extra_arg": Expr["extra_arg"],
    "arglist": Expr["arglist"],
    "exprlist": Expr["exprlist"],

    // literal = primitive | adv_literal
    "literal": TranspileFirstChild,
    // adv_literal = xml | comprehension | abs_literal | map | list | hash
    "adv_literal": TranspileFirstChild,

    /* Rules About Containers */
    "map": Containers["map"],
    "map_item": Containers["map_item"],
    "hash": Containers["hash"],
    "hash_item": Containers["hash_item"],
    "list": Containers["list"],
    "list_item": Containers["list_item"],

    /* Trivial Things */
    "name": func (tree Tree, ptr int) string {
        return EscapeRawString(GetTokenContent(tree, ptr))
    },
    "identifier": func (tree Tree, ptr int) string {
        return VarLookup(GetTokenContent(tree, ptr))
    },
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
