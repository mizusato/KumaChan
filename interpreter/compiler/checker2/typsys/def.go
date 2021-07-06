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
	Parameters  [] Parameter
	Content     TypeDefContent
	CaseInfo
}
type CaseInfo struct {
	Enum        *TypeDef
	CaseIndex   uint
}
type Parameter struct {
	Name      string
	Default   Type      // only available in type definition
	Variance  Variance  // only available in type definition
	Bound     Bound     // only available in function definition
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
	Value  Type  // is null when Kind = null | open*
}
type BoundKind int
const (
	NullBound BoundKind = iota
	SupBound
	InfBound
	OpenTopBound
	OpenBottomBound
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
	Methods  Record
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


