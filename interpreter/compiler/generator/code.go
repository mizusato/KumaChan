package generator

import (
	ch "kumachan/interpreter/compiler/checker"
	. "kumachan/standalone/util/error"
	. "kumachan/interpreter/def"
)


type Code struct {
	InstSeq    [] Instruction
	SourceMap  [] ErrorPoint
}

func CodeFrom(inst Instruction, info ch.ExprInfo) Code {
	return Code {
		InstSeq:   [] Instruction { inst },
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
		InstSeq:   make([] Instruction, 0),
		SourceMap: make([] ErrorPoint, 0),
	}
	return CodeBuffer { code }
}

func (buf CodeBuffer) Write(code Code) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	for _, inst := range code.InstSeq {
		switch inst.OpCode {
		case JIF, JMP, MSJ:
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			*base = append(*base, Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   Long(dest_addr),
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
		case JIF, JMP, MSJ:
			var rel_dest_addr = uint(inst.Arg1)
			var abs_dest_addr = (rel_dest_addr + base_size)
			ValidateDestAddr(abs_dest_addr)
			if rel_dest_addr == last_addr {
				abs_dest_addr = tail_addr
			}
			*base = append(*base, Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   Long(abs_dest_addr),
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


func InstGlobalRef(index uint) Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = GlobalIndex(index)
	return Instruction {
		OpCode: GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstLocalRef(offset uint) Instruction {
	ValidateLocalOffset(offset)
	return Instruction {
		OpCode: LOAD,
		Arg0:   0,
		Arg1:   Long(offset),
	}
}

func InstStore(offset uint) Instruction {
	ValidateLocalOffset(offset)
	return Instruction {
		OpCode: STORE,
		Arg0:   0,
		Arg1:   Long(offset),
	}
}

func InstGet(index uint) Instruction {
	ValidateProductIndex(index)
	return Instruction {
		OpCode: GET,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstPopGet(index uint) Instruction {
	ValidateProductIndex(index)
	return Instruction {
		OpCode: POPGET,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstSet(index uint) Instruction {
	ValidateProductIndex(index)
	return Instruction {
		OpCode: SET,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstRef(index uint, k ch.ReferenceKind, o ch.ReferenceOperand) Instruction {
	ValidateProductIndex(index)
	var op = (func() OpType {
		switch k {
		case ch.RK_Branch:
			switch o {
			case ch.RO_Enum:
				return BR
			case ch.RO_CaseRef:
				return BRB
			case ch.RO_ProjRef:
				return BRF
			default:
				panic("invalid operand")
			}
		case ch.RK_Field:
			switch o {
			case ch.RO_Record:
				return FR
			case ch.RO_ProjRef:
				return FRF
			default:
				panic("invalid operand")
			}
		default:
			panic("impossible branch")
		}
	})()
	return Instruction {
		OpCode: op,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstProduct(size uint) Instruction {
	ValidateProductSize(size)
	return Instruction {
		OpCode: TUP,
		Arg0:   Short(size),
		Arg1:   0,
	}
}

func InstArray(info_index uint) Instruction {
	ValidateGlobalIndex(info_index)
	var a0, a1 = GlobalIndex(info_index)
	return Instruction {
		OpCode: ARRAY,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstSum(index uint) Instruction {
	ValidateSumIndex(index)
	return Instruction {
		OpCode: ENUM,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstJumpIf(index uint, dest uint) Instruction {
	ValidateSumIndex(index)
	ValidateDestAddr(dest)
	return Instruction {
		OpCode: JIF,
		Arg0:   Short(index),
		Arg1:   Long(dest),
	}
}

func InstJump(dest uint) Instruction {
	ValidateDestAddr(dest)
	return Instruction {
		OpCode: JMP,
		Arg0:   0,
		Arg1:   Long(dest),
	}
}

func InstMultiSwitchIndex(index uint) Instruction {
	ValidateSumIndex(index)
	return Instruction {
		OpCode: MSI,
		Arg0:   Short(index),
		Arg1:   0,
	}
}

func InstMultiSwitchJump(dest uint) Instruction {
	ValidateDestAddr(dest)
	return Instruction {
		OpCode: MSJ,
		Arg0:   0,
		Arg1:   Long(dest),
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= GlobalSlotMaxSize {
		panic("global value index exceeded maximum slot capacity")
	}
}

func ValidateLocalOffset(offset uint) {
	if offset >= LocalSlotMaxSize {
		panic("local binding offset exceeded maximum slot capacity")
	}
}

func ValidateDestAddr(addr uint) {
	if addr >= FunCodeMaxLength {
		panic("destination address exceeded limitation")
	}
}

func ValidateProductIndex(index uint) {
	if index >= ProductMaxSize {
		panic("value index exceeded maximum capacity of product type")
	}
}

func ValidateProductSize(size uint) {
	if size > ProductMaxSize {
		panic("given size exceeded maximum capacity of product type")
	}
}

func ValidateSumIndex(index uint) {
	if index >= SumMaxBranches {
		panic("given index exceeded maximum branch limit of sum type")
	}
}
