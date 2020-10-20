package kmd

import "fmt"


type Object = interface{}
type Type struct {
	Kind         TypeKind
	ElementType  *Type   // only available in Container Types
	Identifier   TypeId  // only available in Algebraic Types
}
type TypeKind uint
const (
	// Primitive Types
	Bool TypeKind = iota; Float; Uint32; Int32; Uint64; Int64; Int; String; Binary
	// Container Types
	Array; Optional
	// Algebraic Types
	Record; Tuple; Union
)
type TypeId struct {
	TypeIdBase
	Version  string
}
type TypeIdBase struct {
	Vendor   string
	Name     string
}

var __PrimitiveTypes = make(map[TypeKind] *Type)
var __AlgebraicTypes = make(map[TypeId] *Type)
func PrimitiveType(kind TypeKind) *Type {
	var existing, exists = __PrimitiveTypes[kind]
	if exists {
		return existing
	} else {
		var t = &Type { Kind: kind }
		__PrimitiveTypes[kind] = t
		return t
	}
}
func ContainerType(kind TypeKind, elem *Type) *Type {
	return &Type { Kind: kind, ElementType: elem }
}
func AlgebraicType(kind TypeKind, id TypeId) *Type {
	var existing, exists = __AlgebraicTypes[id]
	if exists {
		return existing
	} else {
		var t = &Type { Kind: kind, Identifier: id }
		__AlgebraicTypes[id] = t
		return t
	}
}

func TypeEqual(t1 *Type, t2 *Type) bool {
	if t1 == nil && t2 == nil { return true }
	if t1 == nil || t2 == nil { return false }
	if t1.Kind != t2.Kind { return false }
	if t1.Identifier != t2.Identifier { return false }
	if !(TypeEqual(t1.ElementType, t2.ElementType)) { return false }
	return true
}

func (kind TypeKind) String() string {
	switch kind {
	case Bool:     return "bool"
	case Float:    return "float"
	case Uint32:   return "uint32"
	case Int32:    return "int32"
	case Uint64:   return "uint64"
	case Int64:    return "int64"
	case Int:      return "int"
	case String:   return "string"
	case Binary:   return "binary"
	case Array:    return "[]"
	case Optional: return "?"
	case Record:   return "{}"
	case Tuple:    return "()"
	case Union:    return "|"
	default:       panic("impossible branch")
	}
}
func (id TypeId) String() string {
	return fmt.Sprintf("%s.%s %s", id.Vendor, id.Name, id.Version)
}
func (t *Type) String() string {
	if t.ElementType != nil {
		return fmt.Sprintf("%s %s", t.Kind, t.ElementType)
	} else if t.Identifier != (TypeId {}) {
		return fmt.Sprintf("%s %s", t.Kind, t.Identifier)
	} else {
		return fmt.Sprintf("%s", t.Kind)
	}
}