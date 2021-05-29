package api

import (
	. "kumachan/interpreter/base"
)


var BitwiseFunctions = map[string] interface{} {
	"&Bit": func(a bool, b bool) bool {
		return a && b
	},
	"|Bit": func(a bool, b bool) bool {
		return a || b
	},
	"~Bit": func(a bool) bool {
		return !a
	},
	"^Bit": func(a bool, b bool) bool {
		return (a || b) && (!(a && b))
	},
	"&Byte": func(a interface{}, b interface{}) uint8 {
		return FromByte(a) & FromByte(b)
	},
	"|Byte": func(a interface{}, b interface{}) uint8 {
		return FromByte(a) | FromByte(b)
	},
	"~Byte": func(a interface{}) uint8 {
		return ^(FromByte(a))
	},
	"^Byte": func(a interface{}, b interface{}) uint8 {
		return FromByte(a) ^ FromByte(b)
	},
	"<<Byte": func(a interface{}, b uint) uint8 {
		return FromByte(a) << b
	},
	">>Byte": func(a interface{}, b uint) uint8 {
		return FromByte(a) >> b
	},
	"&Word": func(a interface{}, b interface{}) uint16 {
		return FromWord(a) & FromWord(b)
	},
	"|Word": func(a interface{}, b interface{}) uint16 {
		return FromWord(a) | FromWord(b)
	},
	"~Word": func(a interface{}) uint16 {
		return ^(FromWord(a))
	},
	"^Word": func(a interface{}, b interface{}) uint16 {
		return FromWord(a) ^ FromWord(b)
	},
	"<<Word": func(a interface{}, b uint) uint16 {
		return FromWord(a) << b
	},
	">>Word": func(a interface{}, b uint) uint16 {
		return FromWord(a) >> b
	},
	"&Dword": func(a interface{}, b interface{}) uint32 {
		return FromDword(a) & FromDword(b)
	},
	"|Dword": func(a interface{}, b interface{}) uint32 {
		return FromDword(a) | FromDword(b)
	},
	"~Dword": func(a interface{}) uint32 {
		return ^(FromDword(a))
	},
	"^Dword": func(a interface{}, b interface{}) uint32 {
		return FromDword(a) ^ FromDword(b)
	},
	"<<Dword": func(a interface{}, b uint) uint32 {
		return FromDword(a) << b
	},
	">>Dword": func(a interface{}, b uint) uint32 {
		return FromDword(a) >> b
	},
	"&Qword": func(a interface{}, b interface{}) uint64 {
		return FromQword(a) & FromQword(b)
	},
	"|Qword": func(a interface{}, b interface{}) uint64 {
		return FromQword(a) | FromQword(b)
	},
	"~Qword": func(a interface{}) uint64 {
		return ^(FromQword(a))
	},
	"^Qword": func(a interface{}, b interface{}) uint64 {
		return FromQword(a) ^ FromQword(b)
	},
	"<<Qword": func(a interface{}, b uint) uint64 {
		return FromQword(a) << b
	},
	">>Qword": func(a interface{}, b uint) uint64 {
		return FromQword(a) >> b
	},
}
