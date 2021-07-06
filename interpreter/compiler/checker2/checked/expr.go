package checked

import (
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
)


type Expr struct {
	Type typsys.Type
	Info ExprInfo
	Expr ExprContent
}
type ExprInfo struct {
	Location  source.Location
}
func ExprInfoFrom(loc source.Location) ExprInfo {
	return ExprInfo { Location: loc }
}
type ExprContent interface { implExpr() }

func (FuncName) implExpr() {}
type FuncName struct {
	Name  name.FunctionName
}

func (LocalName) implExpr() {}
type LocalName struct {
	Name  string
}

func (Tuple) implExpr() {}
type Tuple struct {
	Elements  [] *Expr
}

func (InteriorRef) implExpr() {}
type InteriorRef struct {
	Base     Expr
	Index    uint
	Kind     InteriorRefKind
	Operand  InteriorRefOperand
}
type InteriorRefKind int
const (
	RK_Field InteriorRefKind = iota
	RK_Branch
)
type InteriorRefOperand int
const (
	RO_Record InteriorRefOperand = iota
	RO_Enum
	RO_ProjRef
	RO_CaseRef
)

func (NumericLiteral) implExpr() {}
type NumericLiteral struct {
	Value  interface {}
}


