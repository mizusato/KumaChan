package node


type AttachedDeclaration interface {
    DeclContent
    AttachedDeclaration()
}

func (impl AttachedType) DeclContent() {}
func (impl AttachedType) AttachedDeclaration() {}
type AttachedType struct {
    Node
    Target    AttachedExpr
    Attached  TypeExpr
}

func (impl AttachedFunction) DeclContent() {}
func (impl AttachedFunction) AttachedDeclaration() {}
type AttachedFunction struct {
    Node
    Target    AttachedExpr
    Attached  FunctionItem
}

func (impl AttachedValue) DeclContent() {}
func (impl AttachedValue) AttachedDeclaration() {}
type AttachedValue struct {
    Node
    Target    AttachedExpr
    Attached  Expr
}


type AttachedExpr struct {
    Node
    Type          TypeExpr
    AttachedName  Identifier
}
