package checker

type Type struct {
	Id        Symbol
	Value     TypeVal
	IsOpaque  bool
}

type TypeVal interface { TypeVal() }

func (impl UnionType) TypeVal() {}
type UnionType struct {
	SubTypes  [] Type
}

func (impl TupleType) TypeVal() {}
type TupleType struct {
	Elements  [] Type
}

func (impl BundleType) TypeVal() {}
type BundleType struct {
	Fields  map[string] Type
}

func (impl FuncType) TypeVal() {}
type FuncType struct {
	Input   Type
	Output  Type
}

func (impl NativeType) TypeVal() {}
type NativeType struct {
	Id  NativeTypeId
}

type NativeTypeId string
const (
	// Basic Types
	T_Bool      =  "Bool"
	T_Byte      =  "Byte"
	T_Word      =  "Word"
	T_Dword     =  "Dword"
	T_Qword     =  "Qword"
	// Number Types
	T_Int  NativeTypeId  =  "Int"
	T_Float     =  "Float"
	T_Complex   =  "Complex"
	T_BigInt    =  "BigInt"
	T_BigFloat  =  "BigFloat"
	// Collection Types
	T_Bytes     =  "Bytes"
	T_Map       =  "Map"
	T_Stack     =  "Stack"
	T_Heap      =  "Heap"
	T_List      =  "List"
)