package ast


type MaybeExpr interface { Maybe(Expr,MaybeExpr) }
func (impl Expr) Maybe(Expr,MaybeExpr)  {}
func (impl Expr) ConstValue() {}
type Expr struct {
    Node                       `part:"expr"`
    Term      VariousTerm      `part:"term"`
    Pipeline  [] VariousPipe   `list_rec:"pipes"`
}
func WrapTermAsExpr(term VariousTerm) Expr {
    return Expr {
        Node:     term.Node,
        Term:     term,
        Pipeline: nil,
    }
}

type VariousTerm struct {
    Node         `part:"term"`
    Term  Term   `use:"first"`
}
type Term interface { Term() }

func (impl VariousCall) Term() {}
type VariousCall struct {
    Node         `part:"call"`
    Call  Call   `use:"first"`
}
type Call interface { Call() }
func (impl CallPrefix) Call() {}
type CallPrefix struct {
    Node             `part:"call_prefix"`
    Callee    Expr   `part:"callee.expr"`
    Argument  Expr   `part:"expr"`
}
func (impl CallInfix) Call() {}
type CallInfix struct {
    Node             `part:"call_infix"`
    Operator  Expr   `part:"operator.expr"`
    Left      Expr   `part:"infix_left.expr"`
    Right     Expr   `part:"infix_right.expr"`
}

type VariousPipe struct {
    Node         `part:"pipe"`
    Pipe  Pipe   `use:"first"`
}
type Pipe interface { Pipe() }
func (impl PipeFunc) Pipe() {}
type PipeFunc struct {
    Node                  `part:"pipe_func"`
    Callee    Expr        `part:"callee.expr"`
    Argument  MaybeExpr   `part_opt:"pipe_func_arg.expr"`
}
func (impl PipeGet) Pipe() {}
type PipeGet struct {
    Node               `part:"pipe_get"`
    Key   Identifier   `part:"name"`
}
func (impl PipeCast) Pipe() {}
type PipeCast struct {
    Node                  `part:"pipe_cast"`
    Target  VariousType   `part:"type"`
}


