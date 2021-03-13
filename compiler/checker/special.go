package checker

import (
	"kumachan/compiler/loader"
	"kumachan/stdlib"
)


const IgnoreMark = "_"
const UnitName = "unit"
const NeverTypeName = "never"
const AnyTypeName = "any"
const SuperTypeName = "super"
const TextPlaceholder = '#'
const FuncSuffix = "!!"
const ForceExactSuffix = "!"
const CovariantPrefix = "+"
const ContravariantPrefix = "-"
const BadIndex = ^(uint(0))
const ZeroValueGetter = "zero-value"
const KmdSerializerName = "data-serialize"
const KmdDeserializerName = "data-deserialize"
const KmdAdapterName = "data-adapt"
const KmdValidatorName = "data-validate"
var __Action = CoreSymbol(stdlib.Action)
var __ActionMultiValue = CoreSymbol(stdlib.ActionMultiValue)
var __Reactive = CoreSymbol(stdlib.Reactive)
func IsReactive(t *NamedType) bool {
	return t.Name == __Reactive
}
func Reactive(t Type) Type {
	return &NamedType {
		Name: __Reactive,
		Args: [] Type { t },
	}
}
var __DoTypes = [] Type {
	&NamedType { Name: __Action, Args: [] Type { &AnonymousType { Unit {} } } },
	&NamedType { Name: __ActionMultiValue, Args: [] Type { &NeverType {} } },
}
var __AnyActionType = &NamedType {
	Name: __ActionMultiValue,
	Args: [] Type { &AnyType{}, &AnyType{} },
}
func AnyActionType() Type { return __AnyActionType }
var __ErrorType = &NamedType {
	Name: __Error,
	Args: [] Type {},
}
func ServiceInstanceType(mod string) Type {
	return &NamedType {
		Name: loader.MakeSymbol(mod, stdlib.ServiceInstanceType),
		Args: [] Type {},
	}
}
var __Error = CoreSymbol(stdlib.Error)
var __Bool = CoreSymbol(stdlib.Bool)
var __T_Bool = &NamedType { Name: __Bool, Args: make([] Type, 0) }
var __Yes uint = stdlib.YesIndex
var __Maybe = CoreSymbol(stdlib.Maybe)
var __Float = CoreSymbol(stdlib.Float)
var __String = CoreSymbol(stdlib.String)
var __T_String = &NamedType { Name: __String, Args: make([] Type, 0) }
var __Array = CoreSymbol(stdlib.Array)
var __Int = CoreSymbol(stdlib.Int)
var __Number = CoreSymbol(stdlib.Number)
var __Int64 = CoreSymbol(stdlib.Int64)
var __Uint64 = CoreSymbol(stdlib.Uint64)
var __Qword = CoreSymbol(stdlib.Qword)
var __Int32 = CoreSymbol(stdlib.Int32)
var __Uint32 = CoreSymbol(stdlib.Uint32)
var __Dword = CoreSymbol(stdlib.Dword)
var __Char = CoreSymbol(stdlib.Char)
var __T_Char = &NamedType { Name: __Char, Args: make([] Type, 0) }
var __Int16 = CoreSymbol(stdlib.Int16)
var __Uint16 = CoreSymbol(stdlib.Uint16)
var __Word = CoreSymbol(stdlib.Word)
var __Int8 = CoreSymbol(stdlib.Int8)
var __Uint8 = CoreSymbol(stdlib.Uint8)
var __Byte = CoreSymbol(stdlib.Byte)
var __Bit = CoreSymbol(stdlib.Bit)
var __Bytes = CoreSymbol(stdlib.Bytes)
var __IntegerTypes = []loader.Symbol {
	__Int,   __Number,
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
	return loader.MakeSymbol(stdlib.Mod_core, name)
}
