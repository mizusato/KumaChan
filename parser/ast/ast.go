package ast

import (
	"kumachan/parser/scanner"
	"kumachan/parser/syntax"
)


const _MAX = syntax.MAX_NUM_PARTS

type Tree struct {
	Name     string
	Nodes    []TreeNode
	Code     scanner.Code
	Tokens   scanner.Tokens
	Info     scanner.RowColInfo
	SpanMap  scanner.RowSpanMap
}

type TreeNode struct {
	Part     syntax.Part  // { Id, PartType, Required }
	Parent   int          // pointer of parent node
	Children [_MAX]int    // pointers of children
	Length   int          // number of children
	Status   NodeStatus   // current status
	Tried    int          // number of tried branches
	Index    int          // index of the Part in the branch (reversed)
	Pos      int          // beginning position in Tokens
	Amount   int          // number of tokens that matched by the node
	Span     scanner.Span // spanning interval in code (rune list)
}

type NodeStatus int
const (
	Initial NodeStatus = iota
	Pending
	BranchFailed
	Success
	Failed
)
