package node


func (impl TraitDeclaration) Declaration() {}
type TraitDeclaration struct {
    Node
    Name        string
    TP          [] TypeParameter
    ArgName     string
    Bases       [] TraitDeclaration
    Constraint  TraitConstraint
}


type TraitConstraint interface { TraitConstraint() }

func (impl ExactConstraint) TraitConstraint() {}
type ExactConstraint struct {
    Node
    Type TypeExpr
}

func (impl BoundConstraint) TraitConstraint() {}
type BoundConstraint struct {
    Node
    Kind      BoundKind
    BoundType TypeExpr
}

func (impl AttachedConstraint) TraitConstraint() {}
type AttachedConstraint struct {
    Node
    Items  map[string] AttachedItem
}

func (impl UnionConstraint) TraitConstraint() {}
type UnionConstraint struct {
    Node
    Items  [] TraitConstraint
}

func (impl IntersectionConstraint) TraitConstraint() {}
type IntersectionConstraint struct {
    Node
    Items  [] TraitConstraint
}


type BoundKind int
const (
    B_StrictSuperSet BoundKind = iota
    B_StrictSubSet
    B_SuperSet
    B_SubSet
    B_Equivalent
)

type AttachedItem interface { AttachedItem() }

func (impl AttachedTypeItem) AttachedItem() {}
type AttachedTypeItem struct {
    Node
    TypeTrait   TraitExpr
}

func (impl AttachedValueItem) AttachedItem() {}
type AttachedValueItem struct {
    Node
    ValueType TypeExpr
}

type TraitExpr struct {
    Node
    Module  string
    Name    string
    Args    [] TypeExpr
}
