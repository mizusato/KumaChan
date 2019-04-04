package parser

import "fmt"
import "strings"
import "strconv"
import "unicode/utf8"
import "../syntax"
import "../scanner"

type StrBuf = strings.Builder

func strlen (s string) int {
    return utf8.RuneCountInString(s)
}


type NodeStatus int
const (
    Initial NodeStatus = iota
    Pending
    BranchFailed
    Success
    Failed
)

const M = syntax.MAX_NUM_PARTS
type TreeNode struct {
    Part      syntax.Part   //  { Id, Partype, Required }
    Parent    int           //  pointer of parent node
    Children  [M]int        //  pointers of children
    Length    int           //  number of children
    Status    NodeStatus    //  current status
    Tried     int           //  number of tried branch
    Index     int           //  index of the Part in the branch (reversed)
    Pos       int           //  position in TokenSequence
    Amount    int           //  number of tokens that matched by the node
}

type RawTree = []TreeNode

type Tree struct {
    Tokens  scanner.TokenSequence
    Info    scanner.RowColInfo
    Nodes   RawTree
}


func BuildRawTree (tokens scanner.TokenSequence) RawTree {
    var NameId = syntax.Name2Id["Name"]
    var RootId = syntax.Name2Id[syntax.RootName]
    var RootPart = syntax.Part {
        Id:        RootId,
        Partype:   syntax.Recursive,
        Required:  true,
    }
    var Root = TreeNode {
        Part:    RootPart,  Parent:  -1,
        Length:  0,         Status:  Initial,
        Tried:   0,         Index:   0,
        Pos:     0,         Amount:  0,
    }
    var tree = make(RawTree, 0, 100000)
    tree = append(tree, Root)
    var ptr = 0
    loop: for {
        /*
        fmt.Println("-------------------------------")
        PrintRawTree(tree)
        */
        var node = &tree[ptr]
        var id = node.Part.Id
        var partype = node.Part.Partype
        switch partype {
        case syntax.Recursive:
            if node.Status == Initial {
                node.Status = BranchFailed
            }
            if node.Status == BranchFailed {
                var rule = syntax.Rules[id]
                var num_branches = len(rule.Branches)
                if node.Tried == num_branches {
                    if rule.Emptable {
                        node.Status = Success
                    } else {
                        node.Status = Failed
                    }
                }
            }
        case syntax.MatchToken:
            var pos = node.Pos
            if pos >= len(tokens) { node.Status = Failed; break }
            if tokens[pos].Id == id {
                node.Status = Success
                node.Amount = 1
            } else {
                node.Status = Failed
            }
        case syntax.MatchKeyword:
            if node.Pos >= len(tokens) { node.Status = Failed; break }
            var keyword = syntax.Id2Keyword[id]
            var token = tokens[node.Pos]
            if token.Id != NameId {
                node.Status = Failed
                break
            }
            if len(token.Content) != len(keyword) {
                node.Status = Failed
                break
            }
            var equal = true
            for i, char := range keyword {
                if char != token.Content[i] {
                    equal = false
                }
            }
            if !equal {
                node.Status = Failed
            } else {
                node.Status = Success
                node.Amount = 1
            }
        default:
            panic("invalid part type")
        }
        if node.Part.Required && node.Length == 0 && node.Amount == 0 {
            if node.Status == Success || node.Status == Failed {
                // TODO
                PrintRawTree(tree)
                panic(syntax.Id2Name[id] + " expected")
            }
        }
        switch node.Status {
        case BranchFailed:
            var rule = syntax.Rules[id]
            var next = rule.Branches[node.Tried]
            node.Tried += 1
            node.Length = 0
            var num_parts = len(next.Parts)
            var j = 0
            for i := num_parts-1; i >= 0; i-- {
                var part = next.Parts[i]
                tree = append(tree, TreeNode {
                    Part:    part,   Parent:  ptr,
                    Length:  0,      Status:  Initial,
                    Tried:   0,      Index:   j,
                    Pos:     -1,     Amount:  0,
                })
                j += 1
            }
            ptr = len(tree) - 1
            tree[ptr].Pos = node.Pos
            node.Status = Pending
        case Failed:
            var parent_ptr = node.Parent
            if parent_ptr < 0 { break loop }
            var parent = &tree[parent_ptr]
            parent.Status = BranchFailed
            tree = tree[0: ptr-(node.Index)]
            ptr = parent_ptr
        case Success:
            if partype == syntax.Recursive {
                node.Amount = 0
                for i := 0; i < node.Length; i++ {
                    node.Amount += tree[node.Children[i]].Amount
                }
            }
            var parent_ptr = node.Parent
            if parent_ptr < 0 { break loop }
            var parent = &tree[parent_ptr]
            parent.Children[parent.Length] = ptr
            parent.Length += 1
            if node.Index > 0 {
                ptr -= 1
                tree[ptr].Pos = node.Pos + node.Amount
            } else {
                ptr = parent_ptr
                tree[ptr].Status = Success
            }
        default:
            panic("invalid status")
        }
    }
    /*
    if Root.Amount < len(tokens) {
        panic("parser stuck at " + strconv.Itoa(Root.Amount))
    }
    */
    return tree
}

func BuildTree (code scanner.Code) Tree {
    var tokens, info = scanner.Scan(code)
    var nodes = BuildRawTree(tokens)
    return Tree { Tokens: tokens, Info: info, Nodes: nodes }
}


func PrintTreeNode (ptr int, node TreeNode) {
    var children = make([]string, 0, 20)
    for i := 0; i < node.Length; i++ {
        children = append(children, strconv.Itoa(node.Children[i]))
    }
    var children_str = strings.Join(children, ", ")
    fmt.Printf(
        "(%v) %v [%v] parent=%v, status=%v, tried=%v, pos=%+v, amount=%v\n",
        ptr, syntax.Id2Name[node.Part.Id], children_str,
        node.Parent, node.Status, node.Tried, node.Pos, node.Amount,
    )
}

func PrintRawTree (tree RawTree) {
    for i, n := range tree {
        PrintTreeNode(i, n)
    }
}


func Repeat (n int, f func(int)) {
    for i := 0; i < n; i++ {
        f(i)
    }
}

func Fill (buf *StrBuf, n int, s string, blank string) {
    buf.WriteString(s)
    Repeat(n-strlen(s), func (_ int) {
        buf.WriteString(blank)
    })
}

func PrintTreeRecursively (
    buf *StrBuf, tree Tree, ptr int, depth int, is_last []bool,
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
    fmt.Fprintf(buf, "[%v]", syntax.Id2Name[node.Part.Id])
    buf.WriteRune(' ')
    switch node.Part.Partype {
    case syntax.MatchToken:
        var token = tree.Tokens[node.Pos]
        fmt.Fprintf(buf, "'%v'", string(token.Content))
        buf.WriteRune(' ')
        var point = tree.Info[tree.Tokens[node.Pos].Pos]
        fmt.Fprintf(buf, "at <%v,%v>", point.Row, point.Col)
        buf.WriteRune('\n')
    case syntax.MatchKeyword:
        buf.WriteRune('\n')
    case syntax.Recursive:
        if node.Length == 0 {
            buf.WriteString("(empty)")
        }
        buf.WriteRune('\n')
        for i := 0; i < node.Length; i++ {
            var child = node.Children[i]
            is_last = append(is_last, i == node.Length-1)
            PrintTreeRecursively(buf, tree, child, depth+1, is_last)
            is_last = is_last[0: len(is_last)-1]
        }
    }
}

func PrintTree (tree Tree) {
    var buf strings.Builder
    var is_last = make([]bool, 0, 1000)
    is_last = append(is_last, true)
    PrintTreeRecursively(&buf, tree, 0, 0, is_last)
    fmt.Println(buf.String())
}
