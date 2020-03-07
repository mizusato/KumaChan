package node


type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}
func (impl Expr) ConstValue() {}
type Expr struct {
    Node                     `part:"expr"`
    Call     Call            `part:"call"`
    Pipeline MaybePipeline   `part_opt:"pipeline"`
}

type MaybeCall interface { MaybeCall() }
func (impl Call) MaybeCall() {}
type Call struct {
    Node                 `part:"call"`
    Func   VariousTerm   `part:"func.term"`
    Arg    MaybeCall     `part_opt:"arg.call"`
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
    Arg       MaybeCall       `part_opt:"pipe_arg.call"`
    Next      MaybePipeline   `part_opt:"pipeline"`
}

type PipeOperator struct {
    Node  `part:"pipe_op"`
}
