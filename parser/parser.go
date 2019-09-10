package parser

import "fmt"
import "kumachan/parser/syntax"
import "kumachan/parser/scanner"


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
    Code    scanner.Code
    Tokens  scanner.TokenSequence
    Info    scanner.RowColInfo
    Semi    scanner.SemiInfo
    Nodes   BareTree
    File    string
    Mock    []string
}


func BuildBareTree (
    root syntax.Id,
    tokens scanner.TokenSequence,
) (BareTree, int, string) {
    var NameId = syntax.Name2Id["Name"]
    var CallId = syntax.Name2Id["Call"]
    var GetId = syntax.Name2Id["Get"]
    var RootId = root
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
            } else if token_id == CallId || token_id == GetId {
                if tokens[node.Pos+1].Id == id {
                    node.Status = Success
                    node.Amount = 2
                } else {
                    node.Status = Failed
                }
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
            InternalError("invalid part type")
        }
        // if node is in final status
        if node.Status == Success || node.Status == Failed {
            // if partype is Recursive, empty match <=> node.Length == 0
            // if partype is otherwise, empty match <=> node.Amount == 0
            // if node.part is required, it should not be empty
            if node.Part.Required && node.Length == 0 && node.Amount == 0 {
                PrintBareTree(tree)
                return tree, ptr, fmt.Sprintf (
                    "error parsing code: syntax unit '%v' expected",
                    syntax.Id2Name[id],
                )
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
            InternalError("invalid status")
        }
    }
    // check if all the tokens have been matched
    var root_node = tree[0]
    if root_node.Amount < len(tokens) {
        PrintBareTree(tree)
        return tree, ptr, "error parsing code: parser stuck"
    }
    return tree, -1, ""
}


func BuildTree (root string, code scanner.Code, file_name string) Tree {
    var tokens, info, semi = scanner.Scan(code)
    var RootId, exists = syntax.Name2Id[root]
    if (!exists) {
        InternalError(fmt.Sprintf("invalid root syntax unit '%v'", root))
    }
    var nodes, err_ptr, err_desc = BuildBareTree(RootId, tokens)
    var tree = Tree {
        Code: code, Tokens: tokens, Info: info, Semi: semi,
        Nodes: nodes, File: file_name,
    }
    if err_ptr != -1 {
        Error(&tree, err_ptr, err_desc)
    }
    return tree
}
