package kmd

import "math/big"


type Transformer struct {
	*Serializer
	*Deserializer
}

type Serializer struct {
	DetermineType  func(Object) *Type
	PrimitiveSerializer
	ContainerSerializer
	AlgebraicSerializer
}
type PrimitiveSerializer struct {
	WriteBool    func(Object) bool
	WriteFloat   func(Object) float64
	WriteUint32  func(Object) uint32
	WriteInt32   func(Object) int32
	WriteUint64  func(Object) uint64
	WriteInt64   func(Object) int64
	WriteInt     func(Object) *big.Int
	WriteString  func(Object) string
	WriteBinary  func(Object) ([] byte)
}
type ContainerSerializer struct {
	IterateArray    func(Object, func(uint,Object) error) error
	UnwrapOptional  func(Object) (Object, bool)
}
type AlgebraicSerializer struct {
	IterateRecord   func(Object, func(string,Object) error) error
	IterateTuple    func(Object, func(uint,Object) error) error
	Enum2Case       func(Object) Object
}

type Deserializer struct {
	PrimitiveDeserializer
	ContainerDeserializer
	AlgebraicDeserializer
}
type PrimitiveDeserializer struct {
	ReadBool    func(bool) Object
	ReadFloat   func(float64) Object
	ReadUint32  func(uint32) Object
	ReadInt32   func(int32) Object
	ReadUint64  func(uint64) Object
	ReadInt64   func(int64) Object
	ReadInt     func(*big.Int) Object
	ReadString  func(string) Object
	ReadBinary  func([] byte) Object
}
type ContainerDeserializer struct {
	CreateArray  func(array_t *Type) Object
	AppendItem   func(array_ptr *Object, item Object)
	Some         func(obj Object, opt_t *Type) Object
	Nothing      func(opt_t *Type) Object
}
type AlgebraicDeserializer struct {
	AssignObject    func(obj Object, from *Type, to *Type) (Object, error)
	CheckRecord     func(record_t TypeId, size uint) error
	GetFieldInfo    func(record_t TypeId, field string) (*Type, uint, error)
	CreateRecord    func(record_t TypeId) Object
	FillField       func(record Object, index uint, value Object)
	FinishRecord    func(record Object, t TypeId) (Object, error)
	CheckTuple      func(tuple_t TypeId, size uint) error
	GetElementType  func(tuple_t TypeId, element uint) *Type
	CreateTuple     func(tuple_t TypeId) Object
	FillElement     func(tuple Object, element uint, value Object)
	FinishTuple     func(tuple Object, t TypeId) (Object, error)
	Case2Enum       func(obj Object, enum_t TypeId, case_t TypeId) (Object, error)
}
