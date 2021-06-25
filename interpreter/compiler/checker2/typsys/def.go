package typsys

import (
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
)


type TypeDef struct {
	attr.TypeAttrs
	Name        name.TypeName
	Implements  [] *TypeDef
	Tables      [] DispatchTable
	Parameters  [] Parameter
	Content     TypeDefContent
	CaseInfo
}
type CaseInfo struct {
	Enum        *TypeDef
	CaseIndex   uint
}
type DispatchTable struct {
	Interface  *Interface
	Methods    [] name.FunctionName
}
type Parameter struct {
	Name      string
	Default   Type  // nullable
	Variance  Variance
	Bound     Bound
}
type Variance int
const (
	Invariant Variance = iota
	Covariant
	Contravariant
	Bivariant
)
type Bound struct {
	Kind   BoundKind
	Value  Type  // is null when Kind = NullBound
}
type BoundKind int
const (
	NullBound BoundKind = iota
	SupBound
	InfBound
)
func (def *TypeDef) ForEachParameter(f func(uint,*Parameter)(*source.Error)) *source.Error {
	for i := range def.Parameters {
		var p = &(def.Parameters[i])
		var err = f(uint(i), p)
		if err != nil { return err }
	}
	return nil
}


type TypeDefContent interface { typeDef() }

func (*Enum) typeDef() {}
type Enum struct {
	CaseTypes  [] *TypeDef
}

func (*Interface) typeDef() {}
type Interface struct {
	Included  [] *Interface
	Methods   Record
}

func (*Box) typeDef() {}
type Box struct {
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


