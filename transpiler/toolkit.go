package transpiler

import "fmt"
import "strings"
import "../parser"
import "../parser/syntax"


type FunType int
const (
    F_Sync FunType = iota
    F_Async
    F_Generator
    F_AsyncGenerator
)


func LazyValueWrapper (expr string) string {
    return fmt.Sprintf("(() => (%v))", expr)
}


func VarLookup (tree Tree, ptr int) string {
    var file = GetFileName(tree)
    var row, col = GetRowColInfo(tree, ptr)
    return fmt.Sprintf (
        "%v(%v, [%v], %v, %v, %v)",
        G(CALL), L_VAR_LOOKUP,
        EscapeRawString(GetTokenContent(tree, ptr)), file, row, col,
    )
}


func EscapeRawString (raw []rune) string {
    // example: ['a', '"', 'b', 'c', '\', 'n'] -> `"a\"bc\\n"`
    // Containers["hash"] requires this function to be consistent when
    //     checking duplicate keys.
    var buf strings.Builder
    buf.WriteRune('"')
    for _, char := range raw {
        if char == '\\' {
            buf.WriteString(`\\`)
        } else if char == '"' {
            buf.WriteString(`\"`)
        } else if char == '\n' {
            buf.WriteString(`\n`)
        } else {
            buf.WriteRune(char)
        }
    }
    buf.WriteRune('"')
    return buf.String()
}


func NotEmpty (tree Tree, ptr int) bool {
    return tree.Nodes[ptr].Length > 0
}


func Empty (tree Tree, ptr int) bool {
    return !NotEmpty(tree, ptr)
}


func FlatSubTree (tree Tree, ptr int, extract string, next string) []int {
    var sequence = make([]int, 0)
    for NotEmpty(tree, ptr) {
        var children = Children(tree, ptr)
        var extract_ptr, exists = children[extract]
        if !exists { panic("cannot extract part " + next) }
        sequence = append(sequence, extract_ptr)
        ptr, exists = children[next]
        if !exists { panic("next part " + next + " not found") }
    }
    return sequence
}


func TranspileSubTree (tree Tree, ptr int, item string, next string) string {
    var item_ptrs = FlatSubTree(tree, ptr, item, next)
    var buf strings.Builder
    buf.WriteRune('[')
    for i, item_ptr := range item_ptrs {
        buf.WriteString(Transpile(tree, item_ptr))
        if i != len(item_ptrs)-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
    return buf.String()
}


func GetTokenContent (tree Tree, ptr int) []rune {
    var node = &tree.Nodes[ptr]
    if node.Part.Partype == syntax.Recursive {
        node = &tree.Nodes[node.Children[0]]
    }
    if node.Part.Partype != syntax.MatchToken {
        panic("trying to get token content of non-token node")
    }
    return tree.Tokens[node.Pos].Content
}


func GetWholeContent (tree Tree, ptr int) []rune {
    if NotEmpty(tree, ptr) {
        var pos = tree.Nodes[ptr].Pos
        var amount = tree.Nodes[ptr].Amount
        var begin_token = tree.Tokens[pos]
        var end_token = tree.Tokens[pos+amount-1]
        return tree.Code[begin_token.Pos : end_token.Pos+len(end_token.Content)]
    } else {
        return []rune("")
    }
}


func GetOperatorInfo (tree Tree, ptr int) syntax.Operator {
    var Id = tree.Nodes[ptr].Part.Id
    var OpId = syntax.Name2Id["operator"]
    var SopId = syntax.Name2Id["set_op"]
    if Id != OpId && Id != SopId {
        panic("unable to get operator info of non-operator")
    }
    var group_ptr = tree.Nodes[ptr].Children[0]
    var group = &tree.Nodes[group_ptr]
    var token_node = &tree.Nodes[group.Children[0]]
    var op_id = token_node.Part.Id
    var info, exists = syntax.Id2Operator[op_id]
    if !exists {
        panic("undefined operator " + syntax.Id2Name[op_id])
    }
    return info
}


func GetGeneralOperatorName (tree Tree, ptr int) (string, bool) {
    var Id = tree.Nodes[ptr].Part.Id
    if Id != syntax.Name2Id["general_op"] {
        panic("invalid usage of GetGeneralOperatorName()")
    }
    // general_op = operator | unary
    var children = Children(tree, ptr)
    var infix, is_infix = children["operator"]
    if is_infix {
        var info = GetOperatorInfo(tree, infix)
        return strings.TrimPrefix(info.Match, "@"), info.CanRedef
    } else {
        var unary = children["unary"]
        var group = tree.Nodes[unary].Children[0]
        var child = tree.Nodes[group].Children[0]
        var match = syntax.Id2Name[tree.Nodes[child].Part.Id]
        var can_redef = false
        for _, redefinable := range syntax.RedefinableOperators {
            if match == redefinable {
                can_redef = true
                break
            }
        }
        return strings.TrimPrefix(match, "@"), can_redef
    }
}


func WriteHelpers (buf *strings.Builder, scope_name string) {
    fmt.Fprintf (
        buf,
        "let {%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v} = %v.%v(%v); ",
        L_METHOD_CALL, L_VAR_LOOKUP, L_VAR_DECL, L_VAR_RESET,
        L_ADD_FUN, L_OP_MOUNT, L_STATIC_SCOPE, L_WRAP,
        L_IMPORT_VAR, L_IMPORT_MOD, L_IMPORT_ALL,
        L_GLOBAL_HELPERS,
        RUNTIME, R_GET_HELPERS, scope_name,
    )
}


func BareFunction (content string) string {
    var buf strings.Builder
    fmt.Fprintf(&buf, "(function (%v) { ", SCOPE)
    WriteHelpers(&buf, SCOPE)
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func BareAsyncFunction (content string) string {
    var buf strings.Builder
    fmt.Fprintf(&buf, "(async function (%v) { ", SCOPE)
    WriteHelpers(&buf, SCOPE)
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func BareGenerator (content string) string {
    var buf strings.Builder
    fmt.Fprintf(&buf, "(function* (%v) { ", SCOPE)
    WriteHelpers(&buf, SCOPE)
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func BareAsyncGenerator (content string) string {
    var buf strings.Builder
    fmt.Fprintf(&buf, "(async function* (%v) { ", SCOPE)
    WriteHelpers(&buf, SCOPE)
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func Commands (tree Tree, ptr int, add_return bool) string {
    var commands = FlatSubTree(tree, ptr, "command", "commands")
    var ReturnId = syntax.Name2Id["cmd_return"]
    var has_return = false
    var prev_row = -1
    var buf strings.Builder
    for i, command := range commands {
        if !has_return {
            var group = tree.Nodes[command].Children[0]
            if group == 0 { panic("invalid command") }
            var real = tree.Nodes[group].Children[0]
            if real == 0 { panic("invalid command") }
            if tree.Nodes[real].Part.Id == ReturnId {
                has_return = true
            }
        }
        var node = &tree.Nodes[command]
        var token = tree.Tokens[node.Pos]
        var row = tree.Info[token.Pos].Row
        if row == prev_row && !tree.Semi[node.Pos] {
            parser.Error(tree, command, "semicolon expected")
        }
        prev_row = row
        buf.WriteString(Transpile(tree, command))
        if i != len(commands)-1 {
            buf.WriteString("; ")
        } else {
            buf.WriteString(";")
        }
    }
    if add_return && !has_return {
        // return Void
        fmt.Fprintf(&buf, " return %v;", G(T_VOID))
    }
    return buf.String()
}


func Function (
    tree Tree, body_ptr int, fun_type FunType,
    desc string, parameters string, value_type string,
) string {
    var body = Transpile(tree, body_ptr)
    var body_children = Children(tree, body_ptr)
    // static_commands? = @static { commands }
    var static_ptr = body_children["static_commands"]
    var static_scope = "null"
    if NotEmpty(tree, static_ptr) {
        var parent_ptr = tree.Nodes[body_ptr].Parent
        var parent_name = syntax.Id2Name[tree.Nodes[parent_ptr].Part.Id]
        if parent_name == "method" || parent_name == "method_implemented" {
            parser.Error (
                tree, static_ptr,
                "static block is not available in method definition",
            )
        }
        var static_commands_ptr = Children(tree, static_ptr)["commands"]
        var static_commands = Commands(tree, static_commands_ptr, true)
        var static_executor = BareFunction(static_commands)
        static_scope = fmt.Sprintf("%v(%v)", L_STATIC_SCOPE, static_executor)
    }
    var raw string
    switch fun_type {
    case F_Sync:
        raw = BareFunction(body)
    case F_Async:
        raw = BareAsyncFunction(body)
    case F_Generator:
        raw = BareGenerator(body)
    case F_AsyncGenerator:
        raw = BareAsyncGenerator(body)
    default:
        panic("invalid FunType")
    }
    return fmt.Sprintf(
        "%v(%v, %v, %v, %v)",
        L_WRAP,
        fmt.Sprintf(
            "{ parameters: %v, value_type: %v }",
            parameters, value_type,
        ),
        static_scope, desc, raw,
    )
}


func InitFunction (tree Tree, ptr int, name []rune) string {
    var children = Children(tree, ptr)
    var params_ptr = children["paralist_strict"]
    var parameters = Transpile(tree, params_ptr)
    var body_ptr = children["body"]
    var desc = Desc (
        name,
        GetWholeContent(tree, params_ptr),
        []rune("Instance"),
    )
    return Function (
        tree, body_ptr, F_Sync,
        desc, parameters, G(T_INSTANCE),
    )
}


func Desc (name []rune, parameters []rune, value_type []rune) string {
    var desc_buf = make([]rune, 0, 120)
    desc_buf = append(desc_buf, name...)
    desc_buf = append(desc_buf, ' ')
    if len(parameters) == 0 || parameters[0] != '(' {
        desc_buf = append(desc_buf, '(')
    }
    desc_buf = append(desc_buf, parameters...)
    if len(parameters) == 0 || parameters[len(parameters)-1] != ')' {
        desc_buf = append(desc_buf, ')')
    }
    desc_buf = append(desc_buf, []rune(" -> ")...)
    desc_buf = append(desc_buf, value_type...)
    return EscapeRawString(desc_buf)
}


func SearchDotParameters (tree Tree, search_root int) []string {
    var names = make([]string, 0, 25)
    var DotParaId = syntax.Name2Id["dot_para"]
    var LambdaId = syntax.Name2Id["lambda"]
    var IIFE_Id = syntax.Name2Id["iife"]
    var do_search func(int)
    do_search = func (ptr int) {
        var node = &tree.Nodes[ptr]
        var id = node.Part.Id
        if (id == LambdaId || id == IIFE_Id) && ptr != search_root {
            return
        } else if id == DotParaId {
            // dot_para = . Name
            var children = Children(tree, ptr)
            var name = GetTokenContent(tree, children["Name"])
            names = append(names, string(name))
        }
        for i := 0; i < node.Length; i++ {
            var child = node.Children[i]
            do_search(child)
        }
    }
    do_search(search_root)
    var name_set = make(map[string]bool)
    var normalized = make([]string, 0, 10)
    for _, name := range names {
        var _, exists = name_set[name]
        if !exists {
            name_set[name] = true
            normalized = append(normalized, name)
        }
    }
    return normalized
}


func UntypedParameters (names []string) string {
    var buf strings.Builder
    buf.WriteRune('[')
    for i, name := range names {
        fmt.Fprintf (
            &buf, "{ name: %v, type: %v }",
            EscapeRawString([]rune(name)),
            G(T_ANY),
        )
        if i != len(names)-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
    return buf.String()
}


func TypedParameterList (tree Tree, namelist_ptr int, type_ string) string {
    var name_ptrs = FlatSubTree(tree, namelist_ptr, "name", "namelist_tail")
    var occurred = make(map[string]bool)
    var buf strings.Builder
    buf.WriteRune('[')
    for i, name_ptr := range name_ptrs {
        var name = Transpile(tree, name_ptr)
        if occurred[name] {
            parser.Error (
                tree, name_ptr, fmt.Sprintf (
                    "duplicate parameter %v",
                    name,
                ),
            )
        }
        occurred[name] = true
        fmt.Fprintf(&buf, "{ name: %v, type: %v }", name, type_)
        if i != len(name_ptrs)-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
    return buf.String()
}


func UntypedParameterList (tree Tree, namelist_ptr int) string {
    return TypedParameterList(tree, namelist_ptr, G(T_ANY))
}


func GenericParameters (tree Tree, gp_ptr int, name []rune) (string, string) {
    var GpId = syntax.Name2Id["generic_params"]
    if tree.Nodes[gp_ptr].Part.Id != GpId || !NotEmpty(tree, gp_ptr) {
        panic("invalid usage of GenericParameters()")
    }
    var children = Children(tree, gp_ptr)
    var parameters string
    var desc string
    var typed_ptr, exists = children["typed_list"]
    if exists {
        parameters = Transpile(tree, typed_ptr)
        desc = Desc(name, GetWholeContent(tree, typed_ptr), []rune("Type"))
    } else {
        var l_ptr = children["namelist"]
        parameters = TypedParameterList(tree, l_ptr, G(T_TYPE))
        desc = Desc(name, GetWholeContent(tree, l_ptr), []rune("Type"))
    }
    return parameters, desc
}


func TypeTemplate (tree Tree, gp_ptr int, name_ptr int, expr string) string {
    var name_raw = GetTokenContent(tree, name_ptr)
    var parameters, desc = GenericParameters(tree, gp_ptr, name_raw)
    var raw = BareFunction(fmt.Sprintf("return %v;", expr))
    var proto = fmt.Sprintf (
        "{ parameters: %v, value_type: %v }",
        parameters, G(T_TYPE),
    )
    var f = fmt.Sprintf (
        "%v(%v, %v, %v, %v)",
        L_WRAP, proto, "null", desc, raw,
    )
    return fmt.Sprintf("%v(%v)", G(C_TEMPLATE), f)
}


func FieldList (tree Tree, ptr int) (string, string, string) {
    var FieldListId = syntax.Name2Id["field_list"]
    if tree.Nodes[ptr].Part.Id != FieldListId {
        panic("invalid usage of FieldList()")
    }
    var names = make([]string, 0, 16)
    var types = make([]string, 0, 16)
    var defaults = make([]string, 0, 16)
    var contains = make([]string, 0, 16)
    var names_hash = make(map[string]bool)
    // field_list = field field_list_tail
    var field_ptrs = FlatSubTree(tree, ptr, "field", "field_list_tail")
    for _, field_ptr := range field_ptrs {
        // field = name : type! field_default | @fields @of type!
        var children = Children(tree, field_ptr)
        var name_ptr, has_name = children["name"]
        if !has_name {
            contains = append(contains, Transpile(tree, children["type"]))
            continue
        }
        var name = Transpile(tree, name_ptr)
        var type_ = Transpile(tree, children["type"])
        var _, exists = names_hash[name]
        if exists {
            parser.Error (
                tree, field_ptr, fmt.Sprintf (
                    "duplicate schema field %v",
                    name,
                ),
            )
        }
        names_hash[name] = true
        var default_ string
        var default_ptr = children["field_default"]
        if NotEmpty(tree, default_ptr) {
            // field_default? = = expr
            default_ = TranspileLastChild(tree, default_ptr)
        } else {
            default_ = ""
        }
        names = append(names, name)
        types = append(types, type_)
        defaults = append(defaults, default_)
    }
    var field_items = make([]string, 0, 16)
    var default_items = make([]string, 0, 16)
    for i := 0; i < len(names); i++ {
        var field_item = fmt.Sprintf("%v: %v", names[i], types[i])
        field_items = append(field_items, field_item)
        if defaults[i] != "" {
            var default_item = fmt.Sprintf("%v: %v", names[i], defaults[i])
            default_items = append(default_items, default_item)
        }
    }
    var t = fmt.Sprintf("{ %v }", strings.Join(field_items, ", "))
    var d = fmt.Sprintf("{ %v }", strings.Join(default_items, ", "))
    var c = fmt.Sprintf("[%v]", strings.Join(contains, ", "))
    return t, d, c
}


func MethodTable (tree Tree, ptr int, extract string, next string) string {
    if NotEmpty(tree, ptr) {
        // argument 'extract' can be "method" or "pf"
        // note: the rule name "method" is depended by Function()
        var method_ptrs = FlatSubTree(tree, ptr, extract, next)
        var buf strings.Builder
        for i, method_ptr := range method_ptrs {
            var children = Children(tree, method_ptr)
            var name = Transpile(tree, children["name"])
            // call another rule function here
            var method = TransMapByName["f_sync"](tree, method_ptr)
            fmt.Fprintf(&buf, "{ name: %v, f: %v }", name, method)
            if i != len(method_ptrs)-1 {
                buf.WriteString(", ")
            }
        }
        return fmt.Sprintf("[ %v ]", buf.String())
    } else {
        return "[]"
    }
}


func Filters (tree Tree, exprlist_ptr int) string {
    var ExprListId = syntax.Name2Id["exprlist"]
    if tree.Nodes[exprlist_ptr].Part.Id != ExprListId {
        panic("invalid usage of Filters()")
    }
    var file = GetFileName(tree)
    var expr_ptrs = FlatSubTree(tree, exprlist_ptr, "expr", "exprlist_tail")
    var buf strings.Builder
    buf.WriteRune('(')
    for i, expr_ptr := range expr_ptrs {
        var row, col = GetRowColInfo(tree, expr_ptr)
        fmt.Fprintf (
            &buf, "%v(%v, [%v], %v, %v, %v)",
            G(CALL), G(REQ_BOOL), Transpile(tree, expr_ptr), file, row, col,
        )
        if i != len(expr_ptrs)-1 {
            buf.WriteString(" && ")
        }
    }
    buf.WriteRune(')')
    return buf.String()
}


func WriteList (buf *strings.Builder, strlist []string) {
    for i, item := range strlist {
        buf.WriteString(item)
        if i != len(strlist)-1 {
            buf.WriteString(", ")
        }
    }
}


func GetKey (tree Tree, ptr int) (string, string) {
    if tree.Nodes[ptr].Part.Id != syntax.Name2Id["get"] {
        panic("invalid usage of GetKey()")
    }
    // get = get_expr | get_name
    var params = Children(tree, tree.Nodes[ptr].Children[0])
    // get_expr = Get [ expr! ]! nil_flag
    // get_name = Get . name! nil_flag
    var nil_flag = Transpile(tree, params["nil_flag"])
    var _, is_get_expr = params["expr"]
    if is_get_expr {
        return Transpile(tree, params["expr"]), nil_flag
    } else {
        return Transpile(tree, params["name"]), nil_flag
    }
}


func ReduceExpression (operators []syntax.Operator) [][3]int {
    /**
     *  Reduce Expression using the Shunting Yard Algorithm
     *
     *  N = the number of operators
     *  input = [1, -1, 2, -2, ..., (N-1), -(N-1), N, 0]
     *          (positive: operand, negative: operator, 0: pusher)
     *  output = index stack of operands (pos: operand, neg: reduced_operand)
     *  temp = index stack of operators
     *  reduced = [[operand1, operand2, operator], ...]
     */
    var pusher = syntax.Operator { Priority: -1, Assoc: syntax.Left }
    var N = len(operators)
    var input = make([]int, 0, 2*N+1+1)
    var output = make([]int, 0, N+1)
    var temp = make([]int, 0, N+1)
    var reduced = make([][3]int, 0, N)
    /* Initialize the Input */
    for i := 0; i <= N; i++ {
        // add operand index (incremented by 1)
        input = append(input, i+1)
        if i < N {
            // add operator index (incremented by 1 and negated)
            input = append(input, -(i+1))
        }
    }
    // add pusher
    input = append(input, 0)
    /* Read the Input */
    for _, I := range input {
        if I > 0 {
            // positive index => operand, push it to output stack
            var operand_index = I-1
            output = append(output, operand_index)
        } else {
            // non-positive index => operator
            // this index will be -1 if I == 0 (operator is pusher)
            var operator_index = -I-1
            // read the operator stack
            for len(temp) > 0 {
                // there is an operator on the operator stack
                var this *syntax.Operator
                if operator_index >= 0 {
                    // index is non-negative => normal operator
                    this = &operators[operator_index]
                } else {
                    // index is -1 => pusher
                    this = &pusher
                }
                // get the dumped operator on the top of the stack
                var dumped_op_index = temp[len(temp)-1]
                var dumped = operators[dumped_op_index]
                // determine if we should reduce output by the dumped operator
                var should_reduce bool
                if (this.Assoc == syntax.Left) {
                    should_reduce = dumped.Priority >= this.Priority
                } else {
                    should_reduce = dumped.Priority > this.Priority
                }
                if should_reduce {
                    // reduce the output stack
                    temp = temp[0:len(temp)-1]
                    var operand1 = output[len(output)-2]
                    var operand2 = output[len(output)-1]
                    output = output[0:len(output)-2]
                    reduced = append(reduced, [3]int {
                        operand1, operand2, dumped_op_index,
                    })
                    var reduced_index = len(reduced)-1
                    output = append(output, -(reduced_index+1))
                } else {
                    // important: if we should not reduce, exit the loop
                    break
                }
            }
            // push the current operator to the operator stack
            temp = append(temp, operator_index)
        }
    }
    /* Return the Result */
    return reduced
}
