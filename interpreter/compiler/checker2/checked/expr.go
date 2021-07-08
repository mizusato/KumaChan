package checked

import (
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
)


type Expr struct {
	Type     typsys.Type
	Info     ExprInfo
	Content  ExprContent
}
type ExprInfo struct {
	Location  source.Location
}
func ExprInfoFrom(loc source.Location) ExprInfo {
	return ExprInfo { Location: loc }
}
type ExprContent interface { implExpr() }

func (FuncRef) implExpr() {}
type FuncRef struct {
	Name  name.FunctionName
}

func (LocalRef) implExpr() {}
type LocalRef struct {
	Binding  *LocalBinding
}

func (Tuple) implExpr() {}
type Tuple struct {
	Elements  [] *Expr
}

func (TupleUpdate) implExpr() {}
type TupleUpdate struct {
	Base      *Expr
	Replaced  [] TupleUpdateElement
}
type TupleUpdateElement struct {
	Index  uint
	Value  *Expr
}

func (InteriorRef) implExpr() {}
type InteriorRef struct {
	Base     *Expr
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


// TODO: consider separate following types to another file

type LocalBinding struct {
	Name     string
	Type     typsys.Type
	Location source.Location
}

type ProductPatternInfo ([]ProductPatternItemInfo)
type ProductPatternItemInfo struct {
	Binding *LocalBinding
	Index1  uint // 0 = whole, 1 = .0
}

func (Lambda) implExpr() {}
type Lambda struct {
	In   ProductPatternInfo
	Out  *Expr
}


