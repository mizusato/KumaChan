package lib

import . "kumachan/runtime/common"


var NativeFunctionMaps = [] (map[string] interface{}) {
	DebuggingFunctions,
	MathFunctions,
	ComparisonFunctions,
	ContainerFunctions,
	EffectFunctions,
	BitwiseFunctions,
	IO_Functions,
	OS_Functions,
	NetFunctions,
	QtFunctions,
	WebUiFunctions,
}
var NativeConstantMaps = [] (map[string] NativeConstant) {
	OS_Constants,
	WebUiConstants,
}

var NativeFunctionMap    map[string] NativeFunction
var NativeFunctionIndex  map[string] uint
var NativeFunctions      [] NativeFunction
var NativeConstantMap    map[string] NativeConstant
var NativeConstantIndex  map[string] uint
var NativeConstants      [] NativeConstant
var _ = (func() interface{} {
	NativeFunctionMap = make(map[string] NativeFunction)
	for _, category := range NativeFunctionMaps {
		for name, f := range category {
			var _, exists = NativeFunctionMap[name]
			if exists { panic("duplicate native function name " + name) }
			NativeFunctionMap[name] = AdaptNativeFunction(f)
		}
	}
	NativeFunctionIndex = make(map[string] uint, len(NativeFunctionMap))
	NativeFunctions = make([] NativeFunction, len(NativeFunctionMap))
	var i = uint(0)
	for name, f := range NativeFunctionMap {
		NativeFunctionIndex[name] = i
		NativeFunctions[i] = f
		i += 1
	}
	// ---------------
	NativeConstantMap = make(map[string] NativeConstant)
	for _, category := range NativeConstantMaps {
		for name, v := range category {
			var _, exists = NativeConstantMap[name]
			if exists { panic("duplicate native constant name " + name) }
			NativeConstantMap[name] = v
		}
	}
	NativeConstantIndex = make(map[string] uint)
	NativeConstants = make([] NativeConstant, len(NativeConstantMap))
	i = 0
	for name, v := range NativeConstantMap {
		NativeConstantIndex[name] = i
		NativeConstants[i] = v
		i += 1
	}
	// ---------------
	return nil
}) ()

func GetNativeFunction(i uint) Value {
	return NativeFunctionValue(NativeFunctions[i])
}

func GetNativeConstant(i uint, h MachineHandle) Value {
	return NativeConstants[i](h)
}
