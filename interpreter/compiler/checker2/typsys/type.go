package typsys

import (
	"kumachan/interpreter/lang/common/attr"
)


type Type interface { _Type() }

func (*UnknownType) _Type() {}
type UnknownType struct {}

func (UnitType) _Type() {}
type UnitType struct {}

func (TopType) _Type() {}
type TopType struct {}

func (BottomType) _Type() {}
type BottomType struct {}

func (ParameterType) _Type() {}
type ParameterType struct {
	Parameter  *Parameter
}

func (*NestedType) _Type() {}
type NestedType struct {
	Content  NestedTypeContent
}


type NestedTypeContent interface { nestedTypeContent() }

func (Ref) nestedTypeContent() {}
type Ref struct {
	Def   *TypeDef
	Args  [] Type
}

func (Tuple) nestedTypeContent() {}
type Tuple struct {
	Elements  [] Type
}

func (Record) nestedTypeContent() {}
type Record struct {
	FieldIndexMap  map[string] uint
	Fields         [] Field
}
type Field struct {
	Attr  attr.FieldAttrs
	Name  string
	Type  Type
}

func (Lambda) nestedTypeContent() {}
type Lambda struct {
	Input   Type
	Output  Type
}


