package ast


type MaybeCall interface { MaybeCall() }
func (impl Call) MaybeCall() {}
func (impl Call) Term() {}
type Call struct {
    Node
    Func  VariousTerm
    Arg   MaybeCall
}
func WrapCallAsTerm(call Call) VariousTerm {
    return VariousTerm {
        Node: call.Node,
        Term: call,
    }
}
func WrapCallAsExpr(call Call) Expr {
    return Expr {
        Node:     call.Node,
        Terms:    Terms {
            Node:  call.Node,
            Terms: [] VariousTerm { WrapCallAsTerm(call) },
        },
        Pipeline: nil,
    }
}

type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr()  {}
func (impl Expr) ConstValue() {}
type Expr struct {
    Node                      `part:"expr"`
    Terms     Terms           `part:"terms"`
    Pipeline  MaybePipeline   `part_opt:"pipeline"`
}
func WrapTermAsExpr(term VariousTerm) Expr {
    return Expr {
        Node:     term.Node,
        Terms:    Terms {
            Node:  term.Node,
            Terms: [] VariousTerm { term },
        },
        Pipeline: nil,
    }
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
    Arg       MaybeTerms      `part_opt:"pipe_arg.terms"`
    Next      MaybePipeline   `part_opt:"pipeline"`
}

type PipeOperator struct {
    Node  `part:"pipe_op"`
}
