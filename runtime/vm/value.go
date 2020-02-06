package vm

type Value interface { RuntimeValue() }

func (impl V_Final) RuntimeValue() {}
type V_Final struct {
	Pointer interface{}
}

func (impl V_Sum) RuntimeValue() {}
type V_Sum struct {
	Index  Short
	Value  Value
}

func (impl V_Product) RuntimeValue() {}
type V_Product struct {
	Elements  [] Value
}

func (impl V_Function) RuntimeValue() {}
type V_Function struct {
	Underlying     *Function
	ContextValues  [] Value
}
