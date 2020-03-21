package compiler

import (
	"kumachan/checker"
	"kumachan/runtime/common"
	"kumachan/transformer/node"
)

type Code struct {
	InstSeq    [] common.Instruction
	SourceMap  [] *node.Node
}

type Context struct {

}

func Compile(expr checker.Expr, ctx Context) Code {
	return Code{}  // TODO
}
