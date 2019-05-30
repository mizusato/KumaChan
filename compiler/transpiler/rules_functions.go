package transpiler

import "fmt"
import "strings"


var Functions = map[string]TransFunction {
    // function = f_sync | f_async | generator
    "function": TranspileFirstChild,
    // f_sync = @function name Call paralist_strict! -> type body
    "f_sync": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var type_ptr = children["type"]
        var value_type = Transpile(tree, type_ptr)
        var body_ptr = children["body"]
        var desc = Desc (
            GetWholeContent(tree, name_ptr),
            GetWholeContent(tree, params_ptr),
            GetWholeContent(tree, type_ptr),
        )
        return Function (
            tree, body_ptr, F_Sync,
            desc, parameters, value_type,
        )
    },
    // f_async = @async name Call paralist_strict! body
    "f_async": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var body_ptr = children["body"]
        var desc = Desc (
            GetWholeContent(tree, name_ptr),
            GetWholeContent(tree, params_ptr),
            []rune("Promise"),
        )
        return Function (
            tree, body_ptr, F_Async,
            desc, parameters, "__.pm",
        )
    },
    // generator = @generator name Call paralist_strict! body
    "generator": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var params_ptr = children["paralist_strict"]
        var parameters = Transpile(tree, params_ptr)
        var body_ptr = children["body"]
        var desc = Desc (
            GetWholeContent(tree, name_ptr),
            GetWholeContent(tree, params_ptr),
            []rune("Iterator"),
        )
        return Function (
            tree, body_ptr, F_Generator,
            desc, parameters, "__.it",
        )
    },
    // lambda = lambda_block | lambda_inline
    "lambda": TranspileFirstChild,
    // iife = invoke | iterator | promise
    "iife": TranspileFirstChild,
    // lambda_block = lambda_sync | lambda_async | lambda_generator
    "lambda_block": TranspileFirstChild,
    // lambda_sync = @lambda paralist_block ret_lambda body!
    "lambda_sync": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var paralist_ptr = children["paralist_block"]
        var parameters = Transpile(tree, paralist_ptr)
        var ret_ptr = children["ret_lambda"]
        var value_type = Transpile(tree, ret_ptr)
        var body_ptr = children["body"]
        var value_type_desc string
        if value_type == "__.a" {
            value_type_desc = "Object"
        } else {
            var t = string(GetWholeContent(tree, ret_ptr))
            t = strings.TrimPrefix(t, "->")
            t = strings.TrimLeft(t, " ")
            value_type_desc = t
        }
        var desc = Desc (
            []rune("lambda"),
            GetWholeContent(tree, paralist_ptr),
            []rune(value_type_desc),
        )
        return Function (
            tree, body_ptr, F_Sync,
            desc, parameters, value_type,
        )
    },
    // invoke = @invoke body
    "invoke": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Object"))
        var f = Function(tree, body_ptr, F_Sync, desc, "[]", "__.a")
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("__.c(%v, [], %v, %v, %v)", f, file, row, col)
    },
    // lambda_async = @async paralist_block opt_arrow body!
    "lambda_async": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var paralist_ptr = children["paralist_block"]
        var parameters = Transpile(tree, paralist_ptr)
        var body_ptr = children["body"]
        var desc = Desc (
            []rune("lambda"),
            GetWholeContent(tree, paralist_ptr),
            []rune("Promise"),
        )
        return Function (
            tree, body_ptr, F_Async,
            desc, parameters, "__.pm",
        )
    },
    // promise = @promise body
    "promise": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Promise"))
        var f = Function(tree, body_ptr, F_Async, desc, "[]", "__.pm")
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("__.c(%v, [], %v, %v, %v)", f, file, row, col)
    },
    // lambda_generator = @generator paralist_block opt_arrow body!
    "lambda_generator": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var paralist_ptr = children["paralist_block"]
        var parameters = Transpile(tree, paralist_ptr)
        var body_ptr = children["body"]
        var desc = Desc (
            []rune("lambda"),
            GetWholeContent(tree, paralist_ptr),
            []rune("Iterator"),
        )
        return Function (
            tree, body_ptr, F_Generator,
            desc, parameters, "__.it",
        )
    },
    // iterator = @iterator body
    "iterator": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Iterator"))
        var f = Function(tree, body_ptr, F_Generator, desc, "[]", "__.it")
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("__.c(%v, [], %v, %v, %v)", f, file, row, col)
    },
    // ret_lambda? = -> type | ->
    "ret_lambda": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var type_ptr, type_specified = children["type"]
            if type_specified {
                return Transpile(tree, type_ptr)
            } else {
                return "__.a"
            }
        } else {
            return "__.a"
        }
    },
    // lambda_inline = .{ paralist_inline expr! }!
    "lambda_inline": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var l_ptr = children["paralist_inline"]
        var parameters string
        var desc string
        if NotEmpty(tree, l_ptr) {
            parameters = Transpile(tree, l_ptr)
            var l_children = Children(tree, l_ptr)
            desc = Desc (
                []rune("lambda"),
                GetWholeContent(tree, l_children["namelist"]),
                []rune("Object"),
            )
        } else {
            var names = SearchDotParameters(tree, ptr)
            parameters = UntypedParameters(names)
            desc = Desc (
                []rune("lambda"),
                []rune(strings.Join(names, ", ")),
                []rune("Object"),
            )
        }
        var expr = Transpile(tree, children["expr"])
        var raw = BareFunction(fmt.Sprintf("return %v;", expr))
        var proto = fmt.Sprintf (
            "{ parameters: %v, value_type: __.a }",
            parameters,
        )
        return fmt.Sprintf (
            "w(%v, %v, %v, %v)",
            proto, "null", desc, raw,
        )
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
            buf.WriteString("let e = { pointer: __.gp() }; ")
            buf.WriteString("try { ")
            buf.WriteString(commands)
            buf.WriteString(" } ")
            buf.WriteString(catch_and_finally)
            buf.WriteString(" return __.v;")
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
        fmt.Fprintf(&buf, "let handle_scope = %v.new_scope(scope); ", Runtime)
        WriteHelpers(&buf, "handle_scope")
        buf.WriteString("__.rs(e.pointer); ")
        fmt.Fprintf(&buf, "if (error instanceof %v.RuntimeError)", Runtime)
        buf.WriteString(" { throw error; }; ")
        fmt.Fprintf(&buf, "dl(%v, error); ", Transpile(tree, children["name"]))
        buf.WriteString(Transpile(tree, children["handle_cmds"]))
        buf.WriteString(" throw __.c2f(error);")
        buf.WriteString(" }")
        var finally_ptr = children["finally"]
        if NotEmpty(tree, finally_ptr) {
            fmt.Fprintf(
                &buf, " finally { %v };",
                Transpile(tree, finally_ptr),
            )
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
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var buf strings.Builder
        fmt.Fprintf(&buf, "if (e.type === 1 && e.name === %v)", name)
        buf.WriteString(" { ")
        buf.WriteString("__.c(")
        WriteList(&buf, []string {
            "__.ie",
            fmt.Sprintf("[handle_scope, %v, e]", params),
            file, row, col,
        })
        buf.WriteString(");")
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
        fmt.Fprintf(&buf, "if (e.type === 2 && e.name === %v)", name)
        buf.WriteString(" { ")
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
    // paralist_block? = name | Call paralist
    "paralist_block": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var name_ptr, only_name = children["name"]
            if only_name {
                var name = Transpile(tree, name_ptr)
                return fmt.Sprintf("[{ name: %v, type: __.a }]", name)
            } else {
                return Transpile(tree, children["paralist"])
            }
        } else {
            return "[]"
        }
    },
    // paralist = ( ) | ( namelist ) | ( typed_list! )!
    "paralist": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var namelist_ptr, is_namelist = children["namelist"]
        var typed_list_ptr, is_typed_list = children["typed_list"]
        if is_namelist {
            return UntypedParameterList(tree, namelist_ptr)
        } else if is_typed_list {
            return Transpile(tree, typed_list_ptr)
        } else {
            return "[]"
        }
    },
    // paralist_inline? = namelist -->
    "paralist_inline": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var l_ptr = children["namelist"]
            return UntypedParameterList(tree, l_ptr)
        } else {
            panic("trying to transpile empty paralist_inline")
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
        return fmt.Sprintf("{ name: %v, type: %v }", name, type_)
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
            buf.WriteString("__.g")
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
            buf.WriteString("__.c")
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
