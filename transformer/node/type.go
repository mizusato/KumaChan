package node


// type = type_ordinary | type_attached | type_trait | type_misc
type TypeExpr struct {
    Node                      `part:"type"`
    Content TypeExprContent   `use:"first"`
}

type MaybeTypeExpr interface { MaybeTypeExpr() }
func (impl TypeExpr) MaybeTypeExpr() {}

type TypeExprContent interface { TypeExprContent() }

// type_ordinary = module_prefix name type_args
// module_prefix? = name ::
// type_args? = NoLF [ typelist! ]!
func (impl OrdinaryTypeExpr) TypeExprContent() {}
type OrdinaryTypeExpr struct {
    Node                  `part:"type_ordinary"`
    Module  Identifier    `part_opt:"module_prefix.name"`
    Name    Identifier    `part:"name"`
    Args    [] TypeExpr   `list:"type_args.typelist"`
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


