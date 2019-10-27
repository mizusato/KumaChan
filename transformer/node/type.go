package node


type TypeExpr struct {
    Node
    Content TypeExprContent
}

type MaybeTypeExpr interface { MaybeTypeExpr() }
func (impl TypeExpr) MaybeTypeExpr() {}

type TypeExprContent interface { TypeExprContent() }

func (impl OrdinaryTypeExpr) TypeExprContent() {}
type OrdinaryTypeExpr struct {
    Node
    Module  Identifier
    Name    Identifier
    Args    [] TypeExpr
}

func (impl AttachedTypeExpr) TypeExprContent() {}
type AttachedTypeExpr struct {
    Node
    AttachedExpr  AttachedExpr
}

func (impl TraitTypeExpr) TypeExprContent() {}
type TraitTypeExpr struct {
    Node
    Trait  TraitExpr
}

func (impl TupleTypeExpr) TypeExprContent() {}
type TupleTypeExpr struct {
    Node
    ElementTypes  [] TypeExpr
}

func (impl FunctionTypeExpr) TypeExprContent() {}
type FunctionTypeExpr struct {
    Node
    Signatures  [] Signature
}

func (impl IteratorTypeExpr) TypeExprContent() {}
type IteratorTypeExpr struct {
    Node
    YieldType  TypeExpr
}

func (impl ContinuationTypeExpr) TypeExprContent() {}
type ContinuationTypeExpr struct {
    Node
    ResumingType  TypeExpr
}


type Signature struct {
    Node
    Parameters   [] TypeExpr
    ReturnValue  TypeExpr
}


