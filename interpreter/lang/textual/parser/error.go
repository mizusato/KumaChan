package parser

import (
    . "kumachan/standalone/util/error"
    "kumachan/standalone/util/richtext"
    "kumachan/interpreter/lang/textual/cst"
    "kumachan/interpreter/lang/textual/syntax"
)


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

func (err *Error) Desc() richtext.Block {
    if err.IsScannerError {
        var desc richtext.Block
        desc.WriteLine(err.ScannerError.Error(), richtext.TAG_ERR_NORMAL)
        return desc
    }
    if err.IsEmptyTree() {
        var desc richtext.Block
        desc.WriteLine("empty input", richtext.TAG_ERR_NORMAL)
        return desc
    }
    var node = err.Tree.Nodes[err.NodeIndex]
    var got string
    if node.Pos >= len(err.Tree.Tokens) {
        got = "EOF"
    } else {
        got = syntax.Id2Name(err.Tree.Tokens[node.Pos].Id)
    }
    var desc richtext.Block
    if err.HasExpectedPart {
        desc.WriteSpan("Syntax unit", richtext.TAG_ERR_NORMAL)
        desc.WriteSpan(syntax.Id2Name(err.ExpectedPart), richtext.TAG_ERR_INLINE)
        desc.WriteSpan("expected", richtext.TAG_ERR_NORMAL)
    } else {
        desc.WriteSpan("Parser stuck", richtext.TAG_ERR_NORMAL)
    }
    desc.WriteSpan("(", richtext.TAG_ERR_NORMAL)
    desc.WriteSpan("got", richtext.TAG_ERR_NORMAL)
    desc.WriteSpan(got, richtext.TAG_ERR_INLINE)
    desc.WriteSpan(")", richtext.TAG_ERR_NORMAL)
    return desc
}

func (err *Error) Message() richtext.Text {
    if err.IsScannerError {
        return err.Desc().ToText()
    }
    if err.IsEmptyTree() {
        return err.Desc().ToText()
    }
    var tree = err.Tree
    var token = cst.GetNodeFirstToken(tree, err.NodeIndex)
    var point = tree.Info[token.Span.Start]
    var desc = err.Desc()
    const FOV = 5
    return cst.FormatError (
        tree.Code, tree.Info, tree.SpanMap,
        point, token.Span, FOV, desc,
    )
}


