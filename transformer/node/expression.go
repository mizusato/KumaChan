package node


type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}
type Expr struct {
    Node             `part:"expr"`
    Casts  [] Cast   `list_rec:"casts"`
    Pipes  [] Pipe   `list_more:"" item:"pipe"`
}

type Cast struct {
    Node                `part:"cast"`
    Target  MaybeType   `part_opt:"cast_target.type"`
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
