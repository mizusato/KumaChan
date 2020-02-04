package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	"kumachan/native/types"
)

type GenericType struct {
	Arity     uint
	IsOpaque  bool
	Value     TypeVal
	Node      node.Node
	Order     uint
}

type TypeVal interface { TypeVal() }

func (impl UnionTypeVal) TypeVal() {}
type UnionTypeVal struct {
	SubTypes  [] loader.Symbol
}

func (impl SingleTypeVal) TypeVal() {}
type SingleTypeVal struct {
	Expr  TypeExpr
}

type TypeExpr interface { TypeExpr() }

func (impl ParameterType) TypeExpr() {}
type ParameterType struct {
	Index  uint
}

func (impl NamedType) TypeExpr() {}
type NamedType struct {
	Name  loader.Symbol
	Args  [] TypeExpr
}

func (impl AnonymousType) TypeExpr() {}
type AnonymousType struct {
	Repr  TypeRepr
}

type TypeRepr interface { TypeRepr() }

func (impl Unit) TypeRepr() {}
type Unit struct {}

func (impl Tuple) TypeRepr() {}
type Tuple struct {
	Elements  [] TypeExpr
}

func (impl Bundle) TypeRepr() {}
type Bundle struct {
	Fields  map[string] TypeExpr
}

func (impl Func) TypeRepr() {}
type Func struct {
	Input   TypeExpr
	Output  TypeExpr
}

func (impl NativeType) TypeRepr() {}
type NativeType struct {
	Id  types.NativeTypeId
}
