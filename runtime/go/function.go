const PARAMETER_MAX = 10


type Constraint struct {
    category string
    abstract string
}


type ParaExpr struct {
    name string
    constraint Constraint
}


type ProtoExpr struct {
    affect EffectRange
    value Constraint
    parameters [PARAMETER_MAX]ParaExpr
    size int
}


type FunData struct {
    proto ProtoExpr
    body []Instruction
    name string
    info string
}


type Parameter struct {
    name string
    constraint AbstractObject
}


type Prototype struct {
    affect EffectRange
    value AbstractObject
    parameters [PARAMETER_MAX]Parameter
    size int
}


type FunctionObject struct {
    module_index int
    data_index int
    context *Scope
    prototype Prototype
}


func (f *FunctionObject) get_type() Type { return Function }


type OverloadObject struct {    
    functions []*FunctionObject
}


func (o *OverloadObject) get_type() Type { return Overload }
