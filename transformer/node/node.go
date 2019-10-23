package node

import "kumachan/parser/scanner"

type Node struct {
    Span  scanner.Span
    Info  interface{}
}

type Module struct {
    Node
    MetaData  ModuleMetaData
    Imports   [] Import
}

type ModuleMetaData struct {
    Node
    Shebang     string
    Exported    map[string] string
    Resolving   map[string] ModuleSource
}

type ModuleSource struct {
    Node
    IsBuiltIn  bool
    Version    string
    URL        string
}

type Import struct {
    Node
    FromModule  string
    Names       [] ImportedName
}

type ImportedName struct {
    Node
    Name   string
    Alias  string
}

type Declaration interface { Declaration() }

func (impl Section) Declaration() {}
type Section struct {
    Node
    Name   string
    Decls  [] Declaration
}

func (impl Function) Declaration() {}
type Function struct {
    Node
    Name   string
    TP     [] TypeParameter
    Items  [] FunctionItem
}

type TypeParameter struct {
    Node
    Name        string
    Constraint  TraitExpr
}

type FunctionItem struct {
    Node
    Parameters  [] Parameter
    ReturnValue TypeExpr
    Body        FunctionBody
}

type TraitExpr struct {
    Node
    Module  string
    Name    string
    Args    [] TypeExpr
}

type Parameter struct {
    Node
    Name string
    Type TypeExpr
}

type FunctionBody struct {
    Node
    StaticBlock  [] Command
    Commands     [] Command
}

type TypeExpr struct {
    Node
    Content TypeExprContent
}

type TypeExprContent interface { TypeExprContent() }

func (impl OrdinaryTypeExpr) TypeExprContent() {}
type OrdinaryTypeExpr struct {
    Node
    Module  string
    Name    string
    Args    [] TypeExpr
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
    YieldType TypeExpr
}

func (impl ContinuationTypeExpr) TypeExprContent() {}
type ContinuationTypeExpr struct {
    Node
    ResumingType TypeExpr
}

type AttachedExpr struct {
    Node
    Type         TypeExpr
    AttachedName string
}

type Signature struct {
    Node
    Parameters  [] TypeExpr
    ReturnValue TypeExpr
}

type TypeDeclaration interface {
    Declaration
    TypeDeclaration()
}

func (impl SingletonTypes) Declaration() {}
func (impl SingletonTypes) TypeDeclaration() {}
type SingletonTypes struct {
    Node
    Names  [] string
}

func (impl UnionType) Declaration() {}
func (impl UnionType) TypeDeclaration() {}
type UnionType struct {
    Node
    Name      string
    Elements  [] TypeExpr
}

func (impl SchemaType) Declaration() {}
func (impl SchemaType) TypeDeclaration() {}
type SchemaType struct {
    Node
    Name    string
    Attrs   SchemaAttrs
    TP      [] TypeParameter
    Bases   [] TypeExpr
    Fields  [] SchemaField
}

func (impl ClassType) Declaration() {}
func (impl ClassType) TypeDeclaration() {}
type ClassType struct {
    Node
    Name     string
    Attrs    ClassAttrs
    TP       [] TypeParameter
    Bases    [] TypeExpr
    Impls    [] TypeExpr
    Init     ClassInit
    PF       map[string] FunctionItem
    Methods  map[string] FunctionItem
}

func (impl InterfaceType) Declaration() {}
func (impl InterfaceType) TypeDeclaration() {}
type InterfaceType struct {
    Node
    Name     string
    Attrs    InterfaceAttrs
    TP       [] TypeParameter
    Methods  map[string] MethodSignature
}

type SchemaAttrs struct {
    Mutable     bool
    Extensible  bool
}

type SchemaField struct {
    Node
    Name    string
    Type    TypeExpr
    Default Expr
}

type ClassAttrs struct {
    Extensible  bool
}

type ClassInit struct {
    Node
    Parameters  [] Parameter
    Body        FunctionBody
}

type InterfaceAttrs struct {
    IsNative  bool
}

type MethodSignature struct {
    Node
    Parameters  [] Parameter
    ReturnValue TypeExpr
}

type AttachedDeclaration interface {
    Declaration
    AttachedDeclaration()
}

func (impl AttachedType) Declaration() {}
func (impl AttachedType) AttachedDeclaration() {}
type AttachedType struct {
    Node
    Target   AttachedExpr
    Attached TypeExpr
}

func (impl AttachedFunction) Declaration() {}
func (impl AttachedFunction) AttachedDeclaration() {}
type AttachedFunction struct {
    Node
    Target    AttachedExpr
    Attached  FunctionItem
}

func (impl AttachedValue) Declaration() {}
func (impl AttachedValue) AttachedDeclaration() {}
type AttachedValue struct {
    Node
    Target    AttachedExpr
    Attached  Expr
}

func (impl TraitDeclaration) Declaration() {}
type TraitDeclaration struct {
    Node
    Name        string
    TP          [] TypeParameter
    ArgName     string
    Bases       [] TraitDeclaration
    Constraint  TraitConstraint
}

type TraitConstraint interface {
    TraitConstraint()
}

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

type Block struct {
    Node
    Commands  [] Command
}

type Identifier struct {
    Node
    Name string
}

type Expr struct {
    Node
    Content  ExprContent
}

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
    Value  interface{}
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

type MaybeTypeExpr interface { MaybeTypeExpr() }
func (impl TypeExpr) MaybeTypeExpr() {}

type MaybeExpr interface { MaybeExpr() }
func (impl Expr) MaybeExpr() {}

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
    DestructedNames  [] string
}