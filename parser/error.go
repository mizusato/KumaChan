package parser

import "fmt"
import "kumachan/parser/syntax"
import . "kumachan/error"

func InternalError (msg string) {
    panic(fmt.Sprintf("Internal Parser Error: %v", msg))
}

type Error struct {
    HasExpectedPart  bool
    ExpectedPart     syntax.Id
    NodeIndex        int  // may be bigger than the index of last token
}

func (err *Error) Message() ErrorMessage {
    var msg = make(ErrorMessage, 0)
    if err.HasExpectedPart {
        msg.WriteText(TS_ERROR, "Syntax unit")
        msg.WriteInnerText(TS_BOLD, syntax.Id2Name[err.ExpectedPart])
        msg.WriteText(TS_ERROR, "expected")
    } else {
        msg.WriteText(TS_ERROR, "Parser stuck")
    }
    return msg
}

func (err *Error) DetailedMessage(tree *Tree) ErrorMessage {
    var token_index int
    var token = &tree.Tokens[token_index]
    var token_span_size = token.Span.End - token.Span.Start
    for (token_index + 1) < len(tree.Tokens) && token_span_size == 0 {
        token_index += 1
        token = &tree.Tokens[token_index]
        token_span_size = token.Span.End - token.Span.Start
    }
    var point = tree.Info[token.Span.Start]
    var desc = err.Message()
    return FormatError (
        tree.Code,  tree.Info,  tree.SpanMap,
        tree.Name,  point,      token.Span,
        ERR_FOV,    TS_ERROR,   desc, nil,
    )
}
