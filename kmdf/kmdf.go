package kmdf

import "math/big"


type Object = interface{}
type Binary = [] byte
type Type struct {
	Kind         TypeKind
	ElementType  *Type   // only available in Container Types
	Identifier   TypeId  // only available in Algebraic Types
}
type TypeKind uint
const (
	// Primitive Types
	Bool TypeKind = iota
	Number
	Float
	Uint64
	Int64
	Int
	String
	// Container Types
	Array
	Map
	Optional
	Result
	// Algebraic Types
	Record
	Tuple
	Union
)
type TypeId struct {
	Vendor   string
	Name     string
	Version  string
}

type Serializer struct {
	DetermineType  func(Object) *Type
	PrimitiveSerializer
	ContainerSerializer
	AlgebraicSerializer
}
type PrimitiveSerializer struct {
	WriteBool    func(Object) bool
	WriteNumber  func(Object) uint
	WriteFloat   func(Object) float64
	WriteUint64  func(Object) uint64
	WriteInt64   func(Object) int64
	WriteInt     func(Object) *big.Int
	WriteString  func(Object) string
}
type ContainerSerializer struct {
	GetArrayLength  func(Object) uint
	IterateArray    func(Object, func(uint,Object))
	GetMapSize      func(Object) uint
	IterateMap      func(Object, func(string,Object))
	UnwrapOptional  func(Object) (Object, bool)
}
type AlgebraicSerializer struct {
	IterateRecord   func(Object, func(string,Object))
	IterateTuple    func(Object, func(Object))
	Union2Case      func(Object) Object
}

type Deserializer struct {
	PrimitiveDeserializer
	ContainerDeserializer
	AlgebraicDeserializer
}
type PrimitiveDeserializer struct {
	ReadBool    func(bool) Object
	ReadNumber  func(uint) Object
	ReadFloat   func(float64) Object
	ReadUint64  func(uint64) Object
	ReadInt64   func(int64) Object
	ReadInt     func(*big.Int) Object
	ReadString  func(string) Object
}
type ContainerDeserializer struct {
	CreateArray  func() Object
	AppendItem   func(array Object, item Object)
	CreateMap    func() Object
	InsertItem   func(map_ Object, key Object, value Object)
	Just         func(Object) Object
	Nothing      Object
}
type AlgebraicDeserializer struct {
	AssignObject    func(obj Object, from *Type, to *Type) (Object, error)
	CheckRecord     func(record_t TypeId, size uint) error
	GetFieldType    func(record_t TypeId, field string) (*Type, error)
	CreateRecord    func(record_t TypeId) (Object, uint)
	FillField       func(record Object, field string, value Object)
	CheckTuple      func(tuple_t TypeId, size uint) error
	GetElementType  func(tuple_t TypeId, element uint) (*Type, error)
	CreateTuple     func(tuple_t TypeId) (Object, uint)
	FillElement     func(tuple Object, element uint, value Object)
	Case2Union      func(obj Object, union_t TypeId, case_t TypeId) (Object, error)
}