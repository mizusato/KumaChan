package transpiler

import "fmt"
import "strings"


var OO = map[string]TransFunction {
    // class = @class name generic_params supers { init methods class_opt }
    "class": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var gp_ptr = children["generic_params"]
        var name = Transpile(tree, name_ptr)
        var impls = Transpile(tree, children["supers"])
        var init = Transpile(tree, children["init"])
        var methods = Transpile(tree, children["methods"])
        var options = Transpile(tree, children["class_opt"])
        var def_point = fmt.Sprintf (
            "{ file: %vrow: %v, col: %v }",
            file, row, col,
        )
        var class = fmt.Sprintf (
            "__.cc(%v, %v, %v, %v, %v, %v)",
            name, impls, init, methods, options, def_point,
        )
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, class)
        } else {
            value = class
        }
        return fmt.Sprintf (
            "__.c(dl, [%v, %v, true, __.t], %v, %v, %v)",
            name, value, file, row, col,
        )
    },
    // supers? = @is exprlist
    "supers": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            return fmt.Sprintf("[%v]", Transpile(tree, children["exprlist"]))
        } else {
            return "[]"
        }
    },
    // init = @init Call paralist_strict! body!
    "init": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var type_ptr = children["type"]
        var value_type = Transpile(tree, type_ptr)
        var body_ptr = children["body"]
        var class_ptr = tree.Nodes[ptr].Parent
        var class_children = Children(tree, class_ptr)
        var name_ptr = class_children["name"]
        var class_name = GetTokenContent(tree, name_ptr)
        var desc = Desc (
            append(class_name, []rune(".Initializer")...),
            GetWholeContent(tree, params_ptr),
            GetWholeContent(tree, type_ptr),
        )
        return Function (
            tree, body_ptr, F_Sync,
            desc, parameters, value_type,
        )
    },
    // methods? = method methods
    "methods": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var method_ptrs = FlatSubTree(tree, ptr, "method", "methods")
            var buf strings.Builder
            for i, method_ptr := range method_ptrs {
                var children = Children(tree, method_ptr)
                var name = Transpile(tree, children["name"])
                var method = Transpile(tree, method_ptr)
                fmt.Fprintf(&buf, "{ name: %v, f: %v }", name, method)
                if i != len(method_ptrs)-1 {
                    buf.WriteString(", ")
                }
            }
            return fmt.Sprintf("[ %v ]", buf.String())
        } else {
            return "[]"
        }
    },
    // method = name Call paralist_strict! ->! type! body!
    "method": func (tree Tree, ptr int) string {
        return Functions["f_sync"](tree, ptr)
    },
    // class_opt = operator_defs data
    "class_opt": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var ops = Transpile(tree, children["operator_defs"])
        var data = Transpile(tree, children["data"])
        return fmt.Sprintf("{ ops: %v, data: %v }", ops, data)
    },
    // data? = @data hash
    "data": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return TranspileLastChild(tree, ptr)
        } else {
            return "{}"
        }
    },
}
