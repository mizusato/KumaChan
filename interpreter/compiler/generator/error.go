package generator

import . "kumachan/standalone/util/error"


type Error struct {
	Point     ErrorPoint
	Concrete  ConcreteError
}

type ConcreteError interface { GeneratorError() }

func (impl E_CircularThunkDependency) GeneratorError() {}
type E_CircularThunkDependency struct {
	Names  [] string
}

func (impl E_UnusedPrivateFunctions) GeneratorError() {}
type E_UnusedPrivateFunctions struct {
	Names  [] string
}

func (impl E_UnusedBinding) GeneratorError() {}
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
	case E_CircularThunkDependency:
		desc.WriteText(TS_ERROR, "Circular dependency detected among thunks:")
		desc.Write(T_SPACE)
		var names = e.Names
		for i, name := range names {
			desc.WriteText(TS_INLINE_CODE, name)
			if i != len(names)-1 {
				desc.WriteText(TS_ERROR, ", ")
			}
		}
	case E_UnusedPrivateFunctions:
		desc.WriteText(TS_ERROR, "Unused private (unexported) function(s):")
		desc.Write(T_SPACE)
		var names = e.Names
		for i, name := range names {
			desc.WriteText(TS_INLINE_CODE, name)
			if i != len(names)-1 {
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
