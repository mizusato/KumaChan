package transformer

import (
    "fmt"
    "reflect"
    "runtime"
    "unsafe"
)
import "strings"
import "kumachan/parser"
import "kumachan/parser/syntax"
import ."kumachan/transformer/node"


type Tree = *parser.Tree
type Pointer = int
type Context = map[string]interface{}
type Transformer = func(Tree, Pointer) reflect.Value

func Transform (tree Tree) Module {
    var dive func(Tree, Pointer, []syntax.Id) (Pointer, bool)
    dive = func (tree Tree, ptr Pointer, path []syntax.Id) (Pointer, bool) {
        if len(path) == 0 {
            return ptr, true
        }
        var parser_node = &tree.Nodes[ptr]
        var L = parser_node.Length
        for i := 0; i < L; i += 1 {
            var child_ptr = parser_node.Children[i]
            var child = &tree.Nodes[child_ptr]
            if child.Part.Id == path[0] {
                return dive(tree, child_ptr, path[1:])
            }
        }
        return -1, false
    }
    var transform Transformer
    transform = func (tree Tree, ptr Pointer) reflect.Value {
        var parser_node = &tree.Nodes[ptr]
        var info = GetNodeInfoById(parser_node.Part.Id)
        var node = reflect.New(info.Type)
        var meta = node.Elem().FieldByName("Node").Addr().Interface().(*Node)
        meta.Span = parser_node.Span
        meta.Point = tree.Info[meta.Span.Start]
        var transform_dived = func (child_info *NodeChildInfo, f Transformer) {
            var field_index = child_info.FieldIndex
            var field = info.Type.Field(field_index)
            var field_value = node.Elem().Field(field_index)
            var path = child_info.DivePath
            var dived_ptr, exists = dive(tree, ptr, path)
            if exists {
                field_value.Set(f(tree, dived_ptr))
            } else {
                if child_info.Fallback != -1 {
                    var fallback_id = child_info.Fallback
                    var L = parser_node.Length
                    for i := 0; i < L; i += 1 {
                        var child_ptr = parser_node.Children[i]
                        var child_parser_node = &tree.Nodes[child_ptr]
                        if child_parser_node.Part.Id == fallback_id {
                            var value = transform(tree, child_ptr)
                            if field.Type.Kind() == reflect.Slice {
                                var t = reflect.MakeSlice(field.Type, 1, 1)
                                t.Index(0).Set(value)
                                field_value.Set(t)
                            } else {
                                field_value.Set(value)
                            }
                            break
                        }
                    }
                } else if child_info.Optional {
                    if field.Type.Kind() == reflect.Slice {
                        var empty_slice = reflect.MakeSlice(field.Type, 0, 0)
                        field_value.Set(empty_slice)
                    } else if field.Type.Kind() == reflect.Bool {
                        field_value.Set(reflect.ValueOf(false))
                    }
                } else {
                    var path_strlist = make([]string, len(path))
                    for i, segment := range path {
                        path_strlist[i] = syntax.Id2Name[segment]
                    }
                    var path_str = strings.Join(path_strlist, ".")
                    panic(fmt.Sprintf (
                        "transform(): path `%v` cannot be found in `%v`",
                        path_str, syntax.Id2Name[parser_node.Part.Id],
                    ))
                }
            }
        }
        if info.First != -1 {
            if parser_node.Length > 0 {
                var field_value = node.Elem().Field(info.First)
                field_value.Set(transform(tree, parser_node.Children[0]))
            } else {
                panic(fmt.Sprintf (
                    "transform(): cannot get first child of empty node `%v`",
                    syntax.Id2Name[parser_node.Part.Id],
                ))
            }
        }
        if info.Last != -1 {
            var L = parser_node.Length
            if L > 0 {
                var field_value = node.Elem().Field(info.Last)
                field_value.Set(transform(tree, parser_node.Children[L-1]))
            } else {
                panic(fmt.Sprintf (
                    "transform(): cannot get last child of empty node `%v`",
                    syntax.Id2Name[parser_node.Part.Id],
                ))
            }
        }
        for _, child_info := range info.Children {
            transform_dived(&child_info, transform)
        }
        for _, child_info := range info.Strings {
            transform_dived (
                &child_info,
                func (tree Tree, dived_ptr Pointer) reflect.Value {
                    var dived_node = &tree.Nodes[dived_ptr]
                    if dived_node.Part.Partype == syntax.MatchToken {
                        var content = GetTokenContent(tree, dived_ptr)
                        return reflect.ValueOf(content)
                    } else {
                        panic(fmt.Sprintf (
                            "cannot get token content of non-token part %v",
                            syntax.Id2Name[dived_node.Part.Id],
                        ))
                    }
                },
            )
        }
        for _, child_info := range info.Options {
            transform_dived (
                &child_info,
                func (tree Tree, _ Pointer) reflect.Value {
                    return reflect.ValueOf(true)
                },
            )
        }
        for _, list_info := range info.Lists {
            transform_dived (
                (*NodeChildInfo)(unsafe.Pointer(&list_info)),
                func (tree Tree, dived_ptr Pointer) reflect.Value {
                    var item_id = list_info.ItemId
                    var tail_id = list_info.TailId
                    var item_ptrs = FlatSubTree(tree, dived_ptr, item_id, tail_id)
                    var field_index = list_info.FieldIndex
                    var field = info.Type.Field(field_index)
                    if field.Type.Kind() != reflect.Slice {
                        panic("cannot transform list to non-slice field")
                    }
                    var N = len(item_ptrs)
                    var slice = reflect.MakeSlice(field.Type, N, N)
                    for i, item_ptr := range item_ptrs {
                        slice.Index(i).Set(transform(tree, item_ptr))
                    }
                    return slice
                },
            )
        }
        return node.Elem()
    }
    return transform(tree, 0).Interface().(Module)
}

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

func FlatSubTree (
    tree     Tree,       ptr   Pointer,
    extract  syntax.Id,  next  syntax.Id,
) []Pointer {
    var sequence = make([]int, 0)
    for NotEmpty(tree, ptr) {
        var extract_ptr = -1
        var next_ptr = -1
        var node = &tree.Nodes[ptr]
        var L = node.Length
        for i := 0; i < L; i += 1 {
            var child_ptr = node.Children[i]
            var child = &tree.Nodes[child_ptr]
            if child.Part.Id == extract {
                extract_ptr = child_ptr
            }
            if child.Part.Id == next {
                next_ptr = child_ptr
            }
        }
        if extract_ptr == -1 {
            panic("cannot extract part " + syntax.Id2Name[extract])
        }
        sequence = append(sequence, extract_ptr)
        if next_ptr == -1 {
            panic("next part " + syntax.Id2Name[next] + " not found")
        }
        ptr = next_ptr
    }
    return sequence
}
/*
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
*/
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
    }
}

func GetFileName (tree Tree) string {
    return tree.Name
}

func GetTokenContent (tree Tree, ptr int) []rune {
    var node = &tree.Nodes[ptr]
    return tree.Tokens[node.Pos + node.Amount - 1].Content
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


func PrintNodeRecursively (
    buf *strings.Builder,
    node reflect.Value, name string, depth int, is_last []bool,
) {
    var T = node.Type()
    if T.Kind() == reflect.Interface {
        PrintNodeRecursively(buf, node.Elem(), name, depth, is_last)
        return
    }
    const INC = 2
    const SPACE = " "
    parser.Repeat(depth+1, func (i int) {
        if depth > 0 && i < depth {
            if is_last[i] {
                parser.Fill(buf, INC, "", SPACE)
            } else {
                parser.Fill(buf, INC, "│", SPACE)
            }
        } else {
            if is_last[depth] {
                parser.Fill(buf, INC, "└", "─")
            } else {
                parser.Fill(buf, INC, "├", "─")
            }
        }
    })
    if T.Kind() == reflect.Struct && T.NumField() > 0 {
        buf.WriteString("┬─")
    } else if T.Kind() == reflect.Slice && node.Len() > 0 && !(T.AssignableTo(reflect.TypeOf([]rune{}))) {
        // TODO: fixme: condition expression too long
        buf.WriteString("┬─")
    } else {
        buf.WriteString("──")
    }
    fmt.Fprintf(buf, "\033[1m\033[%vm", parser.GetANSIColor(depth))
    fmt.Fprintf(buf, "[%v] %v", name, T.String())
    fmt.Fprintf(buf, "\033[0m")
    fmt.Fprintf(buf, "\033[%vm", parser.GetANSIColor(depth))
    buf.WriteRune(' ')
    var newline = func () {
        fmt.Fprintf(buf, "\033[0m\n")
    }
    switch T.Kind() {
    case reflect.Slice:
        if T.AssignableTo(reflect.TypeOf([]rune{})) {
            fmt.Fprintf(buf, "'%v'", string(node.Interface().([]rune)))
            newline()
        } else {
            var L = node.Len()
            if L == 0 {
                buf.WriteString("(empty)")
            }
            fmt.Fprintf(buf, "\033[0m")
            buf.WriteRune('\n')
            for i := 0; i < L; i += 1 {
                var child = node.Index(i)
                is_last = append(is_last, i == L-1)
                PrintNodeRecursively (
                    buf, child, fmt.Sprintf("%d", i),
                    (depth + 1), is_last,
                )
                is_last = is_last[:len(is_last)-1]
            }
        }
    case reflect.Struct:
        var L = node.NumField()
        newline()
        for i := 0; i < L; i += 1 {
            var field = node.Field(i)
            var field_info = T.Field(i)
            if field_info.Name == "Node" {
                continue
            }
            is_last = append(is_last, i == L-1)
            PrintNodeRecursively (
                buf, field, field_info.Name,
                (depth + 1), is_last,
            )
            is_last = is_last[:len(is_last)-1]
        }
    default:
        newline()
    }
}


func PrintNode (node reflect.Value) {
    var buf strings.Builder
    var is_last = make([]bool, 0, 1000)
    is_last = append(is_last, true)
    PrintNodeRecursively(&buf, node, "Module", 0, is_last)
    fmt.Println(buf.String())
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
