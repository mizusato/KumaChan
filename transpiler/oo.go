package transpiler

import "fmt"
import "strings"


var OO_Map = map[string]TransFunction {
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
        var pfs = Transpile(tree, children["pfs"])
        var methods = Transpile(tree, children["methods"])
        var options = Transpile(tree, children["class_opt"])
        var def_point = fmt.Sprintf (
            "{ file: %v, row: %v, col: %v }",
            file, row, col,
        )
        var class = fmt.Sprintf (
            "%v(%v, [%v, %v, %v, %v, %v, %v, %v], %v, %v, %v)",
            G(CALL), G(C_CLASS),
            name, impls, init, pfs, methods, options, def_point,
            file, row, col,
        )
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, class)
        } else {
            value = class
        }
        return fmt.Sprintf (
            "%v(%v, [%v, %v, true, %v], %v, %v, %v)",
            G(CALL), L_VAR_DECL, name, value, G(T_TYPE),
            file, row, col,
        )
    },
    // supers? = @is typelist
    "supers": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) {
            return "[]"
        } else {
            return TranspileLastChild(tree, ptr)
        }
    },
    // init = @init Call paralist_strict! body! creators
    "init": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var class_ptr = tree.Nodes[ptr].Parent  // note: access of parent node
        var class_children = Children(tree, class_ptr)
        var name = GetTokenContent(tree, class_children["name"])
        var main = InitFunction(tree, ptr, name)
        var alternatives = Transpile(tree, children["creators"])
        return fmt.Sprintf("[%v, %v]", main, alternatives)
    },
    // creators? = creator creators
    "creators": func (tree Tree, ptr int) string {
        var init_ptr = tree.Nodes[ptr].Parent  // note: access of parent node
        var class_ptr = tree.Nodes[init_ptr].Parent
        var class_children = Children(tree, class_ptr)
        var name = GetTokenContent(tree, class_children["name"])
        var creator_ptrs = FlatSubTree(tree, ptr, "creator", "creators")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, creator_ptr := range creator_ptrs {
            // creator = @create Call paralist_strict! body!
            buf.WriteString(InitFunction(tree, creator_ptr, name))
            if i != len(creator_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
    },
    // pfs? = pf pfs
    "pfs": func (tree Tree, ptr int) string {
        return MethodTable(tree, ptr, "pf", "pfs")
    },
    // methods? = method methods
    "methods": func (tree Tree, ptr int) string {
        return MethodTable(tree, ptr, "method", "methods")
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
    // interface = @interface name generic_params { members }
    "interface": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var gp_ptr = children["generic_params"]
        var name = Transpile(tree, name_ptr)
        var members = Transpile(tree, children["members"])
        var def_point = fmt.Sprintf (
            "{ file: %v, row: %v, col: %v }",
            file, row, col,
        )
        var interface_ = fmt.Sprintf (
            "%v(%v, [%v, %v, %v], %v, %v, %v)",
            G(CALL), G(C_INTERFACE), name, members, def_point,
            file, row, col,
        )
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, interface_)
        } else {
            value = interface_
        }
        return fmt.Sprintf (
            "%v(%v, [%v, %v, true, %v], %v, %v, %v)",
            G(CALL), L_VAR_DECL, name, value, G(T_TYPE), file, row, col,
        )
    },
    // members? = member members
    "members": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "[]" }
        return TranspileSubTree(tree, ptr, "member", "members")
    },
    // member = method_implemented | method_blank
    "member": TranspileFirstChild,
    // method_implemented = name Call paralist_strict! ->! type! body
    "method_implemented": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var method = TransMapByName["f_sync"](tree, ptr)
        return fmt.Sprintf (
            "{ name: %v, f: %v }",
            name, method,
        )
    },
    // method_blank = name Call paralist_strict! -> type
    "method_blank": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var parameters = Transpile(tree, children["paralist_strict"])
        var value_type = Transpile(tree, children["type"])
        var proto = fmt.Sprintf (
            "{ parameters: %v, value_type: %v }",
            parameters, value_type,
        )
        return fmt.Sprintf (
            "{ name: %v, f: %v }",
            name, proto,
        )
    },
}
