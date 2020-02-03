package types

type NativeTypeId uint
const (
	// Basic Types
	T_Bit  NativeTypeId  =  iota
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

var __NativeTypes = map[string] NativeTypeId {
	"Bit":  T_Bit,
	"Byte":  T_Byte,
	"Word":  T_Word,
	"Dword": T_Dword,
	"Qword": T_Qword,
	"Int":   T_Int,
	"Float": T_Float,
	"Bytes": T_Bytes,
	"Map":   T_Map,
	"Stack": T_Stack,
	"Heap":  T_Heap,
	"List": T_List,
	"Effect": T_Effect,
}

func GetNativeTypeId (str_id string) (NativeTypeId, bool) {
	var id, exists = __NativeTypes[str_id]
	return id, exists
}