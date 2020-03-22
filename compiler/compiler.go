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

func CodeSingleInst (inst common.Instruction, info checker.ExprInfo) Code {
	return Code {
		InstSeq:   [] common.Instruction { inst },
		SourceMap: [] *node.Node { &(info.ErrorPoint.Node) },
	}
}

type Context struct {
	GlobalRefs  [] GlobalRef
	LocalScope  *Scope
}

func (ctx *Context) AppendDataRef(v common.DataValue) uint {
	var index = uint(len(ctx.GlobalRefs))
	ctx.GlobalRefs = append(ctx.GlobalRefs, RefData { v })
	return index
}

func (ctx *Context) AppendFunRef(ref checker.RefFunction) uint {
	var index = uint(len(ctx.GlobalRefs))
	ctx.GlobalRefs = append(ctx.GlobalRefs, RefFun(ref))
	return index
}

func (ctx *Context) AppendConstRef(ref checker.RefConstant) uint {
	var index = uint(len(ctx.GlobalRefs))
	ctx.GlobalRefs = append(ctx.GlobalRefs, RefConst(ref))
	return index
}

type GlobalRef interface { GlobalRef() }
func (impl RefData) GlobalRef() {}
type RefData struct {
	Value  common.DataValue
}
func (impl RefFun) GlobalRef() {}
type RefFun checker.RefFunction
func (impl RefConst) GlobalRef() {}
type RefConst  checker.RefConstant

type DataInteger  checker.IntLiteral
func (d DataInteger) ToValue() common.Value {
	return d.Value
}
type DataSmallInteger  checker.SmallIntLiteral
func (d DataSmallInteger) ToValue() common.Value {
	return d.Value
}
type DataFloat  checker.FloatLiteral
func (d DataFloat) ToValue() common.Value {
	return d.Value
}

type Scope struct {
	Bindings    [] Binding
	BindingMap  map[string] uint
}
type Binding struct {
	Name  string
	Used  bool
}

func Compile(expr checker.Expr, ctx *Context) Code {
	switch v := expr.Value.(type) {
	case checker.IntLiteral:
		var index = ctx.AppendDataRef(DataInteger(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.SmallIntLiteral:
		var index = ctx.AppendDataRef(DataSmallInteger(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.FloatLiteral:
		var index = ctx.AppendDataRef(DataFloat(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefFunction:
		var index = ctx.AppendFunRef(v)
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefConstant:
		var index = ctx.AppendConstRef(v)
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefLocal:
		panic("not implemented")  // TODO
	default:
		panic("unknown expression kind")
	}
}

func InstGlobalRef(index uint) common.Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = common.RegIndex(index)
	return common.Instruction {
		OpCode: common.GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= common.RegistryMaxSize {
		panic("global value index exceeded maximum registry capacity")
	}
}
