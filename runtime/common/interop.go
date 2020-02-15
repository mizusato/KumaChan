package common


type MachineHandle interface {
	Call(fv Value, arg Value) Value
	CallAsync(fv Value, arg Value, cb func(Value))
}

type NativeFunction  func(arg Value, handle MachineHandle) Value
