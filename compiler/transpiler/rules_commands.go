package transpiler

import "fmt"
import "strings"
import "../syntax"


var CommandsMap = map[string]TransFunction {
    // cmd_def = function | abs_def
    "cmd_def": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var fun_ptr, is_fun = children["function"]
        if is_fun {
            var f = Transpile(tree, fun_ptr)
            var f_children = Children(tree, fun_ptr)
            var name = Transpile(tree, f_children["name"])
            var file = GetFileName(tree)
            var row, col = GetRowColInfo(tree, ptr)
            return fmt.Sprintf(
                "__.c(df, [%v, %v], %v, %v, %v)",
                name, f, file, row, col,
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
            return "return __.v"
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
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf(
            "__.c(dl, [%v, %v, true, %v], %v, %v, %v)",
            name, value, T, file, row, col,
        )
    },
    // cmd_var = @var name var_type = expr
    "cmd_var": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var name = Transpile(tree, children["name"])
        var T = Transpile(tree, children["var_type"])
        var value = Transpile(tree, children["expr"])
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf(
            "__.c(dl, [%v, %v, false, %v], %v, %v, %v)",
            name, value, T, file, row, col,
        )
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
                "__.c(%v, [%v, %v], %v, %v, %v)",
                Transpile(tree, op_ptr),
                TransMapByName["identifier"](tree, name_ptr),
                value, file, row, col,
            )
        }
        return fmt.Sprintf(
            "__.c(rt, [%v, %v], %v, %v, %v)",
            name, value, file, row, col,
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
            return "__.a"   // Types.Any
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
                "__.g(%v, %v, %v, %v, %v, %v)",
                t, key, nil_flag, file, row, col,
            )
        }
        var object = t
        var row, col = GetRowColInfo(tree, ptr)
        var set_key, _ = GetKey(tree, tail)
        var op_ptr = children["set_op"]
        if NotEmpty(tree, op_ptr) {
            var operator = Transpile(tree, op_ptr)
            var previous = fmt.Sprintf(
                "__.g(%v, %v, %v, %v, %v, %v)",
                object, set_key, "false", file, row, col,
            )
            value = fmt.Sprintf(
                "__.c(%v, [%v, %v], %v, %v, %v)",
                operator, previous, value, file, row, col,
            )
        }
        return fmt.Sprintf(
            "__.c(__.s, [%v, %v, %v], %v, %v, %v)",
            object, set_key, value, file, row, col,
        )
    },
    // cmd_flow = cmd_if | cmd_switch | cmd_while | cmd_for | cmd_loop_ctrl
    "cmd_flow": TranspileFirstChild,
    // block = { commands }!
    "block": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var commands = Commands(tree, children["commands"], false)
        var ProgramId = syntax.Name2Id["program"]
        var BodyId = syntax.Name2Id["body"]
        var BlockId = syntax.Name2Id["block"]
        var ForId = syntax.Name2Id["cmd_for"]
        var depth = 0
        var node = tree.Nodes[ptr]
        for node.Part.Id != BodyId && node.Part.Id != ProgramId {
            if node.Part.Id == BlockId {
                depth += 1
            }
            node = tree.Nodes[node.Parent]
        }
        var upper string
        if depth-1 > 0 {
            upper = fmt.Sprintf("scope%v", depth-1)
        } else {
            upper = "scope"
        }
        var current = fmt.Sprintf("scope%v", depth)
        var buf strings.Builder
        buf.WriteString("{ ")
        fmt.Fprintf(
            &buf, "let %v = %v.new_scope(%v); ",
            current, Runtime, upper,
        )
        WriteHelpers(&buf, current)
        var parent_node = tree.Nodes[tree.Nodes[ptr].Parent]
        if parent_node.Part.Id == ForId {
            buf.WriteString("if (l.key) { dl(l.key, I.key) }; ")
            buf.WriteString("if (l.value) { dl(l.value, I.value) }; ")
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
        return fmt.Sprintf(
            "if (__.c(__.rb, [%v], %v, %v, %v)) %v%v%v",
            condition, file, row, col, block, elifs, else_,
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
            " else if (__.c(__.rb, [%v], %v, %v, %v)) %v",
            condition, file, row, col, block,
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
        return fmt.Sprintf("if (false) { void(0) }%v%v", cases, default_)
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
            " else if (__.c(__.rb, [%v], %v, %v, %v)) %v",
            condition, file, row, col, block,
        )
    },
    // default? = @default block!
    "default": func (tree Tree, ptr int) string {
        if NotEmpty(tree, ptr) {
            var children = Children(tree, ptr)
            var block = Transpile(tree, children["block"])
            return fmt.Sprintf(" else %v", block)
        } else {
            return ""
        }
    },
    // cmd_for = @for for_params! @in expr! block!
    "cmd_for": func (tree Tree, ptr int) string {
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        var children = Children(tree, ptr)
        var params = Transpile(tree, children["for_params"])
        var expr = Transpile(tree, children["expr"])
        var block = Transpile(tree, children["block"])
        return fmt.Sprintf(
            "for (let I of __.c(__.f, [%v], %v, %v, %v)) { %v %v; }",
            expr, file, row, col, params, block,
        )
    },
    // for_params = for_params_list | for_params_hash | for_params_value
    "for_params": TranspileFirstChild,
    // for_params_list = for_value [ for_index! ]!
    "for_params_list": func (tree Tree, ptr int) string {
        var children = Children(tree, ptr)
        var value = Transpile(tree, children["for_value"])
        var key = Transpile(tree, children["for_index"])
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
        return fmt.Sprintf(
            "while (__.c(__.rb, [%v], %v, %v, %v)) %v",
            condition, file, row, col, block,
        )
    },
    // cmd_loop_ctrl = @break | @continue
    "cmd_loop_ctrl": func (tree Tree, ptr int) string {
        var child_ptr = tree.Nodes[ptr].Children[0]
        var id = tree.Nodes[child_ptr].Part.Id
        if id == syntax.Name2Id["break"] {
            return "break"
        } else {
            return "continue"
        }
    },
    // cmd_err = cmd_throw | cmd_assert | cmd_ensure | cmd_try | cmd_panic
    "cmd_err": TranspileFirstChild,
    // cmd_throw = @throw expr!
    "cmd_throw": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf(
            "__.c(__.th, [%v], %v, %v, %v)",
            expr, file, row, col,
        )
    },
    // cmd_assert = @assert expr!
    "cmd_assert": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf(
            "__.c(__.as, [%v], %v, %v, %v)",
            expr, file, row, col,
        )
    },
    // cmd_panic = @panic expr!
    "cmd_panic": func (tree Tree, ptr int) string {
        var expr = TranspileLastChild(tree, ptr)
        var file = GetFileName(tree)
        var row, col = GetRowColInfo(tree, ptr)
        return fmt.Sprintf(
            "__.c(__.pa, [%v], %v, %v, %v)",
            expr, file, row, col,
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
        return fmt.Sprintf(
            "if (!(__.c(__.rb, [%v], %v, %v, %v))) " +
                "{ __.ef(e, %v, %v, %v, %v, %v) }",
            expr, file, e_row, e_col,
            name, args, file, row, col,
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
            "try { %v } catch (err) { __.tf(e, err, %v) }",
            commands, name,
        )
    },
}
