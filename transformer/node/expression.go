package node

import "math/big"

type Expr struct {
    Node
    Content  ExprContent
}

type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}

type ExprContent interface { ExprContent() }

func (impl UnaryApply) ExprContent()  {}
type UnaryApply struct {
    Node
    OpName   string
    Operand  Expr
}

func (impl InfixApply) ExprContent() {}
type InfixApply struct {
    Node
    OpName  string
    Left    Expr
    Right   Expr
}

func (impl MemberAccess) ExprContent() {}
type MemberAccess struct {
    Node
    Target  Expr
    Member  Identifier
    Args    [] Expr
}

func (impl Call) ExprContent() {}
type Call struct {
    Node
    Target  Expr
    Args    [] Expr
}

func (impl With) ExprContent() {}
type With struct {
    Node
    Former  Expr
    Added   [] StructField
}

func (impl Lambda) ExprContent() {}
type Lambda struct {
    Node
    Parameters   [] LambdaParameter
    ReturnValue  MaybeTypeExpr
    Body         FunctionBody
}

func (impl Generator) ExprContent() {}
type Generator struct {
    Node
    YieldType  MaybeTypeExpr
    Body       FunctionBody
}

func (impl Cast) ExprContent() {}
type Cast struct {
    Node
    Value    Expr
    Target   TypeExpr
    Dynamic  bool
}

func (impl Callcc) ExprContent() {}
type Callcc struct {
    Node
    ResumingType    TypeExpr
    CalledFunction  Expr
}

func (impl TypeObject) ExprContent() {}
type TypeObject struct {
    Node
    Type  TypeExpr
}

func (impl Constructor) ExprContent() {}
type Constructor struct {
    Node
    Class  TypeExpr
}

func (impl AttachedValueExpr) ExprContent() {}
type AttachedValueExpr struct {
    Node
    AttachedExpr  AttachedExpr
}

func (impl StructLiteral) ExprContent() {}
type StructLiteral struct {
    Node
    Schema  TypeExpr
    Fields  [] StructField
}

func (impl TupleLiteral) ExprContent() {}
type TupleLiteral struct {
    Node
    TypeArgs  [] TypeExpr
    Elements  [] Expr
}

func (impl ConstLiteral) ExprContent() {}
type ConstLiteral struct {
    Node
    Content  ConstContent
}
type ConstContent interface { ConstContent() }
func (impl StringLiteral) ConstContent() {}
type StringLiteral struct {
    Node
    Value  string
}
func (impl IntegerLiteral) ConstContent() {}
type IntegerLiteral struct {
    Node
    Value  int
}
func (impl FloatLiteral) ConstContent() {}
type FloatLiteral struct {
    Node
    Value  float64
}
func (impl BooleanLiteral) ConstContent() {}
type BooleanLiteral struct {
    Node
    Value  bool
}
func (impl BigIntLiteral) ConstContent() {}
type BigIntLiteral struct {
    Node
    Value  big.Int
}
func (impl BigFloatLiteral) ConstContent() {}
type BigFloatLiteral struct {
    Node
    Value  big.Float
}

func (impl Text) ExprContent() {}
type Text struct {
    Node
    Segments  [] TextSegment
}

func (impl When) ExprContent() {}
type When struct {
    Node
    Branches  [] WhenBranch
}

func (impl Match) ExprContent() {}
type Match struct {
    Node
    Input     Expr
    Branches  [] MatchBranch
}

func (impl Variable) ExprContent() {}
type Variable struct {
    Node
    Module    string
    Name      string
    TypeArgs  [] TypeExpr
}


type StructField struct {
    Node
    Name   Identifier
    Value  Expr
}

type LambdaParameter struct {
    Node
    Name  string
    Type  MaybeTypeExpr
}

type TextSegment interface {}

func (impl StringTextSegment) TextSegment() {}
type StringTextSegment struct {
    Node
    String  string
}

func (impl ExprTextSegment) TextSegment() {}
type ExprTextSegment struct {
    Node
    Tag   string
    Expr  Expr
}

type WhenBranch struct {
    Node
    Condition  MaybeExpr
    Value      BranchValue
}

type MatchBranch struct {
    Node
    Condition  MaybeTypeExpr
    Pattern    MatchPattern
    Value      BranchValue
}

type BranchValue interface { BranchValue() }

func (impl ExprBranchValue) BranchValue() {}
type ExprBranchValue struct {
    Node
    Expr  Expr
}

func (impl BlockBranchValue) BranchValue() {}
type BlockBranchValue struct {
    Node
    Block  Block
}

type MatchPattern interface { MatchPattern() }

func (impl PlainMatchPattern) MatchPattern() {}
type PlainMatchPattern struct {
    Node
    Matched  Identifier
}

func (impl DestructionMatchPattern) MatchPattern() {}
type DestructionMatchPattern struct {
    Node
    DestructedNames  [] Identifier
}
