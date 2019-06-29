package transpiler

import "os"
import "fmt"
import "strings"
import "path/filepath"
import "../parser/syntax"


var RootMap = map[string]TransFunction {
    // module = @module name! export includes commands
    "module": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var module_name = Transpile(tree, children["name"])
        var export_names = Transpile(tree, children["export"])
        var includes = Transpile(tree, children["includes"])
        var commands = Commands(tree, children["commands"], false)
        var content = fmt.Sprintf("%v %v", includes, commands)
        var init = BareFunction(content)
        return fmt.Sprintf (
            "%v.register_module(%v, %v, %v)",
            Runtime, module_name, export_names, init,
        )
    },
    // export? = @export namelist!
    "export": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return TranspileLastChild(tree, ptr)
        } else {
            return "[]"
        }
    },
    // name = Name
    "name": func (tree Tree, ptr int) string {
        return EscapeRawString(GetTokenContent(tree, ptr))
    },
    // namelist = name namelist_tail
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
    // includes? = include includes
    "includes": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "" }
        var inc_ptrs = FlatSubTree(tree, ptr, "include", "includes")
        var buf strings.Builder
        for i, inc_ptr := range inc_ptrs {
            buf.WriteString(Transpile(tree, inc_ptr))
            if i != len(inc_ptrs)-1 {
                buf.WriteRune(' ')
            }
        }
        return buf.String()
    },
    // include = @include string
    "include": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var s = string(GetTokenContent(tree, children["string"]))
        var raw_path = strings.Trim(s, `'"`)
        var path = filepath.Dir(tree.File) + "/" + raw_path
        var f, err = os.Open(path)
        if err != nil { panic(fmt.Sprintf("error: %v: %v", path, err)) }
        defer f.Close()
        return TranspileFile(f, path, "included")
    },
    // included = includes commands
    "included": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var includes = Transpile(tree, children["includes"])
        var commands = Commands(tree, children["commands"], false)
        return fmt.Sprintf("%v %v", includes, commands)
    },
    // eval = commands
    "eval": func (tree Tree, ptr int) string {
        var cmds_ptr = tree.Nodes[ptr].Children[0]
        var cmd_ptrs = FlatSubTree(tree, cmds_ptr, "command", "commands")
        if len(cmd_ptrs) == 0 {
            return fmt.Sprintf("%v.Void", Runtime)
        }
        var buf strings.Builder
        buf.WriteRune('(')
        for i, command_ptr := range cmd_ptrs {
            var command = TranspileFirstChild(tree, command_ptr)
            var group_ptr = tree.Nodes[command_ptr].Children[0]
            var concrete_cmd_ptr = tree.Nodes[group_ptr].Children[0]
            var FlowId = syntax.Name2Id["cmd_flow"]
            var ErrId = syntax.Name2Id["cmd_err"]
            var concrete = tree.Nodes[concrete_cmd_ptr].Part.Id
            var body string
            if concrete == FlowId || concrete == ErrId {
                body = fmt.Sprintf("%v; return __.v;", command)
            } else {
                body = fmt.Sprintf("return %v", command)
            }
            var prepend = "let e = null; "
            buf.WriteString(BareFunction(prepend + body))
            fmt.Fprintf(&buf, "(%v.Eval)", Runtime)
            if i != len(cmd_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(')')
        return buf.String()
    },
}
