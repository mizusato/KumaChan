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
    Tried     int           //  number of tried branches
    Index     int           //  index of the Part in the branch (reversed)
    Pos       int           //  beginning position in TokenSequence
    Amount    int           //  number of tokens that matched by the node
}

type BareTree = []TreeNode

type Tree struct {
    Tokens  scanner.TokenSequence
    Info    scanner.RowColInfo
    Nodes   BareTree
}


func BuildBareTree (tokens scanner.TokenSequence) BareTree {
    var NameId = syntax.Name2Id["Name"]
    var RootId = syntax.Name2Id[syntax.RootName]
    var RootPart = syntax.Part {
        Id:        RootId,
        Partype:   syntax.Recursive,
        Required:  true,
    }
    var tree = make(BareTree, 0, 100000)
    tree = append(tree, TreeNode {
        Part:    RootPart,  Parent:  -1,
        Length:  0,         Status:  Initial,
        Tried:   0,         Index:   0,
        Pos:     0,         Amount:  0,
    })
    var ptr = 0
    loop: for {
        /*
        fmt.Println("-------------------------------")
        PrintBareTree(tree)
        */
        var node = &tree[ptr]
        var id = node.Part.Id
        var partype = node.Part.Partype
        switch partype {
        case syntax.Recursive:
            if node.Status == Initial {
                node.Status = BranchFailed
            }
            // derivation through a branch failed
            if node.Status == BranchFailed {
                var rule = syntax.Rules[id]
                var num_branches = len(rule.Branches)
                // check if all branches have been tried
                if node.Tried == num_branches {
                    // if tried, switch to a final status
                    if rule.Emptable {
                        node.Status = Success
                        node.Length = 0
                    } else {
                        node.Status = Failed
                    }
                }
            }
        case syntax.MatchToken:
            if node.Pos >= len(tokens) { node.Status = Failed; break }
            if tokens[node.Pos].Id == id {
                node.Status = Success
                node.Amount = 1
            } else {
                node.Status = Failed
            }
        case syntax.MatchKeyword:
            if node.Pos >= len(tokens) { node.Status = Failed; break }
            if tokens[node.Pos].Id != NameId { node.Status = Failed; break }
            var token = tokens[node.Pos]
            var text = token.Content
            var keyword = syntax.Id2Keyword[id]
            if len(text) != len(keyword) { node.Status = Failed; break }
            var equal = true
            for i, char := range keyword {
                if char != text[i] {
                    equal = false
                    break
                }
            }
            if equal {
                node.Status = Success
                node.Amount = 1
            } else {
                node.Status = Failed
            }
        default:
            panic("invalid part type")
        }
        // if node is in final status
        if node.Status == Success || node.Status == Failed {
            // if partype is Recursive, empty match <=> node.Length == 0
            // if partype is otherwise, empty match <=> node.Amount == 0
            // if node.part is required, it should not be empty
            if node.Part.Required && node.Length == 0 && node.Amount == 0 {
                PrintBareTree(tree)
                panic(syntax.Id2Name[id] + " expected")
            }
        }
        switch node.Status {
        case BranchFailed:
            // status == BranchFailed  =>  partype == Recursive
            var rule = syntax.Rules[id]
            var next = rule.Branches[node.Tried]
            node.Tried += 1   // increment the number of tried branches
            node.Length = 0   // clear invalid children
            // derive through the next branch
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
            parent.Status = BranchFailed      // notify failure to parent node
            tree = tree[0: ptr-(node.Index)]  // clear invalid nodes
            ptr = parent_ptr  // go back to parent node
        case Success:
            if partype == syntax.Recursive {
                // calcuate the number of tokens matched by the node
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
                // if node.part is NOT the last part in the branch,
                // go to the node corresponding to the next part
                ptr -= 1
                tree[ptr].Pos = node.Pos + node.Amount
            } else {
                // if node.part is the last part in the branch
                // notify success to the parent node and go to it
                ptr = parent_ptr
                tree[ptr].Status = Success
            }
        default:
            panic("invalid status")
        }
    }
    // check if all the tokens have been matched
    var root_node = tree[0]
    if root_node.Amount < len(tokens) {
        PrintBareTree(tree)
        panic("parser stuck at " + strconv.Itoa(root_node.Amount))
    }
    return tree
}

func BuildTree (code scanner.Code) Tree {
    var tokens, info = scanner.Scan(code)
    var nodes = BuildBareTree(tokens)
    return Tree { Tokens: tokens, Info: info, Nodes: nodes }
}


func PrintTreeNode (ptr int, node TreeNode) {
    var children = make([]string, 0, 20)
    for i := 0; i < node.Length; i++ {
        children = append(children, strconv.Itoa(node.Children[i]))
    }
    var children_str = strings.Join(children, ", ")
    fmt.Printf(
        "(%v) %v [%v] parent=%v, status=%v, tried=%v, index=%v, pos=%+v, amount=%v\n",
        ptr, syntax.Id2Name[node.Part.Id], children_str,
        node.Parent, node.Status, node.Tried, node.Index, node.Pos, node.Amount,
    )
}

func PrintBareTree (tree BareTree) {
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
    var buf StrBuf
    var is_last = make([]bool, 0, 1000)
    is_last = append(is_last, true)
    PrintTreeRecursively(&buf, tree, 0, 0, is_last)
    fmt.Println(buf.String())
}
