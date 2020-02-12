package common


type MachineHandle interface {
	Call(f Value, arg Value) Value
}

type NativeFunction  func(arg Value, handle MachineHandle) Value
