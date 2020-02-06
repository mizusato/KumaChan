package types

type NativeTypeId uint
const (
	// Basic Types
	T_Bit  NativeTypeId  =  iota
	T_Byte
	T_Word
	T_Dword
	T_Qword
	// Collection Types
	T_Bytes
	T_Map
	T_Stack
	T_Heap
	T_List
	// Effect Type
	T_Effect
	// I/O Types
	T_InputStream
	T_OutputStream
)

var __NativeTypes = map[string] NativeTypeId {
	// Core
	"Bit":  T_Bit,
	"Byte":  T_Byte,
	"Word":  T_Word,
	"Dword": T_Dword,
	"Qword": T_Qword,
	"Bytes": T_Bytes,
	"Map":   T_Map,
	"Stack": T_Stack,
	"Heap":  T_Heap,
	"List": T_List,
	"Effect": T_Effect,
	// I/O
	"InputStream": T_InputStream,
	"OutputStream": T_OutputStream,
}

func GetNativeTypeId (str_id string) (NativeTypeId, bool) {
	var id, exists = __NativeTypes[str_id]
	return id, exists
}