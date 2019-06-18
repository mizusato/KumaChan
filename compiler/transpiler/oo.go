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
            "__.c(__.cc, [%v, %v, %v, %v, %v, %v, %v], %v, %v, %v)",
            name, impls, init, pfs, methods, options, def_point, file, row, col,
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
    // supers? = @is typelist
    "supers": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            return fmt.Sprintf("[%v]", Transpile(tree, children["typelist"]))
        } else {
            return "[]"
        }
    },
    // typelist = type typelist_tail
    "typelist": func (tree Tree, ptr int) string {
        var type_ptrs = FlatSubTree(tree, ptr, "type", "typelist_tail")
        var buf strings.Builder
        for i, type_ptr := range type_ptrs {
            buf.WriteString(Transpile(tree, type_ptr))
            if i != len(type_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        return buf.String()
    },
    // init = @init Call paralist_strict! body!
    "init": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var body_ptr = children["body"]
        var class_ptr = tree.Nodes[ptr].Parent
        var class_children = Children(tree, class_ptr)
        var name_ptr = class_children["name"]
        var desc = Desc (
            GetTokenContent(tree, name_ptr),
            GetWholeContent(tree, params_ptr),
            []rune("Instance"),
        )
        return Function (
            tree, body_ptr, F_Sync,
            desc, parameters, "__.i",
        )
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
            "__.c(__.ci, [%v, %v, %v], %v, %v, %v)",
            name, members, def_point, file, row, col,
        )
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, interface_)
        } else {
            value = interface_
        }
        return fmt.Sprintf (
            "__.c(dl, [%v, %v, true, __.t], %v, %v, %v)",
            name, value, file, row, col,
        )
    },
    // members? = member members
    "members": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "[]" }
        var member_ptrs = FlatSubTree(tree, ptr, "member", "members")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, member_ptr := range member_ptrs {
            buf.WriteString(Transpile(tree, member_ptr))
            if i != len(member_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
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