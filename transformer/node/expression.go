package node


type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}
func (impl Expr) ConstValue() {}
type Expr struct {
    Node                      `part:"expr"`
    Call      Terms           `part:"terms"`
    Pipeline  MaybePipeline   `part_opt:"pipeline"`
}

type MaybeTerms interface { MaybeTerms() }
func (impl Terms) MaybeTerms() {}
type Terms struct {
    Node                    `part:"terms"`
    Terms  [] VariousTerm   `list_more:"" item:"term"`
}

type VariousTerm struct {
    Node         `part:"term"`
    Term  Term   `use:"first"`
}
type Term interface { Term() }

type MaybePipeline interface { MaybePipeline() }
func (impl Pipeline) MaybePipeline() {}
type Pipeline struct {
    Node                      `part:"pipeline"`
    Operator  PipeOperator    `part:"pipe_op"`
    Func      VariousTerm     `part:"pipe_func.term"`
    Args      MaybeTerms      `part_opt:"pipe_args.terms"`
    Next      MaybePipeline   `part_opt:"pipeline"`
}

type PipeOperator struct {
    Node  `part:"pipe_op"`
}
