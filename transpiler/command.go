package transpiler

import "fmt"
import "strings"
import "../parser"
import "../parser/syntax"


var CommandMap = map[string]TransFunction {
    // command = cmd_group1 | cmd_group2 | cmd_group3
    "command": TranspileFirstChild,
    // cmd_group1 = cmd_flow | cmd_pause | cmd_err | cmd_return
    "cmd_group1": TranspileFirstChild,
    // cmd_group2 = cmd_module | cmd_scope | cmd_def
    "cmd_group2": TranspileFirstChild,
    // cmd_group3 = cmd_pass | cmd_set | cmd_exec
    "cmd_group3": TranspileFirstChild,
    // cmd_def = function | schema | enum | class | interface
    "cmd_def": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var fun_ptr, is_fun = children["function"]
        fun_ptr = tree.Nodes[fun_ptr].Children[0]
        if is_fun {
            var f = Transpile(tree, fun_ptr)
            var f_children = Children(tree, fun_ptr)
            var name = Transpile(tree, f_children["name"])
            var file = GetFileName(tree)
            var row, col = GetRowColInfo(tree, ptr)
            return fmt.Sprintf(
                "%v(%v, [%v, %v], %v, %v, %v)",
                G(CALL), L_ADD_FUN, name, f, file, row, col,
            )
        } else {
            return TranspileFirstChild(tree, ptr)
        }
    },
    // cmd_exec = expr
    "cmd_exec": TranspileFirstChild,
    // cmd_return = @return Void | @return expr
    "cmd_return": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr, exists = children["expr"]
        if exists {
            return fmt.Sprintf("return %v", Transpile(tree, expr))
        } else {
            return fmt.Sprintf("return %v", G(T_VOID))
        }
    },
    // cmd_scope = cmd_let | cmd_var | cmd_reset
    "cmd_scope": TranspileFirstChild,
    // cmd_let = @let name var_type = expr | @let pattern = expr
    "cmd_let": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["expr"])
        var pattern_ptr, use_pattern = children["pattern"]
        if use_pattern {
            var pattern = Transpile(tree, pattern_ptr)
            return fmt.Sprintf (
                "%v(%v, [%v, %v, %v], %v, %v, %v)",
                G(CALL), L_MATCH, "true", pattern, value, file, row, col,
            )
        }
        var name = Transpile(tree, children["name"])
        var T = Transpile(tree, children["var_type"])
        return fmt.Sprintf(
            "%v(%v, [%v, %v, true, %v], %v, %v, %v)",
            G(CALL), L_VAR_DECL, name, value, T, file, row, col,
        )
    },
    // cmd_type = @type name = @singleton | @type name generic_params = expr
    "cmd_type": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var _, is_singleton = children["@singleton"]
        if is_singleton {
            var name = Transpile(tree, name_ptr)
            var singleton = fmt.Sprintf (
                "%v(%v)", G(C_SINGLETON), name,
            )
            return fmt.Sprintf(
                "%v(%v, [%v, %v, true, %v], %v, %v, %v)",
                G(CALL), L_VAR_DECL, name, singleton, G(T_TYPE),
                file, row, col,
            )
        }
        var gp_ptr = children["generic_params"]
        var name = Transpile(tree, name_ptr)
        var expr = Transpile(tree, children["expr"])
        var value string
        if NotEmpty(tree, gp_ptr) {
            value = TypeTemplate(tree, gp_ptr, name_ptr, expr)
        } else {
            value = expr
        }
        return fmt.Sprintf(
            "%v(%v, [%v, %v, true, %v], %v, %v, %v)",
            G(CALL), L_VAR_DECL, name, value, G(T_TYPE), file, row, col,
        )
    },
    // cmd_var = @var name var_type = expr | @var pattern = expr
    "cmd_var": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["expr"])
        var pattern_ptr, use_pattern = children["pattern"]
        if use_pattern {
            var pattern = Transpile(tree, pattern_ptr)
            return fmt.Sprintf (
                "%v(%v, [%v, %v, %v], %v, %v, %v)",
                G(CALL), L_MATCH, "false", pattern, value, file, row, col,
            )
        }
        var name = Transpile(tree, children["name"])
        var T = Transpile(tree, children["var_type"])
        return fmt.Sprintf (
            "%v(%v, [%v, %v, false, %v], %v, %v, %v)",
            G(CALL), L_VAR_DECL, name, value, T, file, row, col,
        )
    },
    // pattern = pattern_key | pattern_index
    "pattern": TranspileFirstChild,
    // pattern_key = { sub_pattern_list! }! nil_flag
    "pattern_key": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var allow_nil = Transpile(tree, children["nil_flag"])
        var items = SubPatternList(tree, children["sub_pattern_list"], false)
        return fmt.Sprintf (
            "{ is_final: false, extract: null, allow_nil: %v, items: %v }",
            allow_nil, items,
        )
    },
    // pattern_index = [ sub_pattern_list! ]! nil_flag
    "pattern_index": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var allow_nil = Transpile(tree, children["nil_flag"])
        var items = SubPatternList(tree, children["sub_pattern_list"], true)
        return fmt.Sprintf (
            "{ is_final: false, extract: null, allow_nil: %v, items: %v }",
            allow_nil, items,
        )
    },
    // cmd_pause = cmd_yield | cmd_async_for | cmd_await
    "cmd_pause": TranspileFirstChild,
    // cmd_yield = @yield name var_type = expr! | @yield expr!
    "cmd_yield": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr, is_decl = children["name"]
        var T_ptr = children["var_type"]
        var value = Transpile(tree, children["expr"])
        if is_decl {
            var name = Transpile(tree, name_ptr)
            var T = Transpile(tree, T_ptr)
            var file = GetFileName(tree)
            var row, col = GetRowColInfo(tree, ptr)
            return fmt.Sprintf(
                "%v(%v, [%v, (yield %v), false, %v], %v, %v, %v)",
                G(CALL), L_VAR_DECL, name, value, T, file, row, col,
            )
        } else {
            return fmt.Sprintf("((yield %v), %v)", value, G(T_VOID))
        }
    },
    // cmd_await = @await name var_type = expr! | @await expr!
    "cmd_await": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name_ptr, is_decl = children["name"]
        var T_ptr = children["var_type"]
        var expr = Transpile(tree, children["expr"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var await = fmt.Sprintf (
            "(await %v(%v, [%v], %v, %v, %v))",
            G(CALL), G(REQ_PROMISE), expr, file, row, col,
        )
        if is_decl {
            var name = Transpile(tree, name_ptr)
            var T = Transpile(tree, T_ptr)
            return fmt.Sprintf (
                "%v(%v, [%v, %v, false, %v], %v, %v, %v)",
                G(CALL), L_VAR_DECL, name, await, T, file, row, col,
            )
        } else {
            return fmt.Sprintf("(%v, %v)", await, G(T_VOID))
        }
    },
    // cmd_reset = @reset name = expr
    "cmd_reset": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var name_ptr = children["name"]
        var name = Transpile(tree, name_ptr)
        var value = Transpile(tree, children["expr"])
        var op_ptr = children["set_op"]
        if NotEmpty(tree, op_ptr) {
            value = fmt.Sprintf(
                "%v(%v, [%v, %v], %v, %v, %v)",
                G(CALL),
                Transpile(tree, op_ptr),
                VarLookup(tree, name_ptr),
                value, file, row, col,
            )
        }
        return fmt.Sprintf(
            "%v(%v, [%v, %v], %v, %v, %v)",
            G(CALL), L_VAR_RESET, name, value, file, row, col,
        )
    },
    "set_op": func (tree Tree, ptr int) string {
        return TransMapByName["operator"](tree, ptr)
    },
    // var_type? = : type
    "var_type": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            return TranspileLastChild(tree, ptr)
        } else {
            return G(T_ANY)
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
                "%v(%v, %v, %v, %v, %v, %v)",
                G(GET), t, key, nil_flag, file, row, col,
            )
        }
        var object = t
        var row, col = GetRowColInfo(tree, ptr)
        var set_key, _ = GetKey(tree, tail)
        var op_ptr = children["set_op"]
        if NotEmpty(tree, op_ptr) {
            var operator = Transpile(tree, op_ptr)
            var previous = fmt.Sprintf(
                "%v(%v, %v, %v, %v, %v, %v)",
                G(GET), object, set_key, "false", file, row, col,
            )
            value = fmt.Sprintf(
                "%v(%v, [%v, %v], %v, %v, %v)",
                G(CALL), operator, previous, value, file, row, col,
            )
        }
        return fmt.Sprintf(
            "%v(%v, [%v, %v, %v], %v, %v, %v)",
            G(CALL), G(SET), object, set_key, value, file, row, col,
        )
    },
    // cmd_flow = cmd_if | cmd_switch | cmd_while | cmd_for | cmd_loop_ctrl
    "cmd_flow": TranspileFirstChild,
    // block = { commands }!
    "block": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var commands = Commands(tree, children["commands"], false)
        var BodyId = syntax.Name2Id["body"]
        var BlockId = syntax.Name2Id["block"]
        var HandleId = syntax.Name2Id["handle_hook"]
        var ForId = syntax.Name2Id["cmd_for"]
        var AsyncForId = syntax.Name2Id["cmd_async_for"]
        var depth = 0
        var node = &tree.Nodes[ptr]
        for node.Part.Id != BodyId && node.Part.Id != HandleId {
            if node.Part.Id == BlockId {
                depth += 1
            }
            if node.Parent < 0 {
                break
            }
            node = &tree.Nodes[node.Parent]
        }
        var upper string
        if depth-1 > 0 {
            upper = fmt.Sprintf("%v%v", SCOPE, depth-1)
        } else {
            if node.Part.Id == HandleId {
                upper = H_HOOK_SCOPE
            } else {
                upper = SCOPE
            }
        }
        var current = fmt.Sprintf("%v%v", SCOPE, depth)
        var buf strings.Builder
        buf.WriteString("{ ")
        fmt.Fprintf(
            &buf, "let %v = %v.%v(%v); ",
            current, RUNTIME, R_NEW_SCOPE, upper,
        )
        WriteHelpers(&buf, current)
        var parent_node = tree.Nodes[tree.Nodes[ptr].Parent]
        var parent_part_id = parent_node.Part.Id
        if parent_part_id == ForId || parent_part_id == AsyncForId {
            fmt.Fprintf (
                &buf, "if (l.key) { %v(l.key, I.key) }; ",
                L_VAR_DECL,
            )
            fmt.Fprintf (
                &buf, "if (l.value) { %v(l.value, I.value) }; ",
                L_VAR_DECL,
            )
        }
        buf.WriteString(commands)
        buf.WriteString(" }")
        return buf.String()
    },
    // cmd_if = @if expr! block! elifs else
    "cmd_if": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr_ptr = children["expr"]
        var condition = Transpile(tree, expr_ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, expr_ptr)
        var block = Transpile(tree, children["block"])
        var elifs = Transpile(tree, children["elifs"])
        var else_ = Transpile(tree, children["else"])
        return fmt.Sprintf (
            "if (%v(%v, [%v], %v, %v, %v)) %v%v%v",
            G(CALL), G(REQ_BOOL), condition, file, row, col,
            block, elifs, else_,
        )
    },
    // elifs? = elif elifs
    "elifs": func (tree Tree, ptr int) string {
        var elif_ptrs = FlatSubTree(tree, ptr, "elif", "elifs")
        var buf strings.Builder
        for _, elif_ptr := range elif_ptrs {
            buf.WriteString(Transpile(tree, elif_ptr))
        }
        return buf.String()
    },
    // elif = @else @if expr! block! | @elif expr! block!
    "elif": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr_ptr = children["expr"]
        var condition = Transpile(tree, expr_ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, expr_ptr)
        var block = Transpile(tree, children["block"])
        return fmt.Sprintf(
            " else if (%v(%v, [%v], %v, %v, %v)) %v",
            G(CALL), G(REQ_BOOL), condition, file, row, col, block,
        )
    },
    // else? = @else block!
    "else": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var block = Transpile(tree, children["block"])
            return fmt.Sprintf(" else %v", block)
        } else {
            return ""
        }
    },
    // cmd_switch = @switch { cases default }!
    "cmd_switch": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var cases = Transpile(tree, children["cases"])
        var default_ = Transpile(tree, children["default"])
        return fmt.Sprintf (
            "if (false) { void(0) }%v%v",
            cases, default_,
        )
    },
    // cases? = case cases
    "cases": func (tree Tree, ptr int) string {
        var case_ptrs = FlatSubTree(tree, ptr, "case", "cases")
        var buf strings.Builder
        for _, case_ptr := range case_ptrs {
            buf.WriteString(Transpile(tree, case_ptr))
        }
        return buf.String()
    },
    // case = @case expr! block!
    "case": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr_ptr = children["expr"]
        var condition = Transpile(tree, expr_ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, expr_ptr)
        var block = Transpile(tree, children["block"])
        return fmt.Sprintf(
            " else if (%v(%v, [%v], %v, %v, %v)) %v",
            G(CALL), G(REQ_BOOL), condition, file, row, col, block,
        )
    },
    // default? = @default block!
    "default": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var block = Transpile(tree, children["block"])
            return fmt.Sprintf(" else %v", block)
        } else {
            var file = GetFileName(tree)
            var row, col = GetRowColInfo(tree, ptr)
            return fmt.Sprintf (
                " else { %v(%v, [], %v, %v, %v) }",
                G(CALL), G(SWITCH_FAILED), file, row, col,
            )
        }
    },
    // cmd_for = @for for_params! @in expr! block!
    "cmd_for": func (tree Tree, ptr int) string {
        // note: rule name "cmd_for" is depended by CommandMap["block"]
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var params_ptr = children["for_params"]
        var params = Transpile(tree, params_ptr)
        var expr = Transpile(tree, children["expr"])
        var block = Transpile(tree, children["block"])
        var c = tree.Nodes[params_ptr].Children[0]
        var params_type = syntax.Id2Name[tree.Nodes[c].Part.Id]
        var loop_type string
        switch params_type {
        case "for_params_list":
            loop_type = G(FOR_LOOP_ITER)
        case "for_params_hash":
            loop_type = G(FOR_LOOP_ENUM)
        case "for_params_value":
            loop_type = G(FOR_LOOP_VALUE)
        default:
            panic("impossible switch branch")
        }
        return fmt.Sprintf (
            "for (let I of %v(%v, [%v], %v, %v, %v)) { %v %v; }",
            G(CALL), loop_type, expr, file, row, col, params, block,
        )
    },
    // cmd_async_for = @await name @in expr! block!
    "cmd_async_for": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var item_name = Transpile(tree, children["name"])
        var expr = Transpile(tree, children["expr"])
        var block = Transpile(tree, children["block"])
        var params = fmt.Sprintf("let l = { value: %v };", item_name)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "for await (let I of %v(%v, [%v], %v, %v, %v)) { %v %v; }",
            G(CALL), G(FOR_LOOP_ASYNC), expr, file, row, col, params, block,
        )
    },
    // for_params = for_params_list | for_params_hash | for_params_value
    "for_params": TranspileFirstChild,
    // for_params_list = for_value [ for_index! ]!
    "for_params_list": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["for_value"])
        var key = Transpile(tree, children["for_index"])
        if key == value {
            parser.Error (
                tree, ptr, fmt.Sprintf (
                    "duplicate parameter name %v in for statement",
                    key,
                ),
            )
        }
        return fmt.Sprintf("let l = { key: %v, value: %v };", key, value)
    },
    // for_params_hash = { for_key :! for_value! }!
    "for_params_hash": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var key = Transpile(tree, children["for_key"])
        var value = Transpile(tree, children["for_value"])
        return fmt.Sprintf("let l = { key: %v, value: %v };", key, value)
    },
    // for_params_value = for_value
    "for_params_value": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["for_value"])
        return fmt.Sprintf("let l = { value: %v };", value)
    },
    // for_value = name
    "for_value": TranspileFirstChild,
    // for_index = name
    "for_index": TranspileFirstChild,
    // for_key = name
    "for_key": TranspileFirstChild,
    // cmd_while = @while expr! block!
    "cmd_while": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var expr_ptr = children["expr"]
        var condition = Transpile(tree, expr_ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, expr_ptr)
        var block = Transpile(tree, children["block"])
        return fmt.Sprintf (
            "while (%v(%v, [%v], %v, %v, %v)) %v",
            G(CALL), G(REQ_BOOL), condition, file, row, col, block,
        )
    },
    // cmd_loop_ctrl = @break | @continue
    "cmd_loop_ctrl": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        var id = tree.Nodes[child_ptr].Part.Id
        var name = syntax.Id2Name[id]
        if name == "@break" {
            return "break"
        } else if name == "@continue" {
            return "continue"
        } else {
            panic("impossible branch")
        }
    },
    // cmd_err = cmd_throw | cmd_assert | cmd_ensure | cmd_try | cmd_panic
    "cmd_err": TranspileFirstChild,
    // cmd_throw = @throw expr!
    "cmd_throw": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), G(THROW), expr, file, row, col,
        )
    },
    // cmd_assert = @assert expr!
    "cmd_assert": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), G(ASSERT), expr, file, row, col,
        )
    },
    // cmd_panic = @panic expr!
    "cmd_panic": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), G(PANIC), expr, file, row, col,
        )
    },
    // cmd_ensure = @ensure name! ensure_args { expr! }!
    "cmd_ensure": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var args = Transpile(tree, children["ensure_args"])
        var expr_ptr = children["expr"]
        var expr = Transpile(tree, expr_ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var e_row, e_col = GetRowColInfo(tree, expr_ptr)
        return fmt.Sprintf (
            "if (!(%v(%v, [%v], %v, %v, %v))) " +
                "{ %v(%v, %v, %v, %v, %v, %v) }",
            G(CALL), G(REQ_BOOL), expr, file, e_row, e_col,
            G(ENSURE_FAILED), ERROR_DUMP, name, args, file, row, col,
        )
    },
    // ensure_args? = Call ( exprlist )
    "ensure_args": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var exprlist = Transpile(tree, children["exprlist"])
            return fmt.Sprintf("[%v]", exprlist)
        } else {
            return "[]"
        }
    },
    // cmd_try = @try opt_to name { commands }!
    "cmd_try": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var commands = Commands(tree, children["commands"], false)
        return fmt.Sprintf(
            "try { %v } catch (%v) { %v(%v, %v, %v) }",
            commands, TRY_ERROR, G(TRY_FAILED), ERROR_DUMP, TRY_ERROR, name,
        )
    },
    // cmd_pass = @do @nothing
    "cmd_pass": func (tree Tree, ptr int) string {
        return G(T_VOID)
    },
    // cmd_module = cmd_import
    "cmd_module": TranspileFirstChild,
    // cmd_import = import_all | import_names | import_module
    "cmd_import": TranspileFirstChild,
    // import_names = @import as_list @from name
    "import_names": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var module_name = Transpile(tree, children["name"])
        var configs = Transpile(tree, children["as_list"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v, %v], %v, %v, %v)",
            G(CALL), L_IMPORT_VAR, module_name, configs, file, row, col,
        )
    },
    // import_module = @import as_item
    "import_module": func (tree Tree, ptr int) string {
        var config = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), L_IMPORT_MOD, config, file, row, col,
        )
    },
    // import_all = @import * @from name
    "import_all": func (tree Tree, ptr int) string {
        var name = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf (
            "%v(%v, [%v], %v, %v, %v)",
            G(CALL), L_IMPORT_ALL, name, file, row, col,
        )
    },
    // as_list = as_item as_list_tail
    "as_list": func (tree Tree, ptr int) string {
        return TranspileSubTree(tree, ptr, "as_item", "as_list_tail")
    },
    // as_item = name @as name! | name
    "as_item": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var _, has_alias = children["@as"]
        if has_alias {
            var name = TranspileFirstChild(tree, ptr)
            var alias = TranspileLastChild(tree, ptr)
            return fmt.Sprintf("[%v, %v]", name, alias)
        } else {
            var name = TranspileFirstChild(tree, ptr)
            return fmt.Sprintf("[%v, %v]", name, name)
        }
    },
}
