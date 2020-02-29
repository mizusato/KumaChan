package lib

import (
	"math"
	"math/big"
)
import . "kumachan/runtime/common"

var ArithmeticFunctions = map[string] NativeFunction {
	"+int": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		var c big.Int
		return IntValue(c.Add(IntFrom(a), IntFrom(b)))
	},
	"-int": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		var c big.Int
		return IntValue(c.Sub(IntFrom(a), IntFrom(b)))
	},
	"*int": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		var c big.Int
		return IntValue(c.Mul(IntFrom(a), IntFrom(b)))
	},
	"quorem": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		var q, m big.Int
		q.QuoRem(IntFrom(a), IntFrom(b), &m)
		return Tuple2(IntValue(&q), IntValue(&m))
	},
	"divmod": func(arg Value, handle MachineHandle) Value {
		var a, b = FromTuple2(arg)
		var q, m big.Int
		q.DivMod(IntFrom(a), IntFrom(b), &m)
		return Tuple2(IntValue(&q), IntValue(&m))
	},
	"+int8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int8Value(Int8From(a) + Int8From(b))
	},
	"-int8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int8Value(Int8From(a) - Int8From(b))
	},
	"*int8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int8Value(Int8From(a) * Int8From(b))
	},
	"/int8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int8Value(Int8From(a) / Int8From(b))
	},
	"%int8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int8Value(Int8From(a) % Int8From(b))
	},
	"+uint8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) + Uint8From(b))
	},
	"-uint8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) - Uint8From(b))
	},
	"*uint8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) * Uint8From(b))
	},
	"/uint8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) / Uint8From(b))
	},
	"%uint8": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint8Value(Uint8From(a) % Uint8From(b))
	},
	"+int16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int16Value(Int16From(a) + Int16From(b))
	},
	"-int16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int16Value(Int16From(a) - Int16From(b))
	},
	"*int16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int16Value(Int16From(a) * Int16From(b))
	},
	"/int16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int16Value(Int16From(a) / Int16From(b))
	},
	"%int16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int16Value(Int16From(a) % Int16From(b))
	},
	"+uint16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) + Uint16From(b))
	},
	"-uint16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) - Uint16From(b))
	},
	"*uint16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) * Uint16From(b))
	},
	"/uint16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) / Uint16From(b))
	},
	"%uint16": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint16Value(Uint16From(a) % Uint16From(b))
	},
	"+int32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int32Value(Int32From(a) + Int32From(b))
	},
	"-int32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int32Value(Int32From(a) - Int32From(b))
	},
	"*int32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int32Value(Int32From(a) * Int32From(b))
	},
	"/int32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int32Value(Int32From(a) / Int32From(b))
	},
	"%int32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int32Value(Int32From(a) % Int32From(b))
	},
	"+uint32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) + Uint32From(b))
	},
	"-uint32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) - Uint32From(b))
	},
	"*uint32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) * Uint32From(b))
	},
	"/uint32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) / Uint32From(b))
	},
	"%uint32": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint32Value(Uint32From(a) % Uint32From(b))
	},
	"+int64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int64Value(Int64From(a) + Int64From(b))
	},
	"-int64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int64Value(Int64From(a) - Int64From(b))
	},
	"*int64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int64Value(Int64From(a) * Int64From(b))
	},
	"/int64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int64Value(Int64From(a) / Int64From(b))
	},
	"%int64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Int64Value(Int64From(a) % Int64From(b))
	},
	"+uint64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) + Uint64From(b))
	},
	"-uint64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) - Uint64From(b))
	},
	"*uint64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) * Uint64From(b))
	},
	"/uint64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) / Uint64From(b))
	},
	"%uint64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Uint64Value(Uint64From(a) % Uint64From(b))
	},
	"+float64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Float64Value(Float64From(a) + Float64From(b))
	},
	"-float64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Float64Value(Float64From(a) - Float64From(b))
	},
	"*float64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Float64Value(Float64From(a) * Float64From(b))
	},
	"/float64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Float64Value(Float64From(a) / Float64From(b))
	},
	"%float64": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return Float64Value(math.Mod(Float64From(a), Float64From(b)))
	},
	"+float": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return FloatValue(FloatFrom(a) + FloatFrom(b))
	},
	"-float": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return FloatValue(FloatFrom(a) - FloatFrom(b))
	},
	"*float": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return FloatValue(FloatFrom(a) * FloatFrom(b))
	},
	"/float": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return FloatValue(FloatFrom(a) / FloatFrom(b))
	},
	"%float": func(arg Value, _ MachineHandle) Value {
		var a, b = FromTuple2(arg)
		return FloatValue(math.Mod(FloatFrom(a), FloatFrom(b)))
	},
}
