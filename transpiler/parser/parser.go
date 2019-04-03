package parser

import "fmt"
import "strings"
import "strconv"
import "../syntax"
import "../scanner"


type NodeStatus int
const (
    Initial NodeStatus = iota
    Pending
    BranchFailed
    Success
    Failed
)

const M = syntax.MAX_NUM_PARTS
type RawTreeNode struct {
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

type RawTree = []RawTreeNode


func BuildRawTree (tokens scanner.TokenSequence) RawTree {
    var NameId = syntax.Name2Id["Name"]
    var RootId = syntax.Name2Id[syntax.RootName]
    var RootPart = syntax.Part {
        Id:        RootId,
        Partype:   syntax.Recursive,
        Required:  true,
    }
    var Root = RawTreeNode {
        Part:    RootPart,  Parent:  -1,
        Length:  0,         Status:  Initial,
        Tried:   0,         Index:   0,
        Pos:     0,         Amount:  0,
    }
    var tree = make(RawTree, 0, 100000)
    tree = append(tree, Root)
    var ptr = 0
    loop: for {
        // PrintRawTree(tree)
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
                tree = append(tree, RawTreeNode {
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
            var parent_ptr = node.Parent
            if parent_ptr < 0 { break loop }
            var parent = &tree[parent_ptr]
            parent.Children[parent.Length] = ptr
            parent.Length += 1
            if partype == syntax.Recursive {
                node.Amount = 0
                for i := 0; i < node.Length; i++ {
                    node.Amount += tree[node.Children[i]].Amount
                }
            }
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

    return tree
}

func PrintRawTreeNode (ptr int, node RawTreeNode) {
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
    fmt.Println("------------------------------")
    for i, n := range tree {
        PrintRawTreeNode(i, n)
    }
}
