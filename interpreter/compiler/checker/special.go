package checker

import (
	"kumachan/interpreter/def"
	"kumachan/stdlib"
)


const IgnoreMark = "_"
const UnitName = "unit"
const NeverTypeName = "never"
const AnyTypeName = "any"
const SuperTypeName = "super"
const TextPlaceholder = '#'
const ForceExactSuffix = "*"
const CovariantPrefix = "+"
const ContravariantPrefix = "-"
const BadIndex = ^(uint(0))
const DefaultValueGetter = "@default"
const KmdSerializerName = "@serialize"
const KmdDeserializerName = "@deserialize"
const KmdAdapterName = "@adapt"
const KmdValidatorName = "@validate"
var __Observable = CoreSymbol(stdlib.Observable)
var __Async = CoreSymbol(stdlib.Async)
var __ProjRef = CoreSymbol(stdlib.ProjRef)
var __CaseRef = CoreSymbol(stdlib.CaseRef)
var __ProjRefParams = [] TypeParam {
	TypeParam {
		Name:     "T",
		Variance: Invariant,
	},
	TypeParam {
		Name:     "C",
		Variance: Covariant,
	},
}
var __CaseRefParams = [] TypeParam {
	TypeParam {
		Name:     "T",
		Variance: Invariant,
	},
	TypeParam {
		Name:     "C",
		Variance: Covariant,
	},
}
var __ProjRefToBeInferred = MarkParamsAsBeingInferred(ProjRef(
	&ParameterType { Index: 0 },
	&ParameterType { Index: 1 },
))
var __CaseRefToBeInferred = MarkParamsAsBeingInferred(CaseRef(
	&ParameterType { Index: 0 },
	&ParameterType { Index: 1 },
))
func ProjRef(t Type, c Type) Type {
	return &NamedType {
		Name: __ProjRef,
		Args: [] Type { t, c },
	}
}
func CaseRef(t Type, c Type) Type {
	return &NamedType {
		Name: __CaseRef,
		Args: [] Type { t, c },
	}
}
var __DoTypes = [] Type {
	&NamedType { Name: __Async, Args: [] Type {&AnonymousType { Unit {} } } },
	&NamedType { Name: __Observable, Args: [] Type { &NeverType {} } },
}
var __VariousEffectType = &NamedType {
	Name: __Observable,
	Args: [] Type { &AnyType{}, &AnyType{} },
}
func VariousEffectType() Type { return __VariousEffectType }
var __ErrorType = &NamedType {
	Name: __Error,
	Args: [] Type {},
}
func ServiceInstanceType(mod string) Type {
	return &NamedType {
		Name: def.MakeSymbol(mod, stdlib.ServiceInstanceType),
		Args: [] Type {},
	}
}
var __Error = CoreSymbol(stdlib.Error)
var __Bool = CoreSymbol(stdlib.Bool)
var __T_Bool = &NamedType { Name: __Bool, Args: make([] Type, 0) }
var __Yes uint = stdlib.YesIndex
var __Maybe = CoreSymbol(stdlib.Maybe)
var __Float = CoreSymbol(stdlib.Float)
var __NormalFloat = CoreSymbol(stdlib.NormalFloat)
var __NormalComplex = CoreSymbol(stdlib.NormalComplex)
var __String = CoreSymbol(stdlib.String)
var __T_String = &NamedType { Name: __String, Args: make([] Type, 0) }
var __HardCodedString = CoreSymbol(stdlib.HardCodedString)
var __T_HardCodedString = &NamedType { Name: __HardCodedString, Args: make([] Type, 0) }
var __List = CoreSymbol(stdlib.List)
var __Integer = CoreSymbol(stdlib.Integer)
var __Number = CoreSymbol(stdlib.Number)
var __T_Integer = &NamedType { Name: __Integer, Args: make([] Type, 0) }
var __T_Number = &NamedType { Name: __Number, Args: make([] Type, 0) }
var __Qword = CoreSymbol(stdlib.Qword)
var __Dword = CoreSymbol(stdlib.Dword)
var __Char = CoreSymbol(stdlib.Char)
var __T_Char = &NamedType { Name: __Char, Args: make([] Type, 0) }
var __Word = CoreSymbol(stdlib.Word)
var __Byte = CoreSymbol(stdlib.Byte)
var __Bit = CoreSymbol(stdlib.Bit)
var __Bytes = CoreSymbol(stdlib.Bytes)
var __IntegerTypes = [] def.Symbol {
	__Integer, __Number,
	__Qword,
	__Dword, __Char,
	__Word,
	__Byte,
	__Bit,
}
var __IntegerTypeMap = (func() (map[def.Symbol] string) {
	var int_type_map = make(map[def.Symbol] string)
	for _, sym := range __IntegerTypes {
		int_type_map[sym] = sym.SymbolName
	}
	return int_type_map
})()

func CoreSymbol(name string) def.Symbol {
	return def.MakeSymbol(stdlib.Mod_core, name)
}

