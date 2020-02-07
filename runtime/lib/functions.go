package lib

import . "kumachan/runtime/common"

type MachineHandle interface {
	Call(f FunctionValue, arg Value) Value
}

type NativeFunction  func(arg Value, handle MachineHandle) Value

var NativeFunctionMap = map[string] NativeFunction {

}
var NativeFunctionIndex = make(map[string] int)
var NativeFunctions = func() []NativeFunction {
	var array = make([]NativeFunction, len(NativeFunctionMap))
	var i = 0
	for name, f := range NativeFunctionMap {
		NativeFunctionIndex[name] = i
		array[i] = f
		i += 1
	}
	return array
} ()