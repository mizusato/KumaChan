package generator

import . "kumachan/util/error"


type Error struct {
	Point     ErrorPoint
	Concrete  ConcreteError
}

type ConcreteError interface { CompilerError() }

func (impl E_NativeFunctionNotFound) CompilerError() {}
type E_NativeFunctionNotFound struct {
	Name  string
}

func (impl E_NativeConstantNotFound) CompilerError() {}
type E_NativeConstantNotFound struct {
	Name  string
}

func (impl E_CircularConstantDependency) CompilerError() {}
type E_CircularConstantDependency struct {
	Constants  [] string
}

func (impl E_UnusedBinding) CompilerError() {}
type E_UnusedBinding struct {
	Name  string
}

func (err *Error) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *Error) ErrorConcrete() interface{} {
	return err.Concrete
}

func (err *Error) Desc() ErrorMessage {
	var desc = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_NativeFunctionNotFound:
		desc.WriteText(TS_ERROR, "No such native function:")
		desc.WriteEndText(TS_INLINE, e.Name)
	case E_NativeConstantNotFound:
		desc.WriteText(TS_ERROR, "No such native constant:")
		desc.WriteEndText(TS_INLINE, e.Name)
	case E_CircularConstantDependency:
		desc.WriteText(TS_ERROR,
			"Circular dependency detected within constants:")
		desc.Write(T_SPACE)
		for i, item := range e.Constants {
			desc.WriteText(TS_INLINE_CODE, item)
			if i != len(e.Constants)-1 {
				desc.WriteText(TS_ERROR, ", ")
			}
		}
	case E_UnusedBinding:
		desc.WriteText(TS_ERROR, "Unused binding:")
		desc.WriteEndText(TS_INLINE_CODE, e.Name)
	}
	return desc
}

func (err *Error) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *Error) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, [] ErrorMessage {
		err.Message(),
	})
	return msg.String()
}
