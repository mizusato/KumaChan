package parser

import (
    "fmt"
    "kumachan/parser/ast"
)
import "kumachan/parser/syntax"
import . "kumachan/error"

func InternalError (msg string) {
    panic(fmt.Sprintf("Internal Parser Error: %v", msg))
}

type Error struct {
    // TODO: refactor: add reference to Tree
    HasExpectedPart  bool
    ExpectedPart     syntax.Id
    NodeIndex        int  // may be bigger than the index of last token
}

func (err *Error) Message() ErrorMessage {
    var msg = make(ErrorMessage, 0)
    if err.HasExpectedPart {
        msg.WriteText(TS_ERROR, "Syntax unit")
        msg.WriteInnerText(TS_INLINE, syntax.Id2Name[err.ExpectedPart])
        msg.WriteText(TS_ERROR, "expected")
    } else {
        msg.WriteText(TS_ERROR, "Parser stuck")
    }
    return msg
}

func (err *Error) DetailedMessage(tree *ast.Tree) ErrorMessage {
    var node = &tree.Nodes[err.NodeIndex]
    var token_index int
    var got string
    if node.Pos >= len(tree.Tokens) {
        token_index = len(tree.Tokens)-1
        got = "EOF"
    } else {
        token_index = node.Pos
        got = syntax.Id2Name[tree.Tokens[token_index].Id]
    }
    var token = &tree.Tokens[token_index]
    var token_span_size = token.Span.End - token.Span.Start
    for (token_index + 1) < len(tree.Tokens) && token_span_size == 0 {
        token_index += 1
        token = &tree.Tokens[token_index]
        token_span_size = token.Span.End - token.Span.Start
    }
    var point = tree.Info[token.Span.Start]
    var desc = err.Message()
    desc.Write(T_SPACE)
    desc.WriteText(TS_ERROR, "(got")
    desc.Write(T_SPACE)
    desc.WriteText(TS_INLINE, got)
    desc.WriteText(TS_ERROR, ")")
    return FormatError (
        tree.Code,  tree.Info,  tree.SpanMap,
        tree.Name,  point,      token.Span,
        ERR_FOV,    TS_SPOT,    desc, nil,
    )
}
