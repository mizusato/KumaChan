package transpiler

import "strings"


var Functions = map[string]TransFunction {
    // function = fun_header name Call paralist_strict! ret {! body }!
    "function": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        return Transpile(tree, children["paralist_strict"])
    },
    // paralist_strict = ( ) | ( typed_list! )!
    "paralist_strict": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var list, exists = children["typed_list"]
        if exists {
            return Transpile(tree, list)
        } else {
            return "[]"
        }
    },
    // typed_list = typed_list_item typed_list_tail
    "typed_list": func (tree Tree, ptr int) string {
        var items = FlatSubTree(tree, ptr, "typed_list_item", "typed_list_tail")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, item := range items {
            buf.WriteString(Transpile(tree, item))
            if i != len(items)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
    },
    // typed_list_item = name :! type!
    "typed_list_item": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var type_ = Transpile(tree, children["type"])
        var buf strings.Builder
        buf.WriteRune('{')
        buf.WriteString("name: ")
        buf.WriteString(name)
        buf.WriteString(", ")
        buf.WriteString("type: ")
        buf.WriteString(type_)
        buf.WriteRune('}')
        return buf.String()
    },
    // type = identifier type_gets type_arglist
    "type": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var children = Children(tree, ptr)
        var expr, exists = children["expr"]
        if exists { return Transpile(tree, expr) }
        var id = Transpile(tree, children["identifier"])
        var gets_ptr = children["type_gets"]
        var gets = FlatSubTree(tree, gets_ptr, "type_get", "type_gets")
        var t = id
        for _, get := range gets {
            // type_get = . name
            var key = TranspileLastChild(tree, get)
            var row, col = GetRowColInfo(tree, get)
            var buf strings.Builder
            buf.WriteString("g")
            buf.WriteRune('(')
            WriteList(&buf, []string {
                t, key, "false", file, row, col,
            })
            buf.WriteRune(')')
            t = buf.String()
        }
        var arglist_ptr = children["type_arglist"]
        if NotEmpty(tree, arglist_ptr) {
            var arglist = Transpile(tree, arglist_ptr)
            var row, col = GetRowColInfo(tree, arglist_ptr)
            var buf strings.Builder
            buf.WriteString("c")
            buf.WriteRune('(')
            WriteList(&buf, []string {
                t, arglist, file, row, col,
            })
            buf.WriteRune(')')
            return buf.String()
        } else {
            return t
        }
    },
    // type_args? = Call < type_arglist! >!
    "type_args": TranspileChild("type_arglist"),
    // type_arglist = type_arg type_arglist_tail
    "type_arglist": func (tree Tree, ptr int) string {
        var args = FlatSubTree(tree, ptr, "type_arg", "type_arglist_tail")
        var buf strings.Builder
        buf.WriteRune('[')
        for i, arg := range args {
            buf.WriteString(Transpile(tree, arg))
            if i != len(args)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteRune(']')
        return buf.String()
    },
    // type_arg = type | primitive
    "type_arg": TranspileFirstChild,
}
