package parser

import "fmt"
import "kumachan/parser/cst"
import "kumachan/parser/syntax"
import "kumachan/parser/scanner"


/**
 *  Universal LL Parser
 *
 *  This parser can parse some Unicode text into a Concrete Syntax Tree
 *    according to the syntax defined in the subpackage `syntax`.
 *  In this file, the `buildTree()` function describes the
 *    core logic of the parser; the `Parse()` function is intended
 *    to be used outside the package as an API.
 *  Note that the CST generated by this parser is in a raw form,
 *    which means further transform of the CST (into AST) is required.
 */

func buildTree(root syntax.Id, tokens scanner.Tokens) ([]cst.TreeNode, *Error) {
    /**
     *  This function performs a top-down derivation on the tokens.
     *
     *  A stack (considered as a raw CST at same time) is used
     *    to record state instead of a recursion process.
     *  Although this function does not contain too much code,
     *    its logic is extraordinarily complicated.
     *  Therefore, be careful when modifying the following code.
     */
    var Name = syntax.Name2Id[syntax.IdentifierPartName]
    var RootId = root
    var RootPart = syntax.Part {
        Id:       RootId,
        PartType: syntax.Recursive,
        Required: true,
    }
    var ZeroSpan = scanner.Span { Start: 0, End: 0 }
    // prepare the stack (as well as tree), push the root node
    var tree = make([]cst.TreeNode, 0)
    tree = append(tree, cst.TreeNode {
        Part:    RootPart,  Parent:  -1,
        Length:  0,         Status:  cst.Initial,
        Tried:   0,         Index:   0,
        Pos:     0,         Amount:  0,
        Span:    ZeroSpan,
    })
    // set node pointer (current index) to zero and start looping
    var ptr = 0
    loop: for {
        var node = &tree[ptr]
        var id = node.Part.Id
        var part_type = node.Part.PartType
        switch part_type {
        case syntax.Recursive:
            if node.Status == cst.Initial {
                node.Status = cst.BranchFailed
            }
            // derivation through a branch failed
            if node.Status == cst.BranchFailed {
                var rule = syntax.Rules[id]
                var num_branches = len(rule.Branches)
                // check if all branches have been tried
                if node.Tried == num_branches {
                    // if tried, switch to a final status
                    if rule.Nullable {
                        node.Status = cst.Success
                        node.Length = 0
                    } else {
                        node.Status = cst.Failed
                    }
                }
            }
        case syntax.MatchToken:
            if node.Pos >= len(tokens) { node.Status = cst.Failed; break }
            var token_id = tokens[node.Pos].Id
            if token_id == id {
                node.Status = cst.Success
                node.Amount = 1
                node.Span = tokens[node.Pos].Span
            } else {
                node.Status = cst.Failed
            }
        case syntax.MatchKeyword:
            if node.Pos >= len(tokens) { node.Status = cst.Failed; break }
            var token = tokens[node.Pos]
            var token_id = token.Id
            if token_id != Name { node.Status = cst.Failed; break }
            var text = token.Content
            var keyword = syntax.Id2ConditionalKeyword[id]
            if len(text) != len(keyword) { node.Status = cst.Failed; break }
            var equal = true
            for i, char := range keyword {
                if char != text[i] {
                    equal = false
                    break
                }
            }
            if equal {
                node.Status = cst.Success
                node.Amount = 1
                node.Span = token.Span
            } else {
                node.Status = cst.Failed
            }
        default:
            panic("invalid part type")
        }
        // if node is in final status
        if node.Status == cst.Success || node.Status == cst.Failed {
            // if part_type is Recursive, empty match <=> node.Length == 0
            // if part_type is otherwise, empty match <=> node.Amount == 0
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
        case cst.BranchFailed:
            // status == BranchFailed  =>  part_type == Recursive
            var rule = syntax.Rules[id]
            var next = rule.Branches[node.Tried]
            node.Tried += 1   // increment the number of tried branches
            node.Length = 0   // clear invalid children
            // derive through the next branch
            var num_parts = len(next.Parts)
            var j = 0
            for i := num_parts-1; i >= 0; i-- {
                var part = next.Parts[i]
                tree = append(tree, cst.TreeNode {
                    Part:    part,   Parent:  ptr,
                    Length:  0,      Status:  cst.Initial,
                    Tried:   0,      Index:   j,
                    Pos:     -1,     Amount:  0,
                    Span:    ZeroSpan,
                })
                j += 1
            }
            ptr = len(tree) - 1
            tree[ptr].Pos = node.Pos
            node.Status = cst.Pending
        case cst.Failed:
            var parent_ptr = node.Parent
            if parent_ptr < 0 { break loop }
            var parent = &tree[parent_ptr]
            parent.Status = cst.BranchFailed // notify failure to parent node
            tree = tree[0: ptr-(node.Index)] // clear invalid nodes
            ptr = parent_ptr                 // go back to parent node
        case cst.Success:
            if part_type == syntax.Recursive {
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
                tree[ptr].Status = cst.Success
            }
        default:
            panic("invalid status")
        }
    }
    // check if all the tokens have been matched
    var root_node = tree[0]
    if root_node.Amount < len(tokens) {
        // PrintBareTree(tree)
        var last_token_ptr = 0
        var last_token_pos = 0
        for i, node := range tree {
            switch node.Part.PartType {
            case syntax.MatchKeyword,
                 syntax.MatchToken:
                     if node.Pos > last_token_pos {
                         last_token_pos = node.Pos
                         last_token_ptr = i
                     }
            }
        }
        return tree, &Error {
            HasExpectedPart: false,
            NodeIndex:       last_token_ptr,
        }
    }
    return tree, nil
}

func Parse(code []rune, root string, name string) (*cst.Tree, *Error) {
    var tokens, info, span_map = scanner.Scan(code)
    var Root, exists = syntax.Name2Id[root]
    if (!exists) {
        panic(fmt.Sprintf("invalid root syntax unit '%v'", root))
    }
    var nodes, err = buildTree(Root, tokens)
    var tree = cst.Tree {
        Name: name,  Nodes:   nodes,
        Code: code,  Tokens:  tokens,
        Info: info,  SpanMap: span_map,
    }
    if err != nil {
        err.Tree = &tree
    }
    return &tree, err
}
