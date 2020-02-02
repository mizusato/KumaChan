package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
)

type GenericType struct {
	Arity     uint
	IsOpaque  bool
	Value     TypeVal
	Node      node.Node
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

func (impl Nil) TypeRepr() {}
type Nil struct {}

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
	Id  NativeTypeId
}

type NativeTypeId uint
const (
	// Basic Types
	T_Bool   NativeTypeId   =  iota
	T_Byte
	T_Word
	T_Dword
	T_Qword
	// Number Types
	T_Int
	T_Float
	// Collection Types
	T_Bytes
	T_Map
	T_Stack
	T_Heap
	T_List
	// Effect Type
	T_Effect
)

var NativeTypes = map[string] NativeTypeId {
	"Bool": T_Bool,
	"Byte": T_Byte,
	"Word": T_Word,
	"Dword": T_Dword,
	"Qword": T_Qword,
	"Int": T_Int,
	"Float": T_Float,
	"Bytes": T_Bytes,
	"Map": T_Map,
	"Stack": T_Stack,
	"Heap": T_Heap,
	"List": T_List,
	"Effect": T_Effect,
}