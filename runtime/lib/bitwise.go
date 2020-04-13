package lib

import (
	. "kumachan/runtime/common"
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
		return ByteFrom(a) & ByteFrom(b)
	},
	"|Byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) | ByteFrom(b)
	},
	"~Byte": func(a interface{}) uint8 {
		return ^(ByteFrom(a))
	},
	"^Byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) ^ ByteFrom(b)
	},
	"<<Byte": func(a interface{}, b uint) uint8 {
		return ByteFrom(a) << b
	},
	">>Byte": func(a interface{}, b uint) uint8 {
		return ByteFrom(a) >> b
	},
	"&Word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) & WordFrom(b)
	},
	"|Word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) | WordFrom(b)
	},
	"~Word": func(a interface{}) uint16 {
		return ^(WordFrom(a))
	},
	"^Word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) ^ WordFrom(b)
	},
	"<<Word": func(a interface{}, b uint) uint16 {
		return WordFrom(a) << b
	},
	">>Word": func(a interface{}, b uint) uint16 {
		return WordFrom(a) >> b
	},
	"&Dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) & DwordFrom(b)
	},
	"|Dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) | DwordFrom(b)
	},
	"~Dword": func(a interface{}) uint32 {
		return ^(DwordFrom(a))
	},
	"^Dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) ^ DwordFrom(b)
	},
	"<<Dword": func(a interface{}, b uint) uint32 {
		return DwordFrom(a) << b
	},
	">>Dword": func(a interface{}, b uint) uint32 {
		return DwordFrom(a) >> b
	},
	"&Qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) & QwordFrom(b)
	},
	"|Qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) | QwordFrom(b)
	},
	"~Qword": func(a interface{}) uint64 {
		return ^(QwordFrom(a))
	},
	"^Qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) ^ QwordFrom(b)
	},
	"<<Qword": func(a interface{}, b uint) uint64 {
		return QwordFrom(a) << b
	},
	">>Qword": func(a interface{}, b uint) uint64 {
		return QwordFrom(a) >> b
	},
}
