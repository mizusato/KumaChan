package lib

import (
	. "kumachan/runtime/common"
)


var BitwiseFunctions = map[string] interface{} {
	"&bit": func(a bool, b bool) bool {
		return a && b
	},
	"|bit": func(a bool, b bool) bool {
		return a || b
	},
	"^bit": func(a bool) bool {
		return !a
	},
	"&byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) & ByteFrom(b)
	},
	"|byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) | ByteFrom(b)
	},
	"^byte": func(a interface{}) uint8 {
		return ^(ByteFrom(a))
	},
	"<<byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) << ByteFrom(b)
	},
	">>byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) >> ByteFrom(b)
	},
	"&word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) & WordFrom(b)
	},
	"|word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) | WordFrom(b)
	},
	"^word": func(a interface{}) uint16 {
		return ^(WordFrom(a))
	},
	"<<word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) << ByteFrom(b)
	},
	">>word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) >> ByteFrom(b)
	},
	"&dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) & DwordFrom(b)
	},
	"|dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) | DwordFrom(b)
	},
	"^dword": func(a interface{}) uint32 {
		return ^(DwordFrom(a))
	},
	"<<dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) << ByteFrom(b)
	},
	">>dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) >> ByteFrom(b)
	},
	"&qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) & QwordFrom(b)
	},
	"|qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) | QwordFrom(b)
	},
	"^qword": func(a interface{}) uint64 {
		return ^(QwordFrom(a))
	},
	"<<qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) << ByteFrom(b)
	},
	">>qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) >> ByteFrom(b)
	},
}
