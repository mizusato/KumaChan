package cst

import (
	"fmt"
	"kumachan/interpreter/lang/textual/scanner"
	"kumachan/interpreter/lang/textual/syntax"
	"kumachan/interpreter/lang/common/source"
	"kumachan/standalone/util/richtext"
)


const _MAX = syntax.MAX_NUM_PARTS

type Tree struct {
	Name     string
	Nodes    [] TreeNode
	Code     scanner.Code
	Tokens   scanner.Tokens
	Info     scanner.RowColInfo
	SpanMap  scanner.RowSpanMap
}

type TreeNode struct {
	Part      syntax.Part   // { Id, PartType, Required }
	Parent    int           // pointer of parent node
	Children  [_MAX] int    // pointers of children
	Length    int           // number of children
	Status    NodeStatus    // current status
	Tried     int           // number of tried branches
	Index     int           // index of the Part in the branch (reversed)
	Pos       int           // beginning position in Tokens
	Amount    int           // number of tokens that matched by the node
	Span      scanner.Span  // spanning interval in code (rune list)
}

type NodeStatus int
const (
	Initial NodeStatus = iota
	Pending
	BranchFailed
	Success
	Failed
)

func (tree *Tree) GetPath() string {
	return tree.Name
}
func (tree *Tree) DescribePosition(pos source.Position) string {
	var start = uint(pos.StartTokenIndex)
	if start < uint(len(tree.Info)) {
		var point = tree.Info[start]
		return fmt.Sprintf("(row %d, column %d)", point.Row, point.Col)
	} else {
		return fmt.Sprintf("(token %d)", start)
	}
}
func (tree *Tree) FormatMessage(pos source.Position, desc richtext.Block) richtext.Text {
	var span = scanner.Span {
		Start: int(pos.StartTokenIndex),
		End:   int(pos.EndTokenIndex),
	}
	var point scanner.Point
	if span.Start < len(tree.Info) {
		point = tree.Info[span.Start]
	}
	const FOV = 5
	return FormatError (
		tree.Code, tree.Info, tree.SpanMap,
		point, span, FOV, desc,
	)
}

func GetNodeFirstToken(tree *Tree, index int) scanner.Token {
	var node = tree.Nodes[index]
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
	return token
}


