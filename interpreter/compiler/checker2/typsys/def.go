package typsys

import (
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/name"
)


type TypeDef struct {
	Attr        attr.TypeAttr
	Name        name.TypeName
	Implements  [] DispatchTable
	Parameters  [] Parameter
	Content     TypeDefContent
	CaseInfo
}
type CaseInfo struct {
	IsCaseType  bool
	Enum        *TypeDef
	CaseIndex   uint
	CaseParams  [] uint
}
type DispatchTable struct {
	Interface  *TypeDef  // Content should be an *Interface
	Methods    [] name.FunctionName
}
type Parameter struct {
	Name      string
	Variance  Variance
	SupBound  Type  // nullable
	InfBound  Type  // nullable
	Default   Type  // nullable
}
type Variance int
const (
	Invariant Variance = iota
	Covariant
	Contravariant
	Bivariant
)


type TypeDefContent interface { typeDef() }

func (*Enum) typeDef() {}
type Enum struct {
	CaseTypes  [] *TypeDef
}

func (*Interface) typeDef() {}
type Interface struct {
	Included  [] IncludedInterface
	Content   Record
}
type IncludedInterface struct {
	Interface  *TypeDef
	Content    [] uint
}

func (*Boxed) typeDef() {}
type Boxed struct {
	BoxKind       BoxKind
	WeakWrapping  bool
	InnerType     Type
}
type BoxKind int
const (
	Isomorphic BoxKind = iota
	Protected
	Opaque
)

func (*Native) typeDef() {}
type Native struct {}


