package checker

import "kumachan/loader"

const IgnoreMark = "_"
const TextPlaceholder = '#'
const BadIndex = ^(uint(0))
/* should be consistent with `stdlib/core.km` */
var __EffectSingleValue = CoreSymbol("Effect")
var __DoType = NamedType {
	Name: __EffectSingleValue,
	Args: []Type { AnonymousType { Unit{} }, AnonymousType { Unit{} } },
}
var __Bool = CoreSymbol("Bool")
var __Yes uint = 0
// var __Maybe = CoreSymbol("Maybe")
// var __Just uint = 0
// var __Nothing uint = 1
var __Float = CoreSymbol("Float")
var __String = CoreSymbol("String")
var __Array = CoreSymbol("Array")
var __Integer = CoreSymbol("Integer")
var __Natural = CoreSymbol("Natural")
var __Int64 = CoreSymbol("Int64")
var __Uint64 = CoreSymbol("Uint64")
var __Qword = CoreSymbol("Qword")
var __Int32 = CoreSymbol("Int32")
var __Uint32 = CoreSymbol("Uint32")
var __Dword = CoreSymbol("Dword")
var __Char = CoreSymbol("Char")
var __Int16 = CoreSymbol("Int16")
var __Uint16 = CoreSymbol("Uint16")
var __Word = CoreSymbol("Word")
var __Int8 = CoreSymbol("Int8")
var __Uint8 = CoreSymbol("Uint8")
var __Byte = CoreSymbol("Byte")
var __Bit = CoreSymbol("Bit")
var __IntegerTypes = []loader.Symbol {
	__Integer, __Natural,
	__Int64, __Uint64, __Qword,
	__Int32, __Uint32, __Dword, __Char,
	__Int16, __Uint16, __Word,
	__Int8,  __Uint8,  __Byte,
	__Bit,
}
var __IntegerTypeMap = (func() map[loader.Symbol]string {
	var int_type_map = make(map[loader.Symbol]string)
	for _, sym := range __IntegerTypes {
		int_type_map[sym] = sym.SymbolName
	}
	return int_type_map
})()

func CoreSymbol(name string) loader.Symbol {
	return loader.NewSymbol(loader.CoreModule, name)
}
