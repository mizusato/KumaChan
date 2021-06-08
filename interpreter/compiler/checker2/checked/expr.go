package checked

import (
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/lang/common/name"
)


type Expr struct {
	Type typsys.Type
	Info ExprInfo
	Expr ExprContent
}
type ExprInfo struct {
	Point  ErrorPoint
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


