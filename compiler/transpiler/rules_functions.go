package transpiler

import "strings"


var Functions = map[string]TransFunction {
    // function = fun_header name Call paralist_strict! type {! body }!
    "function": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var type_ptr = children["type"]
        var value_type = Transpile(tree, type_ptr)
        var body_ptr = children["body"]
        var body = Transpile(tree, body_ptr)
        var body_children = Children(tree, body_ptr)
        // static_commands? = @static { commands }
        var static_ptr = body_children["static_commands"]
        var vals = "null"
        if NotEmpty(tree, static_ptr) {
            var static_commands_ptr = Children(tree, static_ptr)["commands"]
            var static_commands = Commands(tree, static_commands_ptr, true)
            var static_executor = BareFunction(static_commands)
            vals = "gv(" + static_executor + ")"
        }
        var buf strings.Builder
        buf.WriteString("w")
        buf.WriteRune('(')
        buf.WriteString("{ ")
        buf.WriteString("parameters: ")
        buf.WriteString(parameters)
        buf.WriteString(", ")
        buf.WriteString("value_type: ")
        buf.WriteString(value_type)
        buf.WriteString(" }")
        buf.WriteString(", ")
        buf.WriteString(vals)
        buf.WriteString(", ")
        var desc = make([]rune, 0, 120)
        desc = append(desc, GetWholeContent(tree, name_ptr)...)
        desc = append(desc, ' ')
        desc = append(desc, GetWholeContent(tree, params_ptr)...)
        desc = append(desc, []rune(" -> ")...)
        desc = append(desc, GetWholeContent(tree, type_ptr)...)
        buf.WriteString(EscapeRawString(desc))
        buf.WriteString(", ")
        buf.WriteString(BareFunction(body))
        buf.WriteRune(')')
        return buf.String()
    },
    // body = { static_commands commands mock_hook handle_hook }!
    "body": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        // mock_hook? = _at @mock name! { commands }
        var mock_ptr = children["mock_hook"]
        var should_mock = false
        var mock_commands_ptr = -1
        if NotEmpty(tree, mock_ptr) {
            var children = Children(tree, mock_ptr)
            var name_ptr = children["name"]
            var name = string(GetTokenContent(tree, name_ptr))
            for _, mocked := range tree.Mock {
                if name == mocked {
                    should_mock = true
                    mock_commands_ptr = children["commands"]
                    break
                }
            }
        }
        // commands? = command commands
        var commands_ptr = children["commands"]
        if should_mock {
            commands_ptr = mock_commands_ptr
        }
        var commands = Commands(tree, commands_ptr, true)
        var handle_ptr = children["handle_hook"]
        if NotEmpty(tree, handle_ptr) {
            var catch_and_finally = Transpile(tree, handle_ptr)
            var buf strings.Builder
            buf.WriteString("let e = {}; ")
            buf.WriteString("try { ")
            buf.WriteString(commands)
            buf.WriteString(" } ")
            buf.WriteString(catch_and_finally)
            buf.WriteString(" return v;")
            return buf.String()
        } else {
            return commands
        }
    },
    // handle_hook? = _at @handle name { handle_cmds }! finally
    "handle_hook": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var buf strings.Builder
        buf.WriteString("catch (error) { ")
        buf.WriteString("if (error instanceof ")
        buf.WriteString(Runtime)
        buf.WriteString(".RuntimeError")
        buf.WriteString(") { throw error; };")
        buf.WriteString(" let handle_scope = ")
        buf.WriteString(Runtime)
        buf.WriteString(".new_scope(scope); ")
        WriteHelpers(&buf, "handle_scope")
        buf.WriteString("dl(")
        buf.WriteString(Transpile(tree, children["name"]))
        buf.WriteString(", ")
        buf.WriteString("error")
        buf.WriteString("); ")
        buf.WriteString(Transpile(tree, children["handle_cmds"]))
        buf.WriteString(" throw error;")
        buf.WriteString(" }")
        var finally_ptr = children["finally"]
        if NotEmpty(tree, finally_ptr) {
            buf.WriteString(" finally { ")
            buf.WriteString(Transpile(tree, finally_ptr))
            buf.WriteString(" };")
        } else {
            buf.WriteString(";")
        }
        return buf.String()
    },
    // finally? = _at @finally { commands }
    "finally": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        return Commands(tree, children["commands"], false)
    },
    // handle_cmds? = handle_cmd handle_cmds
    "handle_cmds": func (tree Tree, ptr int) string {
        var cmds = FlatSubTree(tree, ptr, "handle_cmd", "handle_cmds")
        var buf strings.Builder
        for _, cmd := range cmds {
            buf.WriteString(Transpile(tree, cmd))
            buf.WriteString("; ")
        }
        return buf.String()
    },
    // handle_cmd = unless | failed | command
    "handle_cmd": TranspileFirstChild,
    // unless = @unless name unless_para { commands }
    "unless": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var params = Transpile(tree, children["unless_para"])
        var buf strings.Builder
        buf.WriteString("if (e.type === 1 && e.name === ")
        buf.WriteString(name)
        buf.WriteString(") { ")
        buf.WriteString("ie(handle_scope, ")
        buf.WriteString(params)
        buf.WriteString(", e);")
        buf.WriteString(Commands(tree, children["commands"], false))
        buf.WriteString(" }")
        return buf.String()
    },
    // unless_para? = Call ( namelist )
    "unless_para": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return Transpile(tree, Children(tree, ptr)["namelist"])
        } else {
            return "[]"
        }
    },
    // failed = @failed opt_to name { commands }
    "failed": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var buf strings.Builder
        buf.WriteString("if (e.type === 2 && e.name === ")
        buf.WriteString(name)
        buf.WriteString(") { ")
        buf.WriteString(Commands(tree, children["commands"], false))
        buf.WriteString(" }")
        return buf.String()
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
    // type = identifier type_gets type_args | ( expr )
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
        var args_ptr = children["type_args"]
        if NotEmpty(tree, args_ptr) {
            var args = Transpile(tree, args_ptr)
            var row, col = GetRowColInfo(tree, args_ptr)
            var buf strings.Builder
            buf.WriteString("c")
            buf.WriteRune('(')
            WriteList(&buf, []string {
                t, args, file, row, col,
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
