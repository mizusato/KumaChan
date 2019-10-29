package parser

import "fmt"
import "kumachan/parser/syntax"
import "kumachan/parser/scanner"


const CMAX = syntax.MAX_NUM_PARTS

type NodeStatus int
const (
    Initial NodeStatus = iota
    Pending
    BranchFailed
    Success
    Failed
)

type TreeNode struct {
    Part      syntax.Part   // { Id, Partype, Required }
    Parent    int           // pointer of parent node
    Children  [CMAX]int     // pointers of children
    Length    int           // number of children
    Status    NodeStatus    // current status
    Tried     int           // number of tried branches
    Index     int           // index of the Part in the branch (reversed)
    Pos       int           // beginning position in Tokens
    Amount    int           // number of tokens that matched by the node
    Span      scanner.Span  // spanning interval in code (rune list)
}

type Tree struct {
    Name    string
    Nodes   []TreeNode
    Code    scanner.Code
    Tokens  scanner.Tokens
    Info    scanner.RowColInfo
}


func BuildTree (root syntax.Id,  tokens scanner.Tokens) ([]TreeNode, *Error) {
    var Name = syntax.Name2Id["Name"]
    var NoLF = syntax.Name2Id["NoLF"]
    var LF = syntax.Name2Id["LF"]
    var RootId = root
    var RootPart = syntax.Part {
        Id:        RootId,
        Partype:   syntax.Recursive,
        Required:  true,
    }
    var ZeroSpan = scanner.Span { Start: 0, End: 0 }
    var tree = make([]TreeNode, 0, 100000)
    tree = append(tree, TreeNode {
        Part:    RootPart,  Parent:  -1,
        Length:  0,         Status:  Initial,
        Tried:   0,         Index:   0,
        Pos:     0,         Amount:  0,
        Span:    ZeroSpan,
    })
    var ptr = 0
    loop: for {
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
            var token_id = tokens[node.Pos].Id
            if token_id == id {
                node.Status = Success
                node.Amount = 1
                node.Span = tokens[node.Pos].Span
            } else if token_id == NoLF || token_id == LF {
                var next = node.Pos + 1
                if next < len(tokens) && tokens[next].Id == id {
                    node.Status = Success
                    node.Amount = 2
                    node.Span = tokens[node.Pos+1].Span
                } else {
                    node.Status = Failed
                }
            } else {
                node.Status = Failed
            }
        case syntax.MatchKeyword:
            if node.Pos >= len(tokens) { node.Status = Failed; break }
            var token = tokens[node.Pos]
            var token_id = token.Id
            var next = node.Pos + 1
            var use_next = false
            if next < len(tokens) && (token_id == NoLF || token_id == LF) {
                token = tokens[next]
                token_id = token.Id
                use_next = true
            }
            if token_id != Name { node.Status = Failed; break }
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
                if use_next {
                    node.Amount = 2
                } else {
                    node.Amount = 1
                }
                node.Span = token.Span
            } else {
                node.Status = Failed
            }
        default:
            InternalError("invalid part type")
        }
        // if node is in final status
        if node.Status == Success || node.Status == Failed {
            // if partype is Recursive, empty match <=> node.Length == 0
            // if partype is otherwise, empty match <=> node.Amount == 0
            // if node.part is required, it should not be empty
            if node.Part.Required && node.Length == 0 && node.Amount == 0 {
                // PrintBareTree(tree)
                return tree, &Error {
                    HasExpectedPart: true,
                    ExpectedPart:    id,
                    NodeIndex:       ptr,
                }
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
                    Span:    ZeroSpan,
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
                // calculate the number of tokens matched by the node
                node.Amount = 0
                for i := 0; i < node.Length; i++ {
                    var child = node.Children[i]
                    node.Amount += tree[child].Amount
                }
            }
            var parent_ptr = node.Parent
            if parent_ptr < 0 { break loop }
            var parent = &tree[parent_ptr]
            parent.Children[parent.Length] = ptr
            parent.Length += 1
            if parent.Span == ZeroSpan {
                parent.Span = node.Span
            } else if node.Span != ZeroSpan {
                parent.Span = parent.Span.Merged(node.Span)
            }
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
            InternalError("invalid status")
        }
    }
    // check if all the tokens have been matched
    var root_node = tree[0]
    if root_node.Amount < len(tokens) {
        // PrintBareTree(tree)
        return tree, &Error {
            HasExpectedPart: false,
            NodeIndex:       ptr,
        }
    }
    return tree, nil
}


func Parse (code []rune, root string, name string) (*Tree, *Error) {
    var tokens, info = scanner.Scan(code)
    var Root, exists = syntax.Name2Id[root]
    if (!exists) {
        InternalError(fmt.Sprintf("invalid root syntax unit '%v'", root))
    }
    var nodes, err = BuildTree(Root, tokens)
    var tree = Tree {
        Nodes: nodes, Name: name,
        Code: code, Tokens: tokens, Info: info,
    }
    return &tree, err
}
