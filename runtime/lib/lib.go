package lib

import . "kumachan/runtime/common"


var NativeFunctionMaps = [] (map[string] interface{}) {
	ArithmeticFunctions,
	BitwiseFunctions,
}
var NativeConstantMaps = [] (map[string] Value) {}

var NativeFunctionMap    map[string] NativeFunction
var NativeFunctionIndex  map[string] int
var NativeFunctions      [] NativeFunction
var NativeConstantMap    map[string] Value
var NativeConstantIndex  map[string] int
var NativeConstants      [] Value
var _ = func() interface{} {
	NativeFunctionMap = make(map[string] NativeFunction)
	for _, category := range NativeFunctionMaps {
		for name, f := range category {
			var _, exists = NativeFunctionMap[name]
			if exists { panic("duplicate native function name " + name) }
			NativeFunctionMap[name] = AdaptNativeFunction(f)
		}
	}
	NativeFunctionIndex = make(map[string]int, len(NativeFunctionMap))
	NativeFunctions = make([]NativeFunction, len(NativeFunctionMap))
	var i = 0
	for name, f := range NativeFunctionMap {
		NativeFunctionIndex[name] = i
		NativeFunctions[i] = f
		i += 1
	}
	// ---------------
	NativeConstantMap = make(map[string] Value)
	for _, category := range NativeConstantMaps {
		for name, v := range category {
			var _, exists = NativeConstantMap[name]
			if exists { panic("duplicate native constant name " + name) }
			NativeConstantMap[name] = v
		}
	}
	NativeConstantIndex = make(map[string] int)
	NativeConstants = make([]Value, len(NativeConstantMap))
	i = 0
	for name, v := range NativeConstantMap {
		NativeConstantIndex[name] = i
		NativeConstants[i] = v
	}
	// ---------------
	return nil
} ()
