package transpiler

import "fmt"
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
        buf.WriteString(".scope.Eval")
        buf.WriteRune(')')
        return buf.String()
    },

    // command = cmd_group1 | cmd_group2 | cmd_group3
    "command": TranspileFirstChild,
    "cmd_group1": TranspileFirstChild,
    "cmd_group2": TranspileFirstChild,
    "cmd_group3": TranspileFirstChild,
    // cmd_def = function | abs_def
    "cmd_def": TranspileFirstChild,
    // cmd_exec = expr
    "cmd_exec": TranspileFirstChild,
    // cmd_return = @return Void | @return expr
    "cmd_return": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr, exists = children["expr"]
        if exists {
            return fmt.Sprintf("return %v", Transpile(tree, expr))
        } else {
            return "return v"
        }
    },

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
        var children = Children(tree, ptr)
        var _, is_es = children["EsId"]
        var content = GetTokenContent(tree, ptr)
        if is_es {
            return string(content[2:])
        } else {
            return VarLookup(content)
        }
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
