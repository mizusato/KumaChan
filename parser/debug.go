package parser

import (
	"fmt"
	"kumachan/parser/cst"
)
import "strings"
import "unicode/utf8"
import "kumachan/parser/syntax"


func GetUtf8Length(s string) int {
    return utf8.RuneCountInString(s)
}

func GetANSIColor (n int) int {
    return 31 + n % 6
}

func Repeat (n int, f func(int)) {
    for i := 0; i < n; i++ {
        f(i)
    }
}

func Fill (buf *strings.Builder, n int, s string, blank string) {
    buf.WriteString(s)
    Repeat(n-GetUtf8Length(s), func (_ int) {
        buf.WriteString(blank)
    })
}


func PrintTreeNode (ptr int, node *cst.TreeNode) {
    var buf strings.Builder
    fmt.Fprintf(&buf, "\033[1m\033[%vm", GetANSIColor(ptr))
    fmt.Fprintf(&buf, "(%v)", ptr)
    fmt.Fprintf(&buf, "\033[0m")
    buf.WriteRune(' ')
    fmt.Fprintf(&buf, "\033[1m")
    buf.WriteString(syntax.Id2Name(node.Part.Id))
    fmt.Fprintf(&buf, "\033[0m")
    buf.WriteRune(' ')
    buf.WriteRune('[')
    for i := 0; i < node.Length; i++ {
        var child_ptr = node.Children[i]
        fmt.Fprintf(&buf, "\033[1m\033[%vm", GetANSIColor(child_ptr))
        fmt.Fprintf(&buf, "%v", child_ptr)
        fmt.Fprintf(&buf, "\033[0m")
        if i != node.Length-1 {
            buf.WriteString(", ")
        }
    }
    buf.WriteRune(']')
    fmt.Fprintf(
        &buf, " <\033[1m\033[%vm%v\033[0m> ",
        GetANSIColor(node.Parent), node.Parent,
    )
    fmt.Fprintf(
        &buf, "status=%v, tried=%v, index=%v, pos=%+v, amount=%v\n",
        node.Status, node.Tried, node.Index, node.Pos, node.Amount,
    )
    fmt.Print(buf.String())
}

func PrintBareTree (tree []cst.TreeNode) {
    for i := 0; i < len(tree); i++ {
        PrintTreeNode(i, &tree[i])
    }
}

func PrintTreeRecursively (
    buf *strings.Builder,
    tree *cst.Tree, ptr int, depth int, is_last []bool,
) {
    const INC = 2
    const SPACE = " "
    var node = &tree.Nodes[ptr]
    Repeat(depth+1, func (i int) {
        if depth > 0 && i < depth {
            if is_last[i] {
                Fill(buf, INC, "", SPACE)
            } else {
                Fill(buf, INC, "│", SPACE)
            }
        } else {
            if is_last[depth] {
                Fill(buf, INC, "└", "─")
            } else {
                Fill(buf, INC, "├", "─")
            }
        }
    })
    if node.Length > 0 {
        buf.WriteString("┬─")
    } else {
        buf.WriteString("──")
    }
    fmt.Fprintf(buf, "\033[1m\033[%vm", GetANSIColor(depth))
    fmt.Fprintf(buf, "[%v]", syntax.Id2Name(node.Part.Id))
    fmt.Fprintf(buf, "\033[0m")
    fmt.Fprintf(buf, "\033[%vm", GetANSIColor(depth))
    buf.WriteRune(' ')
    switch node.Part.PartType {
    case syntax.MatchToken:
        var token = tree.Tokens[node.Pos + node.Amount - 1]
        fmt.Fprintf(buf, "'%v'", string(token.Content))
        buf.WriteRune(' ')
        var point = tree.Info[token.Span.Start]
        fmt.Fprintf(buf, "at <%v,%v>", point.Row, point.Col)
        fmt.Fprintf(buf, "\033[0m")
        buf.WriteRune('\n')
    case syntax.MatchKeyword:
        fmt.Fprintf(buf, "\033[0m")
        buf.WriteRune('\n')
    case syntax.Recursive:
        if node.Length == 0 {
            buf.WriteString("(empty)")
        }
        fmt.Fprintf(buf, "\033[0m")
        buf.WriteRune('\n')
        for i := 0; i < node.Length; i++ {
            var child = node.Children[i]
            is_last = append(is_last, i == node.Length-1)
            PrintTreeRecursively(buf, tree, child, depth+1, is_last)
            is_last = is_last[0: len(is_last)-1]
        }
    }
}

func PrintTree (tree *cst.Tree) {
    var buf strings.Builder
    var is_last = make([]bool, 0, 1000)
    is_last = append(is_last, true)
    PrintTreeRecursively(&buf, tree, 0, 0, is_last)
    fmt.Println(buf.String())
}
