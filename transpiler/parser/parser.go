package parser

import "../syntax"
import "../scanner"


type RawTreeNodeStatus int
const (
    Initial  RawTreeNodeStatus  =  iota
    Pending
    BranchFailed
    Success
    Failed
)

type RawTreeNode struct {
    Part      syntax.Part
    Parent    int
    Children  [syntax.MAX_NUM_PARTS]int
    Length    int
    Status    RawTreeNodeStatus
    Tried     int
    Index     int
    Pos       int
    Amount    int
}

type RawTree = []RawTreeNode


func BuildRawTree (tokens scanner.TokenSequence) RawTree {
    var NameId = syntax.Name2Id["Name"]
    var RootId = syntax.Name2Id[syntax.RootName]
    var RootPart = syntax.Part {
        Id: RootId,
        Partype: syntax.Recursive,
        Required: true,
    }
    var Root = RawTreeNode {
        Part: RootPart,
        Parent: -1,
        Tried: 0,
        Status: Initial,
        Pos: 0,
    }
    var tree = make(RawTree, 0, 10000)
    tree = append(tree, Root)
    var ptr = 0
    for {
        var node = tree[ptr]
        if ptr == 0 && node.Status != Initial { break }
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
            if tokens[pos].Id == id {
                node.Status = Success
                node.Amount = 1
            } else {
                node.Status = Failed
            }
        case syntax.MatchKeyword:
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
        if node.Part.Required && node.Amount == 0 {
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
            var num_parts = len(next.Parts)
            var j = 0
            for i := num_parts-1; i >= 0; i-- {
                var part = next.Parts[i]
                tree = append(tree, RawTreeNode {
                    Part: part,
                    Parent: ptr,
                    Tried: 0,
                    Status: Initial,
                    Index: j,
                })
                j += 1
            }
            node.Length = 0
            ptr += num_parts
            tree[ptr].Pos = node.Pos
            node.Status = Pending
        case Failed:
            var parent = tree[node.Parent]
            parent.Status = BranchFailed
        case Success:
            var parent = tree[node.Parent]
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
                ptr = node.Parent
                tree[ptr].Status = Success
            }
        }
    }

    return tree
}
