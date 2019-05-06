package transpiler

import "fmt"
import "strings"
import "../syntax"


var Rules = []map[string]TransFunction {
    Expressions, Containers, Functions, CommandsMap,
}


var TransMapByName = map[string]TransFunction {

    // program = module | eval
    "program": TranspileFirstChild,

    // eval = command
    "eval": func (tree Tree, ptr int) string {
        var command = TranspileFirstChild(tree, ptr)
        var command_ptr = tree.Nodes[ptr].Children[0]
        var group_ptr = tree.Nodes[command_ptr].Children[0]
        var real_cmd_ptr = tree.Nodes[group_ptr].Children[0]
        var FlowId = syntax.Name2Id["cmd_flow"]
        var body string
        if tree.Nodes[real_cmd_ptr].Part.Id == FlowId {
            body = fmt.Sprintf("%v; return __.v;", command)
        } else {
            body = fmt.Sprintf("return %v", command)
        }
        var buf strings.Builder
        buf.WriteString(BareFunction(body))
        fmt.Fprintf(&buf, "(%v.scope.Eval)", Runtime)
        return buf.String()
    },

    // command = cmd_group1 | cmd_group2 | cmd_group3
    "command": TranspileFirstChild,
    "cmd_group1": TranspileFirstChild,
    "cmd_group2": TranspileFirstChild,
    "cmd_group3": TranspileFirstChild,

    // literal = primitive | adv_literal
    "literal": TranspileFirstChild,
    // adv_literal = xml | comprehension | abs_literal | map | list | hash
    "adv_literal": TranspileFirstChild,

    /* Trivial Things */
    "name": func (tree Tree, ptr int) string {
        return EscapeRawString(GetTokenContent(tree, ptr))
    },
    "namelist": func (tree Tree, ptr int) string {
        var names = FlatSubTree(tree, ptr, "name", "namelist_tail")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, name := range names {
            buf.WriteString(Transpile(tree, name))
            if i != len(names)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
    },
    "identifier": func (tree Tree, ptr int) string {
        // depended by CommandsMap["reset"]
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
