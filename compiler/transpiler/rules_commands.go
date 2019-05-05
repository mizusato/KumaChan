package transpiler

import "fmt"
// import "strings"


var CommandsMap = map[string]TransFunction {
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
    // cmd_scope = cmd_let | cmd_var | cmd_reset
    "cmd_scope": TranspileFirstChild,
    // cmd_let = @let name var_type = expr
    "cmd_let": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var T = Transpile(tree, children["var_type"])
        var value = Transpile(tree, children["expr"])
        return fmt.Sprintf("dl(%v, %v, true, %v)", name, value, T)
    },
    // cmd_var = @var name var_type = expr
    "cmd_var": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var T = Transpile(tree, children["var_type"])
        var value = Transpile(tree, children["expr"])
        return fmt.Sprintf("dl(%v, %v, false, %v)", name, value, T)
    },
    // cmd_reset = @reset name = expr
    "cmd_reset": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var value = Transpile(tree, children["expr"])
        return fmt.Sprintf("rt(%v, %v)", name, value)
    },
    // var_type? = : type
    "var_type": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return TranspileLastChild(tree, ptr)
        } else {
            return "a"   // Types.Any
        }
    },
}
