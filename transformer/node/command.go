package node


type Command interface { Command() }

func (impl IfCommand) Command() {}
type IfCommand struct {
    Node
    Condition  Expr
    Block      Block
    Elifs      [] Elif
    Else       Else
}

func (impl WhileCommand) Command() {}
type WhileCommand struct {
    Node
    Condition  Expr
    Block      Block
}

func (impl ForCommand) Command() {}
type ForCommand struct {
    Node
    LoopVars  [] Identifier
    Iterator  Expr
    Block     Block
}

func (impl BreakCommand) Command() {}
type BreakCommand struct {
    Node
}

func (impl ContinueCommand) Command() {}
type ContinueCommand struct {
    Node
}

func (impl ReturnCommand) Command() {}
type ReturnCommand struct {
    Node
    Value  Expr
}

func (impl YieldCommand) Command() {}
type YieldCommand struct {
    Node
    Value  Expr
}

func (impl PanicCommand) Command() {}
type PanicCommand struct {
    Node
    ErrorValue  Expr
}

func (impl AssertCommand) Command() {}
type AssertCommand struct {
    Node
    Condition  Expr
}

func (impl FinallyCommand) Command() {}
type FinallyCommand struct {
    Node
    Block  Block
}

func (impl LetCommand) Command() {}
type LetCommand struct {
    Node
    Variable Identifier
    Value    Expr
}

func (impl InitialCommand) Command()  {}
type InitialCommand struct {
    Node
    Variable     Identifier
    InitialValue Expr
}

func (impl ResetCommand) Command() {}
type ResetCommand struct {
    Node
    Variable  Identifier
    NewValue  Expr
}

func (impl PassCommand) Command() {}
type PassCommand struct {
    Node
}

func (impl SideEffectCommand) Command() {}
type SideEffectCommand struct {
    Node
    Expr  Expr
}


type Elif struct {
    Node
    Condition  Expr
    Block      Block
}

type Else struct {
    Node
    Block      Block
}

