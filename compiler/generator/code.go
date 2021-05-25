package generator

import (
	ch "kumachan/compiler/checker"
	. "kumachan/misc/util/error"
	"kumachan/lang"
)


type Code struct {
	InstSeq    [] lang.Instruction
	SourceMap  [] ErrorPoint
}

func CodeFrom(inst lang.Instruction, info ch.ExprInfo) Code {
	return Code {
		InstSeq:   [] lang.Instruction {inst },
		SourceMap: [] ErrorPoint { info.ErrorPoint },
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
		InstSeq:   make([] lang.Instruction, 0),
		SourceMap: make([] ErrorPoint, 0),
	}
	return CodeBuffer { code }
}

func (buf CodeBuffer) Write(code Code) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	for _, inst := range code.InstSeq {
		switch inst.OpCode {
		case lang.JIF, lang.JMP, lang.MSJ:
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			*base = append(*base, lang.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   lang.Long(dest_addr),
			})
		default:
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
		switch inst.OpCode {
		case lang.JIF, lang.JMP, lang.MSJ:
			var rel_dest_addr = uint(inst.Arg1)
			var abs_dest_addr = (rel_dest_addr + base_size)
			ValidateDestAddr(abs_dest_addr)
			if rel_dest_addr == last_addr {
				abs_dest_addr = tail_addr
			}
			*base = append(*base, lang.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   lang.Long(abs_dest_addr),
			})
		default:
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


func InstGlobalRef(index uint) lang.Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = lang.GlobalIndex(index)
	return lang.Instruction {
		OpCode: lang.GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstLocalRef(offset uint) lang.Instruction {
	ValidateLocalOffset(offset)
	return lang.Instruction {
		OpCode: lang.LOAD,
		Arg0:   0,
		Arg1:   lang.Long(offset),
	}
}

func InstStore(offset uint) lang.Instruction {
	ValidateLocalOffset(offset)
	return lang.Instruction {
		OpCode: lang.STORE,
		Arg0:   0,
		Arg1:   lang.Long(offset),
	}
}

func InstGet(index uint) lang.Instruction {
	ValidateProductIndex(index)
	return lang.Instruction {
		OpCode: lang.GET,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstPopGet(index uint) lang.Instruction {
	ValidateProductIndex(index)
	return lang.Instruction {
		OpCode: lang.POPGET,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstSet(index uint) lang.Instruction {
	ValidateProductIndex(index)
	return lang.Instruction {
		OpCode: lang.SET,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstRef(index uint, k ch.ReferenceKind, o ch.ReferenceOperand) lang.Instruction {
	ValidateProductIndex(index)
	var op = (func() lang.OpType {
		switch k {
		case ch.RK_Branch:
			switch o {
			case ch.RO_Enum:
				return lang.BRS
			case ch.RO_CaseRef:
				return lang.BRB
			case ch.RO_ProjRef:
				return lang.BRF
			default:
				panic("invalid operand")
			}
		case ch.RK_Field:
			switch o {
			case ch.RO_Record:
				return lang.FRP
			case ch.RO_ProjRef:
				return lang.FRF
			default:
				panic("invalid operand")
			}
		default:
			panic("impossible branch")
		}
	})()
	return lang.Instruction {
		OpCode: op,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstProduct(size uint) lang.Instruction {
	ValidateProductSize(size)
	return lang.Instruction {
		OpCode: lang.PROD,
		Arg0:   lang.Short(size),
		Arg1:   0,
	}
}

func InstArray(info_index uint) lang.Instruction {
	ValidateGlobalIndex(info_index)
	var a0, a1 = lang.GlobalIndex(info_index)
	return lang.Instruction {
		OpCode: lang.ARRAY,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstSum(index uint) lang.Instruction {
	ValidateSumIndex(index)
	return lang.Instruction {
		OpCode: lang.SUM,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstJumpIf(index uint, dest uint) lang.Instruction {
	ValidateSumIndex(index)
	ValidateDestAddr(dest)
	return lang.Instruction {
		OpCode: lang.JIF,
		Arg0:   lang.Short(index),
		Arg1:   lang.Long(dest),
	}
}

func InstJump(dest uint) lang.Instruction {
	ValidateDestAddr(dest)
	return lang.Instruction {
		OpCode: lang.JMP,
		Arg0:   0,
		Arg1:   lang.Long(dest),
	}
}

func InstMultiSwitchIndex(index uint) lang.Instruction {
	ValidateSumIndex(index)
	return lang.Instruction {
		OpCode: lang.MSI,
		Arg0:   lang.Short(index),
		Arg1:   0,
	}
}

func InstMultiSwitchJump(dest uint) lang.Instruction {
	ValidateDestAddr(dest)
	return lang.Instruction {
		OpCode: lang.MSJ,
		Arg0:   0,
		Arg1:   lang.Long(dest),
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= lang.GlobalSlotMaxSize {
		panic("global value index exceeded maximum slot capacity")
	}
}

func ValidateLocalOffset(offset uint) {
	if offset >= lang.LocalSlotMaxSize {
		panic("local binding offset exceeded maximum slot capacity")
	}
}

func ValidateDestAddr(addr uint) {
	if addr >= lang.FunCodeMaxLength {
		panic("destination address exceeded limitation")
	}
}

func ValidateProductIndex(index uint) {
	if index >= lang.ProductMaxSize {
		panic("value index exceeded maximum capacity of product type")
	}
}

func ValidateProductSize(size uint) {
	if size > lang.ProductMaxSize {
		panic("given size exceeded maximum capacity of product type")
	}
}

func ValidateSumIndex(index uint) {
	if index >= lang.SumMaxBranches {
		panic("given index exceeded maximum branch limit of sum type")
	}
}
