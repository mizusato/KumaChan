package parser

import "kumachan/parser/cst"
import "kumachan/parser/syntax"
import . "kumachan/error"


type Error struct {
    Tree             *cst.Tree
    HasExpectedPart  bool
    ExpectedPart     syntax.Id
    NodeIndex        int  // may be bigger than the index of last token
}

func (err *Error) Desc() ErrorMessage {
    var node = err.Tree.Nodes[err.NodeIndex]
    var got string
    if node.Pos >= len(err.Tree.Tokens) {
        got = "EOF"
    } else {
        got = syntax.Id2Name(err.Tree.Tokens[node.Pos].Id)
    }
    var desc = make(ErrorMessage, 0)
    if err.HasExpectedPart {
        desc.WriteText(TS_ERROR, "Syntax unit")
        desc.WriteInnerText(TS_INLINE, syntax.Id2Name(err.ExpectedPart))
        desc.WriteText(TS_ERROR, "expected")
    } else {
        desc.WriteText(TS_ERROR, "Parser stuck")
    }
    desc.Write(T_SPACE)
    desc.WriteText(TS_ERROR, "(got")
    desc.Write(T_SPACE)
    desc.WriteText(TS_INLINE, got)
    desc.WriteText(TS_ERROR, ")")
    return desc
}

func (err *Error) Message() ErrorMessage {
    var tree = err.Tree
    var token = cst.GetNodeFirstToken(tree, err.NodeIndex)
    var point = tree.Info[token.Span.Start]
    var desc = err.Desc()
    return FormatError (
        tree.Code,  tree.Info,  tree.SpanMap,
        tree.Name,  point,      token.Span,
        ERR_FOV,    TS_SPOT,    desc,
    )
}
