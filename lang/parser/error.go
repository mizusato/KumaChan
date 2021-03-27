package parser

import "kumachan/lang/parser/cst"
import "kumachan/lang/parser/syntax"
import . "kumachan/misc/util/error"


type Error struct {
    IsScannerError   bool
    ScannerError     error
    Tree             *cst.Tree
    HasExpectedPart  bool
    ExpectedPart     syntax.Id
    NodeIndex        int  // may be bigger than the index of last token
}

func (err *Error) IsEmptyTree() bool {
    return err.NodeIndex < 0 || len(err.Tree.Tokens) == 0
}

func (err *Error) Desc() ErrorMessage {
    if err.IsScannerError {
        var desc = make(ErrorMessage, 0)
        desc.WriteText(TS_ERROR, err.ScannerError.Error())
        return desc
    }
    if err.IsEmptyTree() {
        var desc = make(ErrorMessage, 0)
        desc.WriteText(TS_ERROR, "empty input")
        return desc
    }
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
    if err.IsScannerError {
        return err.Desc()
    }
    if err.IsEmptyTree() {
        return err.Desc()
    }
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
