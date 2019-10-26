package transformer

import "runtime"
import "strings"
import "kumachan/parser"
import "kumachan/parser/syntax"
import ."kumachan/transformer/node"


type Tree = *parser.Tree
type Pointer = int
type Context = map[string]interface{}

func Children (tree Tree, ptr Pointer) map[string]int {
    var node = &tree.Nodes[ptr]
    var hash = make(map[string]int)
    for i := node.Length-1; i >= 0; i -= 1 {
        // reversed loop: smaller index will override bigger index
        var child_ptr = node.Children[i]
        var name = syntax.Id2Name[tree.Nodes[child_ptr].Part.Id]
        hash[name] = child_ptr
    }
    return hash
}

func HasChild (name string, tree Tree, ptr Pointer) bool {
    var id = syntax.Name2Id[name]
    var node = &tree.Nodes[ptr]
    for i := 0; i < node.Length; i += 1 {
        var child_ptr = node.Children[i]
        if tree.Nodes[child_ptr].Part.Id == id {
            return true
        }
    }
    return false
}

func FirstLastChild (tree Tree, ptr Pointer) (Pointer, Pointer) {
    var node = &tree.Nodes[ptr]
    var first = node.Children[0]
    var last = node.Children[node.Length-1]
    return first, last
}

func FlatSubTree (tree Tree, ptr Pointer, extract string, next string) []int {
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

func GetChildPointer (tree Tree, parent Pointer) Pointer {
    var pc, _, _, _ = runtime.Caller(1)
    var raw_name = runtime.FuncForPC(pc).Name()
    var t = strings.Split(raw_name, ".")
    var name = strings.TrimRight(t[len(t)-1], "_")
    var id = syntax.Name2Id[name]
    var p = &tree.Nodes[parent]
    if p.Part.Id == id {
        return parent
    } else {
        for i := 0; i < p.Length; i += 1 {
            var child_ptr = p.Children[i]
            var child = &tree.Nodes[child_ptr]
            if child.Part.Id == id {
                return child_ptr
            }
        }
        return -1
    }
}

func GetNode (tree Tree, ptr Pointer, info interface{}) Node {
    return Node {
        Point: tree.Info[tree.Nodes[ptr].Pos],
        Span: tree.Nodes[ptr].Span,
        Info: info,
    }
}

func GetFileName (tree Tree) string {
    return tree.Name
}

func GetTokenContent (tree Tree, ptr int) string {
    return string(tree.Tokens[tree.Nodes[ptr].Pos].Content)
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

func NotEmpty (tree Tree, ptr Pointer) bool {
    return tree.Nodes[ptr].Length > 0
}

func Empty (tree Tree, ptr Pointer) bool {
    return !NotEmpty(tree, ptr)
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
