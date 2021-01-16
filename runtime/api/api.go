package api

import (
	. "kumachan/lang"
	"fmt"
)


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
	UiFunctions,
	UiQtFunctions,
}
var NativeConstantMaps = [] (map[string] NativeConstant) {
	TimeConstants,
	OS_Constants,
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

func GetNativeFunction(id string) Value {
	var f, exists = NativeFunctionMap[id]
	if !(exists) { panic(fmt.Sprintf("no such native function: %s", id)) }
	return NativeFunctionValue(f)
}

func GetNativeConstant(id string, h InteropContext) Value {
	var c, exists = NativeConstantMap[id]
	if !(exists) { panic(fmt.Sprintf("no such native constant: %s", id)) }
	return c(h)
}

