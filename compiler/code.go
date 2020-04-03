package compiler

import (
	ch "kumachan/checker"
	c "kumachan/runtime/common"
	"kumachan/transformer/ast"
)


type Code struct {
	InstSeq    [] c.Instruction
	SourceMap  [] *ast.Node
}

func CodeFrom(inst c.Instruction, info ch.ExprInfo) Code {
	return Code {
		InstSeq:   [] c.Instruction { inst },
		SourceMap: [] *ast.Node { &(info.ErrorPoint.Node) },
	}
}

func (code Code) Length() uint {
	return uint(len(code.InstSeq))
}


type CodeBuffer struct {
	Code  *Code
}

func MakeCodeBuffer() CodeBuffer {
	var code = &Code {
		InstSeq:   make([] c.Instruction, 0),
		SourceMap: make([] *ast.Node, 0),
	}
	return CodeBuffer { code }
}

func (buf CodeBuffer) Write(code Code) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	for _, inst := range code.InstSeq {
		if inst.OpCode == c.JIF || inst.OpCode == c.JMP {
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			*base = append(*base, c.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   c.Long(dest_addr),
			})
		} else {
			*base = append(*base, inst)
		}
	}
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}

func (buf CodeBuffer) WriteAbsolute(code Code) {
	buf.Code.InstSeq = append(buf.Code.InstSeq, code.InstSeq...)
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}

func (buf CodeBuffer) WriteBranch(code Code, tail_addr uint) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	var last_addr = (code.Length() - 1)
	for _, inst := range code.InstSeq {
		if inst.OpCode == c.JIF || inst.OpCode == c.JMP {
			var rel_dest_addr = uint(inst.Arg1)
			var abs_dest_addr = (rel_dest_addr + base_size)
			ValidateDestAddr(abs_dest_addr)
			if rel_dest_addr == last_addr {
				abs_dest_addr = tail_addr
			}
			*base = append(*base, c.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   c.Long(abs_dest_addr),
			})
		} else {
			*base = append(*base, inst)
		}
	}
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}

func (buf CodeBuffer) Collect() Code {
	var code = buf.Code
	buf.Code = nil
	return *code
}


func InstGlobalRef(index uint) c.Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = c.GlobalIndex(index)
	return c.Instruction {
		OpCode: c.GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstLocalRef(offset uint) c.Instruction {
	ValidateLocalOffset(offset)
	return c.Instruction {
		OpCode: c.LOAD,
		Arg0:   0,
		Arg1:   c.Long(offset),
	}
}

func InstStore(offset uint) c.Instruction {
	ValidateLocalOffset(offset)
	return c.Instruction {
		OpCode: c.STORE,
		Arg0:   0,
		Arg1:   c.Long(offset),
	}
}

func InstGet(index uint) c.Instruction {
	ValidateProductIndex(index)
	return c.Instruction {
		OpCode: c.GET,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstSet(index uint) c.Instruction {
	ValidateProductIndex(index)
	return c.Instruction {
		OpCode: c.SET,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstProduct(size uint) c.Instruction {
	ValidateProductSize(size)
	return c.Instruction {
		OpCode: c.PROD,
		Arg0:   c.Short(size),
		Arg1:   0,
	}
}

func InstArray(size uint) c.Instruction {
	ValidateArraySize(size)
	var a0, a1 = c.ArraySize(size)
	return c.Instruction {
		OpCode: c.ARRAY,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstSum(index uint) c.Instruction {
	ValidateSumIndex(index)
	return c.Instruction {
		OpCode: c.SUM,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstJumpIf(index uint, dest uint) c.Instruction {
	ValidateSumIndex(index)
	ValidateDestAddr(dest)
	return c.Instruction {
		OpCode: c.JIF,
		Arg0:   c.Short(index),
		Arg1:   c.Long(dest),
	}
}

func InstJump(dest uint, narrow bool) c.Instruction {
	ValidateDestAddr(dest)
	var narrow_flag uint
	if narrow {
		narrow_flag = 1
	} else {
		narrow_flag = 0
	}
	return c.Instruction {
		OpCode: c.JMP,
		Arg0:   c.Short(narrow_flag),
		Arg1:   c.Long(dest),
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= c.GlobalSlotMaxSize {
		panic("global value index exceeded maximum slot capacity")
	}
}

func ValidateLocalOffset(offset uint) {
	if offset >= c.LocalSlotMaxSize {
		panic("local binding offset exceeded maximum slot capacity")
	}
}

func ValidateDestAddr(addr uint) {
	if addr >= c.FunCodeMaxLength {
		panic("destination address exceeded limitation")
	}
}

func ValidateProductIndex(index uint) {
	if index >= c.ProductMaxSize {
		panic("value index exceeded maximum capacity of product type")
	}
}

func ValidateProductSize(size uint) {
	if size > c.ProductMaxSize {
		panic("given size exceeded maximum capacity of product type")
	}
}

func ValidateSumIndex(index uint) {
	if index >= c.SumMaxBranches {
		panic("given index exceeded maximum branch limit of sum type")
	}
}

func ValidateArraySize(size uint) {
	if size > c.ArrayMaxSize {
		panic("given size exceeded maximum capacity of array literal")
	}
}
