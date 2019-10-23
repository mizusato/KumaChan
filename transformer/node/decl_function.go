package node


func (impl Function) Declaration() {}
type Function struct {
    Node
    Name   string
    TP     [] TypeParameter
    Items  [] FunctionItem
}

type TypeParameter struct {
    Node
    Name        string
    Constraint  TraitExpr
}

type FunctionItem struct {
    Node
    Parameters  [] Parameter
    ReturnValue TypeExpr
    Body        FunctionBody
}

type Parameter struct {
    Node
    Name string
    Type TypeExpr
}

type FunctionBody struct {
    Node
    StaticBlock  Block
    Block        Block
}
