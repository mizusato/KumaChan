package checker

import "kumachan/transformer/node"

type GenericType struct {
	Name      MaybeSymbol
	Value     TypeVal
	IsOpaque  bool
	Arity     int
}
func NewPlainGenericType(val TypeVal) *GenericType {
	return &GenericType {
		Name: nil,
		IsOpaque: false,
		Arity: 0,
		Value: val,
	}
}

type Type struct {
	Template   *GenericType
	Arguments  [] Type
}
func NewPlainType(t *GenericType) *Type {
	return &Type {
		Template: t,
		Arguments: make([]Type, 0),
	}
}

type TypeVal interface { TypeVal() }

func (impl TypePlaceholder) TypeVal() {}
type TypePlaceholder struct {
	Index  int
}
func NewTypePlaceholder(index int) *GenericType {
	return NewPlainGenericType(TypePlaceholder {
		Index: index,
	})
}

func (impl UnionType) TypeVal() {}
type UnionType struct {
	SubTypes  [] Symbol
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
	T_Bool   NativeTypeId   =  "Bool"
	T_Byte   NativeTypeId   =  "Byte"
	T_Word   NativeTypeId   =  "Word"
	T_Dword  NativeTypeId   =  "Dword"
	T_Qword  NativeTypeId   =  "Qword"
	// Number Types
	T_Int       NativeTypeId   =  "Int"
	T_Float     NativeTypeId   =  "Float"
	// Collection Types
	T_Bytes  NativeTypeId   =  "Bytes"
	T_Map    NativeTypeId   =  "Map"
	T_Stack  NativeTypeId   =  "Stack"
	T_Heap   NativeTypeId   =  "Heap"
	T_List   NativeTypeId   =  "List"
	// Effect Type
	T_Effect  NativeTypeId   =  "Effect"
)


type TypeExprError interface { TypeExprError() }
func (impl TE_TypeNotFound) TypeExprError() {}
type TE_TypeNotFound struct {
	Name  Symbol
	Node  node.Node
}
func (impl TE_NativeTypeNotFound) TypeExprError() {}
type TE_NativeTypeNotFound struct {
	Name  string
	Node  node.Node
}
func (impl TE_WrongArgumentQuantity) TypeExprError() {}
type TE_WrongArgumentQuantity struct {
	TypeName  Symbol
	Required  int
	Given     int
	Node      node.Node
}
func (impl TE_DuplicateFields) TypeExprError() {}
type TE_DuplicateFields struct {
	FieldName  string
	Field1     node.Node
	Field2     node.Node
}

type DeclaredTypeGetter func(Symbol) (*GenericType, bool)
type TypeExprContext struct {
	Mod     string
	Lookup  DeclaredTypeGetter
}

func GenericTypeFromDecl (decl *node.DeclType, ctx *TypeExprContext) (*GenericType, TypeExprError) {
	var arity = len(decl.Params)
	var args = make(map[Symbol] *GenericType)
	for i := 0; i < arity; i += 1 {
		args[SymbolFromId(&decl.Params[i], ctx.Mod)] = NewTypePlaceholder(i)
	}
	var self *GenericType
	var self_name = SymbolFromId(&decl.Name, ctx.Mod)
	var local_lookup = func(sym Symbol) (*GenericType, bool) {
		var t, exists = args[sym]
		if exists {
			return t, true
		} else if sym == self_name {
			return self, true
		} else {
			return ctx.Lookup(sym)
		}
	}
	var val, err = TypeValFrom(decl.TypeValue.TypeValue, &TypeExprContext {
		Mod:    ctx.Mod,
		Lookup: local_lookup,
	})
	if err != nil {
		return nil, err
	}
	self = &GenericType {
		Name: self_name,
		IsOpaque: decl.IsOpaque,
		Arity: arity,
		Value: val,
	}
	return self, nil
}

func TypeValFrom (val node.TypeValue, ctx *TypeExprContext) (TypeVal, TypeExprError) {
	switch value := val.(type) {
	case node.SingleType:
		return TypeValFromRepr(value.Repr.Repr, ctx)
	case node.UnionType:
		var u = &UnionType { SubTypes: make([]Symbol, len(value.Items)) }
		for i, item := range value.Items {
			u.SubTypes[i] = SymbolFromId(&item.Name, ctx.Mod)
		}
		return u, nil
	default:
		panic("invalid type value")
	}
}

func TypeValFromRepr (repr node.Repr, ctx *TypeExprContext) (TypeVal, TypeExprError) {
	switch r := repr.(type) {
	case node.ReprTuple:
		var tuple = &TupleType { Elements: make([]Type, len(r.Elements)) }
		for i, el := range r.Elements {
			var t, err = TypeFrom(el.Type, ctx)
			if err != nil { return nil, err }
			tuple.Elements[i] = *t
		}
		return *tuple, nil
	case node.ReprBundle:
		var bundle = &BundleType { Fields: make(map[string] Type) }
		var node_map = make(map[string] node.Node)
		for _, field := range r.Fields {
			var field_name = string(field.Name.Name)
			var _, exists = bundle.Fields[field_name]
			if exists {
				return nil, TE_DuplicateFields {
					FieldName: field_name,
					Field1: field.Node,
					Field2: node_map[field_name],
				}
			}
			node_map[field_name] = field.Node
			var field_type, err = TypeFrom(field.Type.Type, ctx)
			if err != nil { return nil, err }
			bundle.Fields[field_name] = *field_type
		}
		return *bundle, nil
	case node.ReprFunc:
		var input_type, err1 = TypeFrom(r.Input.Type, ctx)
		if err1 != nil { return nil, err1 }
		var output_type, err2 = TypeFrom(r.Output.Type, ctx)
		if err2 != nil { return nil, err2 }
		return FuncType {
			Input: *input_type,
			Output: *output_type,
		}, nil
	case node.ReprNative:
		var native_id = string(r.Ref.Id.Value)
		return NativeType { Id: NativeTypeId(native_id) }, nil
	default:
		panic("invalid type repr")
	}
}

func TypeFrom (type_ node.Type, ctx *TypeExprContext) (*Type, TypeExprError) {
	switch t := type_.(type) {
	case node.TypeRef:
		var sym = SymbolFromRef(&t.Ref, ctx.Mod)
		var referenced, exists = ctx.Lookup(sym)
		if exists {
			var ref_args = t.Ref.TypeArgs
			var result = &Type {
				Template: referenced,
				Arguments: make([]Type, len(ref_args)),
			}
			for i, arg := range ref_args {
				var arg_result, err = TypeFrom(arg.Type, ctx)
				if err != nil { return nil, err }
				result.Arguments[i] = *arg_result
			}
			return result, nil
		} else {
			return nil, TE_TypeNotFound { Name: sym, Node: t.Node }
		}
	case node.TypeLiteral:
		var val, err = TypeValFromRepr(t.Repr.Repr, ctx)
		if err != nil { return nil, err }
		return NewPlainType(NewPlainGenericType(val)), nil
	default:
		panic("invalid type")
	}
}