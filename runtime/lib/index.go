package lib

import . "kumachan/runtime/common"

var NativeFunctionMaps = [] map[string]NativeFunction {
	ArithmeticFunctions,
	BitwiseFunctions,
}

var NativeFunctionMap    map[string] NativeFunction
var NativeFunctionIndex  map[string] int
var NativeFunctions      [] NativeFunction
var __ = func() interface{} {
	NativeFunctionMap = make(map[string] NativeFunction)
	for _, category := range NativeFunctionMaps {
		for name, f := range category {
			NativeFunctionMap[name] = f
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
	return nil
} ()

