package transpiler

import "strings"
import "../syntax"


func LazyValueWrapper (expr string) string {
    var buf strings.Builder
    buf.WriteRune('(')
    buf.WriteString("() => (")
    buf.WriteString(expr)
    buf.WriteString(")")
    buf.WriteRune(')')
    return buf.String()
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
    for tree.Nodes[ptr].Length > 0 {
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


func GetOperatorInfo (tree Tree, ptr int) syntax.Operator {
    if tree.Nodes[ptr].Part.Id != syntax.Name2Id["operator"] {
        panic("unable to get operator info of non-operator")
    }
    var group_ptr = tree.Nodes[ptr].Children[0]
    var group = &tree.Nodes[group_ptr]
    var token_node = &tree.Nodes[group.Children[0]]
    var op_id = token_node.Part.Id
    return syntax.Id2Operator[op_id]
}


func WriteHelpers (buf *strings.Builder) {
    buf.WriteString("let {c,m,o,is,id,dl,rt,w,gv,v} = ")
    buf.WriteString(Runtime)
    buf.WriteString("helpers(scope)")
    buf.WriteString("; ")
}


func BareFunction (content string) string {
    var buf strings.Builder
    buf.WriteString("(function (scope, expose) { ")
    WriteHelpers(&buf)
    buf.WriteString(content)
    buf.WriteString(" })")
    return buf.String()
}


func Commands (tree Tree, ptr int, add_return bool) string {
    var commands = FlatSubTree(tree, ptr, "command", "commands")
    var ReturnId = syntax.Name2Id["cmd_return"]
    var buf strings.Builder
    var has_return = false
    for _, command := range commands {
        if !has_return && tree.Nodes[command].Part.Id == ReturnId {
            has_return = true
        }
        buf.WriteString(Transpile(tree, command))
        buf.WriteString("; ")
    }
    if add_return && !has_return {
        // return Void
        buf.WriteString("return v;")
    }
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
