package transpiler

import "fmt"
import "strings"
import "../syntax"


type FunType int
const (
    F_Sync FunType = iota
    F_Async
    F_Generator
)


func LazyValueWrapper (expr string) string {
    return fmt.Sprintf("(() => (%v))", expr)
}


func VarLookup (variable_name []rune) string {
    var buf strings.Builder
    buf.WriteString("id")
    buf.WriteRune('(')
    buf.WriteString(EscapeRawString(variable_name))
    buf.WriteRune(')')
    return buf.String()
}


func EscapeRawString (raw []rune) string {
    // example: ['a', '"', 'b', 'c', '\', 'n'] -> `"a\"bc\\n"`
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
    return syntax.Id2Operator[op_id]
}


func WriteHelpers (buf *strings.Builder, scope_name string) {
    fmt.Fprintf(
        buf,
        "let {m,id,dl,rt,df,gs,w,__} = %v.helpers(%v); ",
        Runtime, scope_name,
    )
}


func BareFunction (content string) string {
    var buf strings.Builder
    buf.WriteString("(function (scope, expose) { ")
    WriteHelpers(&buf, "scope")
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func BareAsyncFunction (content string) string {
    var buf strings.Builder
    buf.WriteString("(async function (scope) { ")
    WriteHelpers(&buf, "scope")
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func BareGenerator (content string) string {
    var buf strings.Builder
    buf.WriteString("(function* (scope) { ")
    WriteHelpers(&buf, "scope")
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func Commands (tree Tree, ptr int, add_return bool) string {
    var commands = FlatSubTree(tree, ptr, "command", "commands")
    var ReturnId = syntax.Name2Id["cmd_return"]
    var buf strings.Builder
    var has_return = false
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
        buf.WriteString(Transpile(tree, command))
        if i != len(commands)-1 {
            buf.WriteString("; ")
        } else {
            buf.WriteString(";")
        }
    }
    if add_return && !has_return {
        // return Void
        buf.WriteString(" return __.v;")
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
        var static_commands_ptr = Children(tree, static_ptr)["commands"]
        var static_commands = Commands(tree, static_commands_ptr, true)
        var static_executor = BareFunction(static_commands)
        static_scope = fmt.Sprintf("gs(%v)", static_executor)
    }
    var raw string
    switch fun_type {
    case F_Sync:
        raw = BareFunction(body)
    case F_Async:
        raw = fmt.Sprintf("__.aw(%v)", BareAsyncFunction(body))
    case F_Generator:
        raw = BareGenerator(body)
    default:
        panic("invalid FunType")
    }
    return fmt.Sprintf(
        "w(%v, %v, %v, %v)",
        fmt.Sprintf(
            "{ parameters: %v, value_type: %v }",
            parameters, value_type,
        ),
        static_scope, desc, raw,
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
    var do_search func(int)
    do_search = func (ptr int) {
        var node = &tree.Nodes[ptr]
        var id = node.Part.Id
        if id == LambdaId && ptr != search_root {
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
            &buf, "{ name: %v, type: __.a }",
            EscapeRawString([]rune(name)),
        )
        if i != len(names)-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
    return buf.String()
}


func UntypedParameterList (tree Tree, namelist_ptr int) string {
    var name_ptrs = FlatSubTree(tree, namelist_ptr, "name", "namelist_tail")
    var names = make([]string, 0, 10)
    var buf strings.Builder
    buf.WriteRune('[')
    for i, name_ptr := range name_ptrs {
        var name = Transpile(tree, name_ptr)
        fmt.Fprintf(&buf, "{ name: %v, type: __.a }", name)
        if i != len(names)-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
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
