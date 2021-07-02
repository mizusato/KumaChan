package checker2

import (
	"math"
	"unicode"
	"math/big"
	"kumachan/stdlib"
	"kumachan/standalone/util"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/compiler/checker2/typsys"
)


const MaxTypeParameters = 8
const MaxImplemented = 8
const Discarded = "_"

type nominalType func(TypeRegistry)(typsys.Type)

type integerTypeInfo struct {
	which  nominalType
	adapt  integerAdapter
}
type integerAdapter func(*big.Int)(interface{}, bool)

var integerTypes = [] integerTypeInfo {
	{ which: coreNumber, adapt: integerAdapter(func(v *big.Int) (interface{}, bool) {
		if util.IsNonNegative(v) {
			return v, true
		} else {
			return nil, false
		}
	}) },
	{ which: coreInteger, adapt: func(v *big.Int) (interface{}, bool) {
		return v, true
	} },
	{ which: coreQword, adapt: smallUintAdapt(math.MaxUint64, func(v *big.Int) interface{} {
		return v.Uint64()
	}) },
	{ which: coreDword, adapt: smallUintAdapt(math.MaxUint32, func(v *big.Int) interface{} {
		return uint32(v.Uint64())
	}) },
	{ which: coreChar, adapt: smallUintAdapt(unicode.MaxRune, func(v *big.Int) interface{} {
		return rune(v.Uint64())
	}) },
	{ which: coreWord, adapt: smallUintAdapt(math.MaxUint16, func(v *big.Int) interface{} {
		return uint16(v.Uint64())
	}) },
	{ which: coreByte, adapt: smallUintAdapt(math.MaxUint8, func(v *big.Int) interface{} {
		return byte(v.Uint64())
	}) },
}

var floatTypes = [] nominalType {
	coreNormalFloat,
	coreFloat,
}

func typeEqual(t typsys.Type, nt nominalType, reg TypeRegistry) bool {
	return typsys.TypeOpEqual(t, nt(reg))
}

func smallIntAdapt(min int64, max uint64, cast func(*big.Int)(interface{})) integerAdapter {
	return integerAdapter(func(value *big.Int) (interface{}, bool) {
		var min = big.NewInt(0).SetInt64(min)
		var max = big.NewInt(0).SetUint64(max)
		if (min.Cmp(value) <= 0) && (value.Cmp(max) <= 0) {
			return cast(value), true
		} else {
			return nil, false
		}
	})
}

func smallUintAdapt(max uint64, cast func(*big.Int)(interface{})) integerAdapter {
	return smallIntAdapt(0, max, cast)
}


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


var typeBadNames = [] string {
	Discarded,
	typsys.TypeNameUnit,
	typsys.TypeNameUnknown,
	typsys.TypeNameTop,
	typsys.TypeNameBottom,
}
func isValidTypeItemName(name string) bool {
	for _, bad := range typeBadNames {
		if name == bad {
			return false
		}
	}
	return true
}

var functionBadNames = append([] string {}, typeBadNames...)
func isValidFunctionItemName(name string) bool {
	for _, bad := range functionBadNames {
		if name == bad {
			return false
		}
	}
	return true
}


