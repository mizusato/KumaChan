package checker2

import (
	"kumachan/stdlib"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/compiler/checker2/typsys"
)


const MaxTupleSize = 8
const MaxRecordSize = 64
const MaxEnumSize = 64
const MaxTypeParameters = 8
const MaxImplemented = 8
const Discarded = "_"

var typeBadNames = [] string {
	Discarded,
	typsys.TypeNameUnit,
	typsys.TypeNameUnknown,
	typsys.TypeNameTop,
	typsys.TypeNameBottom,
}
var functionBadNames = append([] string {}, typeBadNames...)

var coreFloat = makeCoreType(stdlib.Float)
var coreNormalFloat = makeCoreType(stdlib.NormalFloat)
var coreInteger = makeCoreType(stdlib.Integer)
var coreNumber = makeCoreType(stdlib.Number)
var coreQword = makeCoreType(stdlib.Qword)
var coreDword = makeCoreType(stdlib.Dword)
var coreChar = makeCoreType(stdlib.Char)
var coreWord = makeCoreType(stdlib.Word)
var coreByte = makeCoreType(stdlib.Byte)

var coreTypes = (func() (map[string] struct{}) {
	var set = make(map[string] struct{})
	var list = stdlib.CoreTypeNames()
	for _, name := range list {
		set[name] = struct{}{}
	}
	return set
})()

func isValidTypeItemName(name string) bool {
	for _, bad := range typeBadNames {
		if name == bad {
			return false
		}
	}
	return true
}
func isValidFunctionItemName(name string) bool {
	for _, bad := range functionBadNames {
		if name == bad {
			return false
		}
	}
	return true
}


type nominalType func
	(TypeRegistry)(typsys.Type)

func makeCoreType(item_name string, args ...nominalType) nominalType {
	return nominalType(func(reg TypeRegistry) typsys.Type {
		var n = name.MakeTypeName(stdlib.Mod_core, item_name)
		var def, exists = reg[n]
		if !(exists) { panic("something went wrong") }
		var args_types = make([] typsys.Type, len(args))
		for i, arg := range args {
			args_types[i] = arg(reg)
		}
		return &typsys.NestedType {
			Content: typsys.Ref {
				Def:  def.TypeDef,
				Args: args_types,
			},
		}
	})
}


