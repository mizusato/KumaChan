package node


type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}
func (impl Expr) ConstValue() {}
type Expr struct {
    Node             `part:"expr"`
    Pipes  [] Pipe   `list_more:"" item:"pipe"`
}

type Pipe struct {
    Node                    `part:"pipe"`
    Terms  [] VariousTerm   `list_more:"" item:"term"`
}

type VariousTerm struct {
    Node         `part:"term"`
    Term  Term   `use:"first"`
}
type Term interface { Term() }
