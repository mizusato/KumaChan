package checker

import (
	"kumachan/loader"
	"math/big"
)
import . "kumachan/error"


type Expr struct {
	Type   Type
	Info   ExprInfo
	Value  ExprVal
}

type ExprInfo struct {
	ErrorPoint  ErrorPoint
}

type ExprVal interface { ExprVal() }

func (impl Sum) ExprVal() {}
type Sum struct {
	Value  Expr
	Index  uint
}

func (impl Match) ExprVal() {}
type Match struct {
	Matched   Expr
	Branches  [] Branch
}
type Branch struct {
	Type    Type
	Pattern Pattern
	Value   Expr
}

func (impl Product) ExprVal() {}
type Product struct {
	Values  [] Expr
}

func (impl Get) ExprVal() {}
type Get struct {
	Product  Expr
	Index    uint
}

func (impl Set) ExprVal() {}
type Set struct {
	Product   Expr
	Index     uint
	NewValue  Expr
}

func (impl Call) ExprVal() {}
type Call struct {
	Caller  Expr
	Callee  Expr
}

func (impl Lambda) ExprVal() {}
type Lambda struct {
	Input   Pattern
	Output  Expr
}

func (impl Text) ExprVal() {}
type Text struct {
	Segments  [] TextSegment
}

func (impl UnitValue) ExprVal() {}
type UnitValue struct {}

func (impl Block) ExprVal() {}
type Block struct {
	Bindings  [] Binding
	Returned  Expr
}
type Binding struct {
	Pattern  Pattern
	Value    Expr
}

func (impl RefConst) ExprVal() {}
type RefConst struct {
	Name  loader.Symbol
}

func (impl RefFunction) ExprVal() {}
type RefFunction struct {
	Name   string
	Index  uint
}

func (impl RefLocal) ExprVal() {}
type RefLocal struct {
	Name  string
}

func (impl Array) ExprVal() {}
type Array struct {
	Items  [] Expr
}

func (impl IntLiteral) ExprVal() {}
type IntLiteral struct {
	Value  *big.Int
}

func (impl SmallIntLiteral) ExprVal() {}
type SmallIntLiteral struct {
	Kind   string
	Value  uint64
}

func (impl FloatLiteral) ExprVal() {}
type FloatLiteral struct {
	Value  float64
}

func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  [] rune
}


type Pattern struct {
	Point     ErrorPoint
	Concrete  ConcretePattern
}
type ConcretePattern interface { CheckerPattern() }
func (impl TrivialPattern) CheckerPattern() {}
type TrivialPattern struct {
	ValueName  string
	Point      ErrorPoint
}
func (impl TuplePattern) CheckerPattern() {}
type TuplePattern struct {
	ValueNames  [] string
	Points      [] ErrorPoint
}
func (impl BundlePattern) CheckerPattern() {}
type BundlePattern struct {
	ValueNames  [] string
	Points      [] ErrorPoint
}

type TextSegment interface { TextSegment() }
func (impl PlainSegment) TextSegment() {}
type PlainSegment struct {
	Content  [] rune
}
func (impl PlaceholderSegment) TextSegment() {}
type PlaceholderSegment struct {
	Type Type
	Key  string
}