package api

import (
	. "kumachan/lang"
	"fmt"
)


var NativeFunctionMaps = [] (map[string] interface{}) {
	DebuggingFunctions,
	AssertionFunctions,
	ErrorFunctions,
	MathFunctions,
	ComparisonFunctions,
	ContainerFunctions,
	EffectFunctions,
	BitwiseFunctions,
	IO_Functions,
	OS_Functions,
	NetFunctions,
	RpcFunctions,
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
var _ = (func() interface{} {
	NativeFunctionMap = make(map[string] NativeFunction)
	for _, category := range NativeFunctionMaps {
		for name, f := range category {
			var _, exists = NativeFunctionMap[name]
			if exists { panic("duplicate native function name " + name) }
			NativeFunctionMap[name] = AdaptNativeFunction(f)
		}
	}
	for _, category := range NativeConstantMaps {
		for name, constant := range category {
			var _, exists = NativeFunctionMap[name]
			if exists { panic("duplicate native function name " + name) }
			NativeFunctionMap[name] = constant.ToFunction()
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
	return nil
}) ()

func GetNativeFunctionValue(id string) NativeFunctionValue {
	var f, exists = NativeFunctionMap[id]
	if !(exists) { panic(fmt.Sprintf("no such native function: %s", id)) }
	return ValNativeFun(f)
}

