package transpiler

import "fmt"
import "strings"
import "../parser"
import "../parser/syntax"


var FunctionMap = map[string]TransFunction {
    // function = f_sync
    "function": TranspileFirstChild,
    // f_sync = @function name Call paralist_strict! ->! type! body!
    "f_sync": func (tree Tree, ptr int) string {
        // the rule name "f_sync" is depended by OO_Map["method_implemented"]
        // the rule name "f_sync" is also depended by MethodTable()
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
    // lambda = lambda_block | lambda_inline
    "lambda": TranspileFirstChild,
    // lambda_block = @lambda paralist_block ret_lambda body!
    "lambda_block": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var paralist_ptr = children["paralist_block"]
        var parameters = Transpile(tree, paralist_ptr)
        var ret_ptr = children["ret_lambda"]
        var value_type = Transpile(tree, ret_ptr)
        var body_ptr = children["body"]
        var value_type_desc string
        if value_type == G(T_ANY) {
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
    // ret_lambda? = -> type | ->
    "ret_lambda": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var type_ptr, type_specified = children["type"]
            if type_specified {
                return Transpile(tree, type_ptr)
            } else {
                return G(T_ANY)
            }
        } else {
            return G(T_ANY)
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
            "{ parameters: %v, value_type: %v }",
            parameters, G(T_ANY),
        )
        return fmt.Sprintf (
            "%v(%v, %v, %v, %v)",
            L_WRAP, proto, "null", desc, raw,
        )
    },
    // iife = invoke | iterator | promise | async_iterator
    "iife": TranspileFirstChild,
    // invoke = @invoke body
    "invoke": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Object"))
        var f = Function(tree, body_ptr, F_Sync, desc, "[]", G(T_ANY))
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("%v(%v, [], %v, %v, %v)", G(CALL), f, file, row, col)
    },
    // promise = @promise body
    "promise": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Promise"))
        var f = Function(tree, body_ptr, F_Async, desc, "[]", G(T_PROMISE))
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("%v(%v, [], %v, %v, %v)", G(CALL), f, file, row, col)
    },
    // iterator = @iterator body
    "iterator": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("Iterator"))
        var f = Function(tree, body_ptr, F_Generator, desc, "[]", G(T_ITERATOR))
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("%v(%v, [], %v, %v, %v)", G(CALL), f, file, row, col)
    },
    // async_iterator = @async @iterator body!
    "async_iterator": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var body_ptr = children["body"]
        var desc = Desc([]rune("IIFE"), []rune("()"), []rune("AsyncIterator"))
        var f = Function (
            tree, body_ptr, F_AsyncGenerator, desc, "[]", G(T_ASYNC_ITERATOR),
        )
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf("%v(%v, [], %v, %v, %v)", G(CALL), f, file, row, col)
    },
    // body = { static_commands commands mock_hook handle_hook }!
    "body": func (tree Tree, ptr int) string {
        // note: the rule name "body" is depended by CommandMap["block"]
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
            fmt.Fprintf(&buf, "let %v = {}; ", ERROR_DUMP)
            buf.WriteString("try { ")
            buf.WriteString(commands)
            buf.WriteString(" } ")
            buf.WriteString(catch_and_finally)
            fmt.Fprintf(&buf, " return %v;", G(T_VOID))
            return buf.String()
        } else {
            return commands
        }
    },
    // handle_hook? = _at @handle name { handle_cmds }! finally
    "handle_hook": func (tree Tree, ptr int) string {
        // note: the rule name "handle_hook" is depended by CommandMap["block"]
        var children = Children(tree, ptr)
        var error_name = Transpile(tree, children["name"])
        var buf strings.Builder
        fmt.Fprintf(&buf, "catch (%v) { ", H_HOOK_ERROR)
        fmt.Fprintf (
            &buf, "let %v = %v.%v(%v); ",
            H_HOOK_SCOPE, RUNTIME, R_NEW_SCOPE, SCOPE,
        )
        WriteHelpers(&buf, H_HOOK_SCOPE)
        fmt.Fprintf(&buf, "%v(%v); ", G(ENTER_H_HOOK), H_HOOK_ERROR)
        fmt.Fprintf(&buf, "%v(%v, %v); ", L_VAR_DECL, error_name, H_HOOK_ERROR)
        buf.WriteString(Transpile(tree, children["handle_cmds"]))
        fmt.Fprintf(&buf, " %v(%v); ", G(EXIT_H_HOOK), H_HOOK_ERROR)
        buf.WriteString("}")
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
        fmt.Fprintf (
            &buf, "if (%v.%v === '%v' && %v.%v === %v)",
            ERROR_DUMP, DUMP_TYPE, DUMP_ENSURE, ERROR_DUMP, DUMP_NAME, name,
        )
        buf.WriteString(" { ")
        buf.WriteString(G(CALL))
        buf.WriteString("(")
        WriteList(&buf, []string {
            G(INJECT_E_ARGS),
            fmt.Sprintf("[%v, %v, %v]", H_HOOK_SCOPE, params, ERROR_DUMP),
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
        fmt.Fprintf (
            &buf, "if (%v.%v === '%v' && %v.%v === %v)",
            ERROR_DUMP, DUMP_TYPE, DUMP_TRY, ERROR_DUMP, DUMP_NAME, name,
        )
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
                return fmt.Sprintf("[{ name: %v, type: %v }]", name, G(T_ANY))
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
        var l = FlatSubTree(tree, ptr, "typed_list_item", "typed_list_tail")
        var occurred = make(map[string]bool)
        var buf strings.Builder
        buf.WriteRune('[')
        for i, item_ptr := range l {
            var item_children = Children(tree, item_ptr)
            var name = Transpile(tree, item_children["name"])
            if occurred[name] {
                parser.Error (
                    tree, item_ptr, fmt.Sprintf (
                        "duplicate parameter %v",
                        name,
                    ),
                )
            }
            occurred[name] = true
            buf.WriteString(Transpile(tree, item_ptr))
            if i != len(l)-1 {
                buf.WriteString(",")
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
    // operator_defs? = operator_def operator_defs
    "operator_defs": func (tree Tree, ptr int) string {
        if Empty(tree, ptr) { return "{}" }
        var parent = tree.Nodes[ptr].Parent
        var parent_id = tree.Nodes[parent].Part.Id
        var is_schema bool
        if parent_id == syntax.Name2Id["schema_config"] {
            is_schema = true
        } else if parent_id == syntax.Name2Id["class_opt"] {
            is_schema = false
        } else {
            panic("impossible branch")
        }
        var def_ptrs = FlatSubTree(tree, ptr, "operator_def", "operator_defs")
        var defined = make(map[string]bool)
        var buf strings.Builder
        buf.WriteString("{ ")
        for i, def_ptr := range def_ptrs {
            // operator_def = @operator general_op operator_def_fun
            var children = Children(tree, def_ptr)
            var op_ptr = children["general_op"]
            var op_name, can_redef = GetGeneralOperatorName(tree, op_ptr)
            if !can_redef {
                parser.Error (
                    tree, def_ptr, fmt.Sprintf (
                        "cannot overload non-redefinable operator %v",
                        op_name,
                    ),
                )
            }
            if is_schema && op_name == "copy" {
                parser.Error (
                    tree, def_ptr,
                    "cannot overload copy operator on struct",
                )
            }
            if defined[op_name] {
                parser.Error (
                    tree, def_ptr, fmt.Sprintf (
                        "duplicate definition of operator %v",
                        op_name,
                    ),
                )
            }
            defined[op_name] = true
            var op_escaped = EscapeRawString([]rune(op_name))
            var op_fun = Transpile(tree, children["operator_def_fun"])
            fmt.Fprintf(&buf, "%v: %v", op_escaped, op_fun)
            if i != len(def_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        buf.WriteString(" }")
        return buf.String()
    },
    // operator_def_fun = (! namelist! )! opt_arrow body!
    "operator_def_fun": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var namelist_ptr = children["namelist"]
        var body_ptr = children["body"]
        var parameters = UntypedParameterList(tree, namelist_ptr)
        var desc = Desc (
            []rune("redefined_operator"),
            GetWholeContent(tree, namelist_ptr),
            []rune("Object"),
        )
        return Function(tree, body_ptr, F_Sync, desc, parameters, G(T_ANY))
    },
}
