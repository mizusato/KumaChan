package node


type AttachedDeclaration interface {
    Declaration
    AttachedDeclaration()
}

func (impl AttachedType) Declaration() {}
func (impl AttachedType) AttachedDeclaration() {}
type AttachedType struct {
    Node
    Target    AttachedExpr
    Attached  TypeExpr
}

func (impl AttachedFunction) Declaration() {}
func (impl AttachedFunction) AttachedDeclaration() {}
type AttachedFunction struct {
    Node
    Target    AttachedExpr
    Attached  FunctionItem
}

func (impl AttachedValue) Declaration() {}
func (impl AttachedValue) AttachedDeclaration() {}
type AttachedValue struct {
    Node
    Target    AttachedExpr
    Attached  Expr
}


type AttachedExpr struct {
    Node
    Type          TypeExpr
    AttachedName  string
}
