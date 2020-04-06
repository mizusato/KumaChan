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
	"<<Byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) << ByteFrom(b)
	},
	">>Byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) >> ByteFrom(b)
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
	"<<Word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) << ByteFrom(b)
	},
	">>Word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) >> ByteFrom(b)
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
	"<<Dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) << ByteFrom(b)
	},
	">>Dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) >> ByteFrom(b)
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
	"<<Qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) << ByteFrom(b)
	},
	">>Qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) >> ByteFrom(b)
	},
}
