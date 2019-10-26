package transformer

import "strings"
import ."kumachan/transformer/node"

func String (tree Tree, parent Pointer) StringLiteral {
    var ptr = GetChildPointer(tree, parent)
    return StringLiteral {
        Node: GetNode(tree, ptr, nil),
        Value: strings.Trim(GetTokenContent(tree, ptr), `'`),
    }
}