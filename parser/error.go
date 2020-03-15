package parser

import "kumachan/parser/ast"
import "kumachan/parser/syntax"
import . "kumachan/error"


type Error struct {
    Tree             *ast.Tree
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
        got = syntax.Id2Name[err.Tree.Tokens[node.Pos].Id]
    }
    var desc = make(ErrorMessage, 0)
    if err.HasExpectedPart {
        desc.WriteText(TS_ERROR, "Syntax unit")
        desc.WriteInnerText(TS_INLINE, syntax.Id2Name[err.ExpectedPart])
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
    var node = tree.Nodes[err.NodeIndex]
    var token_index int
    if node.Pos >= len(tree.Tokens) {
        token_index = len(tree.Tokens)-1
    } else {
        token_index = node.Pos
    }
    var token = tree.Tokens[token_index]
    var token_span_size = token.Span.End - token.Span.Start
    for (token_index + 1) < len(tree.Tokens) && token_span_size == 0 {
        token_index += 1
        token = tree.Tokens[token_index]
        token_span_size = token.Span.End - token.Span.Start
    }
    var point = tree.Info[token.Span.Start]
    var desc = err.Desc()
    return FormatError (
        tree.Code,  tree.Info,  tree.SpanMap,
        tree.Name,  point,      token.Span,
        ERR_FOV,    TS_SPOT,    desc,
    )
}
