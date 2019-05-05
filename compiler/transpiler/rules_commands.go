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
    // cmd_set = @set left_val = expr
    "cmd_set": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["expr"])
        // left_val = operand_base gets!
        var left_ptr = children["left_val"]
        var left = Children(tree, left_ptr)
        var base = Transpile(tree, left["operand_base"])
        var gets = FlatSubTree(tree, left["gets"], "get", "gets")
        var real_gets = gets[0:len(gets)-1]
        var tail = gets[len(gets)-1]
        var t = base
        for _, get := range real_gets {
            // get_expr = Get [ expr! ]! nil_flag
            // get_name = Get . name! nil_flag
            var row, col = GetRowColInfo(tree, get)
            var key, nil_flag = GetKey(tree, get)
            t = fmt.Sprintf(
                "g(%v, %v, %v, %v, %v, %v)",
                t, key, nil_flag, file, row, col,
            )
        }
        var object = t
        var row, col = GetRowColInfo(tree, ptr)
        var set_key, _ = GetKey(tree, tail)
        return fmt.Sprintf(
            "c(s, [%v, %v, %v], %v, %v, %v)",
            object, set_key, value, file, row, col,
        )
    },
}
