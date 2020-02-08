package types

// TODO: remove this file: do not distinguish between native type representations

type NativeTypeId uint
const (
	// Basic Types
	T_Bit  NativeTypeId  =  iota
	T_Byte
	T_Word
	T_Dword
	T_Qword
	T_Int
	// Collection Types
	T_Bytes
	T_String
	T_Array
	T_Stack
	T_Queue
	T_Set
	// Effect Type
	T_Effect
	// I/O Types
	T_InputStream
	T_OutputStream
)

var __NativeTypes = map[string] NativeTypeId {
	// Core
	"Bit":     T_Bit,
	"Byte":    T_Byte,
	"Word":    T_Word,
	"Dword":   T_Dword,
	"Qword":   T_Qword,
	"Int":     T_Int,
	"Bytes":   T_Bytes,
	"String":  T_String,
	"Array":   T_Array,
	"Stack":   T_Stack,
	"Queue":   T_Queue,
	"Set":     T_Set,
	"Effect":  T_Effect,
	// I/O
	"InputStream": T_InputStream,
	"OutputStream": T_OutputStream,
}

func GetNativeTypeId (str_id string) (NativeTypeId, bool) {
	var id, exists = __NativeTypes[str_id]
	return id, exists
}