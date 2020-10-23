package kmd

import (
	"fmt"
	"strings"
)


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
func TypeParse(text string) (*Type, bool) {
	if stringHasFirstSegment(text, "[]") ||
		stringHasFirstSegment(text, "?") {
		var kind_text, elem_text = stringSplitFirstSegment(text)
		var kind TypeKind
		switch kind_text {
		case "[]": kind = Array
		case "?":  kind = Optional
		default:   panic("impossible branch")
		}
		var elem_t, ok = TypeParse(elem_text)
		if !(ok) { return nil, false }
		return ContainerType(kind, elem_t), true
	} else if stringHasFirstSegment(text, "{}") ||
		stringHasFirstSegment(text, "()") ||
		stringHasFirstSegment(text, "|") {
		var kind_text, id_text = stringSplitFirstSegment(text)
		var kind TypeKind
		switch kind_text {
		case "{}": kind = Record
		case "()": kind = Tuple
		case "|":  kind = Union
		default:   panic("impossible branch")
		}
		var vendor, name, version string
		var vendor_name_text, version_text = stringSplitFirstSegment(id_text)
		var temp = strings.Split(vendor_name_text, ".")
		if !(len(temp) >= 2) { return nil, false }
		// TODO: some format validation
		vendor = strings.Join(temp[:len(temp)-1], ".")
		name = temp[len(temp)-1]
		version = version_text
		var id = TheTypeId(vendor, name, version)
		return AlgebraicType(kind, id), true
	} else {
		var kind_text = text
		switch kind_text {
		case "bool":   return PrimitiveType(Bool), true
		case "float":  return PrimitiveType(Float), true
		case "uint32": return PrimitiveType(Uint32), true
		case "int32":  return PrimitiveType(Int32), true
		case "uint64": return PrimitiveType(Uint64), true
		case "int64":  return PrimitiveType(Int64), true
		case "int":    return PrimitiveType(Int), true
		case "string": return PrimitiveType(String), true
		case "binary": return PrimitiveType(Binary), true
		default:       return nil, false
		}
	}
}
func stringHasFirstSegment(str string, segment string) bool {
	for i := 0; i < len(str); i += 1 {
		if rune(str[i]) == ' ' {
			if str[:i] == segment {
				return true
			} else {
				return false
			}
		}
	}
	return false
}
func stringSplitFirstSegment(str string) (string, string) {
	for i := 0; i < len(str); i += 1 {
		if rune(str[i]) == ' ' {
			if (i + 1) < len(str) {
				return str[:i], str[(i+1):]
			} else {
				return str[:i], ""
			}
		}
	}
	return "", str
}

func TheTypeId(vendor string, name string, version string) TypeId {
	return TypeId {
		TypeIdBase: TypeIdBase {
			Vendor: vendor,
			Name:   name,
		},
		Version:    version,
	}
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