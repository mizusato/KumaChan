package transformer

import "fmt"
import "strconv"
import "strings"
import "kumachan/parser/syntax"
import "kumachan/parser"

type Tree = *parser.Tree
type Pointer = int
type Context = map[string]interface{}

type Node struct {
    Kind  NodeKind
}

type NodeKind int
const (
    NK_Module NodeKind = iota
    NK_Decls
    NK_Commands
    // ... TODO
)

type Transformer = func(Tree, Pointer, Context) *Node

var __Rules = []map[string]Transformer {
    // TODO
}
var __TransformMapByName = make(map[string]Transformer)
var __TransformMap = make(map[syntax.Id]Transformer)

func Transform (tree Tree, ptr Pointer, ctx Context) *Node {
    // hash returned by Children() is of type map[string]int,
    // which will return 0 if non-existing key requested.
    // so we use -1 to indicate root node instead
    if ptr == 0 {
        panic("invalid usage of Transform(): please use ptr=-1 for root node")
    }
    if ptr == -1 {
        ptr = 0
    }
    var id = tree.Nodes[ptr].Part.Id
    // fmt.Printf("Transform: %v\n", syntax.Id2Name[id])
    var f, exists = __TransformMap[id]
    if exists {
        return f(tree, ptr, ctx)
    } else {
        panic (
            fmt.Sprintf (
                "transform rule for %v does not exist",
                syntax.Id2Name[id],
            ),
        )
    }
}

func TransformFirstChild (tree Tree, ptr Pointer, ctx Context) *Node {
    var node = &tree.Nodes[ptr]
    if node.Length > 0 {
        return Transform(tree, node.Children[0], ctx)
    } else {
        parser.PrintTreeNode(ptr, node)
        panic("unable to transform first child: this node has no child")
    }
}

func TransformLastChild (tree Tree, ptr Pointer, ctx Context) *Node {
    var node = &tree.Nodes[ptr]
    if node.Length > 0 {
        return Transform(tree, node.Children[node.Length-1], ctx)
    } else {
        parser.PrintTreeNode(ptr, node)
        panic("unable to transform last child: this node has no child")
    }
}

func Children (tree Tree, ptr Pointer) map[string]int {
    var node = &tree.Nodes[ptr]
    var hash = make(map[string]int)
    for i := node.Length-1; i >= 0; i-- {
        // reversed loop: smaller index will override bigger index
        var child_ptr = node.Children[i]
        var name = syntax.Id2Name[tree.Nodes[child_ptr].Part.Id]
        hash[name] = child_ptr
    }
    return hash
}

func GetFileName (tree Tree) string {
    return EscapeRawString([]rune(tree.Name))
}

func GetRowColInfo (tree Tree, ptr int) (string, string) {
    var point = tree.Info[tree.Tokens[tree.Nodes[ptr].Pos].Span.Start]
    return strconv.Itoa(point.Row), strconv.Itoa(point.Col)
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


func NotEmpty (tree Tree, ptr Pointer) bool {
    return tree.Nodes[ptr].Length > 0
}


func Empty (tree Tree, ptr Pointer) bool {
    return !NotEmpty(tree, ptr)
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

func __ApplyRules () {
    for _, item := range __Rules {
        for key, value := range item {
            var _, exists = __TransformMapByName[key]
            if exists { panic("duplicate transform rule for " + key) }
            __TransformMapByName[key] = value
        }
    }
}

var __InitCalled = false

func Init () {
    if __InitCalled { return }; __InitCalled = true
    __ApplyRules()
    for name, value := range __TransformMapByName {
        __TransformMap[syntax.Name2Id[name]] = value
    }
}
