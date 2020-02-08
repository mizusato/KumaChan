package lib

import . "kumachan/runtime/common"

var BitwiseFunctions = map[string] NativeFunction {
	"&bit": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return BitValue(BitFrom(a) && BitFrom(b))
	},
	"|bit": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return BitValue(BitFrom(a) || BitFrom(b))
	},
	"^bit": func(arg Value, _ MachineHandle) Value {
		return BitValue(!(BitFrom(arg)))
	},
	"&byte": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) & Uint8From(b))
	},
	"|byte": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) | Uint8From(b))
	},
	"^byte": func(arg Value, _ MachineHandle) Value {
		return Uint8Value(^(Uint8From(arg)))
	},
	"<<byte": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) << Uint8From(b))
	},
	">>byte": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) >> Uint8From(b))
	},
	"&word": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) & Uint16From(b))
	},
	"|word": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) | Uint16From(b))
	},
	"^word": func(arg Value, _ MachineHandle) Value {
		return Uint16Value(^(Uint16From(arg)))
	},
	"<<word": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) << Uint8From(b))
	},
	">>word": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) >> Uint8From(b))
	},
	"&dword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) & Uint32From(b))
	},
	"|dword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) | Uint32From(b))
	},
	"^dword": func(arg Value, _ MachineHandle) Value {
		return Uint32Value(^(Uint32From(arg)))
	},
	"<<dword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) << Uint8From(b))
	},
	">>dword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) >> Uint8From(b))
	},
	"&qword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) & Uint64From(b))
	},
	"|qword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) | Uint64From(b))
	},
	"^qword": func(arg Value, _ MachineHandle) Value {
		return Uint64Value(^(Uint64From(arg)))
	},
	"<<qword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) << Uint8From(b))
	},
	">>qword": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) >> Uint8From(b))
	},
}
