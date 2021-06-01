package def


type InteropContext interface {
	Call(f Value, arg Value) Value
	// TODO
}

type NativeFunction func(arg Value, handle InteropContext) Value
type NativeConstant  func(handle InteropContext) Value
func (c NativeConstant) ToFunction() NativeFunction {
	return func(_ Value, h InteropContext) Value {
		return c(h)
	}
}

