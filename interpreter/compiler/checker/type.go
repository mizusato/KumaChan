package checker

import (
	"kumachan/interpreter/def"
	"kumachan/interpreter/parser/ast"
)


type GenericType struct {
	Section     string
	Node        ast.Node
	Doc         string
	Tags        TypeTags
	Params      [] TypeParam
	Bounds      TypeBounds
	Defaults    map[uint] Type
	Definition  TypeDef
	CaseInfo    CaseInfo
	FieldInfo   map[string] FieldInfo
}
type TypeParam struct {
	Name      string
	Variance  TypeVariance
}
var __NoParams = make([] TypeParam, 0)
func TypeParamsNames(params ([] TypeParam)) ([] string) {
	var draft = make([] string, len(params))
	for i, _ := range draft {
		draft[i] = params[i].Name
	}
	return draft
}
type FieldInfo struct {
	Doc   string
	Tags  FieldTags
}


type TypeDef interface { CheckerTypeDef() }

func (impl *Enum) CheckerTypeDef() {}
type Enum struct {
	CaseTypes  [] CaseType
}
type CaseType struct {
	Name   def.Symbol
	Params [] uint
}
func (impl *Boxed) CheckerTypeDef() {}
type Boxed struct {
	InnerType  Type
	Implicit   bool
	Weak       bool
	// following properties are exclusive
	Protected  bool
	Opaque     bool
}
func (impl *Native) CheckerTypeDef() {}
type Native struct {}


type TypeContext struct {
	TypeBoundsContext
}
type Type interface { CheckerType() }

func (impl *ParameterType) CheckerType() {}
type ParameterType struct {
	Index          uint
	BeingInferred  bool
}
func (impl *NamedType) CheckerType() {}
type NamedType struct {
	Name def.Symbol
	Args [] Type
}
func (impl *AnonymousType) CheckerType() {}
type AnonymousType struct {
	Repr  TypeRepr
}
func (impl *NeverType) CheckerType() {}
type NeverType struct {}
func (impl *AnyType) CheckerType() {}
type AnyType struct {}


type TypeRepr interface { TypeRepr() }

func (impl Unit) TypeRepr() {}
type Unit struct {}
func (impl Tuple) TypeRepr() {}
type Tuple struct {
	Elements  [] Type
}
func (impl Record) TypeRepr() {}
type Record struct {
	// TODO: solve this problem: unordered map can make error position random
	Fields  map[string] Field
}
type Field struct {
	Type   Type
	Index  uint
}
func (impl Func) TypeRepr() {}
type Func struct {
	Input   Type
	Output  Type
}


func TypeFrom(ast_type ast.VariousType, ctx TypeContext) (Type, *TypeError) {
	var t, info, err = TypeNoBoundCheckFrom(ast_type, ctx.TypeValidationContext)
	if err != nil { return nil, err }
	err = CheckTypeBounds(t, info, ctx.TypeBoundsContext)
	if err != nil { return nil, err }
	return t, nil
}

func TypeNoBoundCheckFrom(ast_type ast.VariousType, ctx TypeValidationContext) (Type, (map[Type] ast.Node), *TypeError) {
	var info = make(map[Type] ast.Node)
	var t, err = RawTypeFrom(ast_type, info, ctx.TypeConstructContext)
	if err != nil { return nil, info, err }
	err = ValidateType(t, info, ctx)
	if err != nil { return nil, info, err }
	return t, nil, nil
}

func TypeFromRepr(ast_repr ast.VariousRepr, ctx TypeContext) (Type, *TypeError) {
	var info = make(map[Type] ast.Node)
	var t, err = RawTypeFromRepr(ast_repr, info, ctx.TypeConstructContext)
	if err != nil { return nil, err }
	err = ValidateType(t, info, ctx.TypeValidationContext)
	if err != nil { return nil, err }
	err = CheckTypeBounds(t, info, ctx.TypeBoundsContext)
	if err != nil { return nil, err }
	return t, nil
}


func NormalizeType(t Type, reg TypeRegistry) Type {
	switch T := t.(type) {
	case *AnyType, *NeverType, *ParameterType:
		return t
	case *NamedType:
		var g = reg[T.Name]
		var arity = uint(len(g.Params))
		var full_args = make([] Type, arity)
		for i := uint(0); i < arity; i += 1 {
			if i < uint(len(T.Args)) {
				full_args[i] = T.Args[i]
			} else {
				var default_, exists = g.Defaults[i]
				if !(exists) { panic("something went wrong") }
				full_args[i] = default_
			}
		}
		return &NamedType {
			Name: T.Name,
			Args: full_args,
		}
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Unit:
			return t
		case Tuple:
			var L = len(R.Elements)
			var elements = make([] Type, L)
			for i, el := range R.Elements {
				elements[i] = NormalizeType(el, reg)
			}
			return &AnonymousType { Tuple { elements } }
		case Record:
			var fields = make(map[string] Field)
			for name, field := range R.Fields {
				fields[name] = Field {
					Type:  NormalizeType(field.Type, reg),
					Index: field.Index,
				}
			}
			return &AnonymousType { Record { fields } }
		case Func:
			var input = NormalizeType(R.Input, reg)
			var output = NormalizeType(R.Output, reg)
			return &AnonymousType { Func { Input: input, Output: output } }
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func TypeEqual(type1 Type, type2 Type, reg TypeRegistry) bool {
	return TypeEqualWithoutContext (
		NormalizeType(type1, reg),
		NormalizeType(type2, reg),
	)
}

func TypeEqualWithoutContext(type1 Type, type2 Type) bool {
	switch t1 := type1.(type) {
	case *NeverType:
		switch type2.(type) {
		case *NeverType:
			return true
		default:
			return false
		}
	case *AnyType:
		switch type2.(type) {
		case *AnyType:
			return true
		default:
			return false
		}
	case *ParameterType:
		switch t2 := type2.(type) {
		case *ParameterType:
			return t1.Index == t2.Index
		default:
			return false
		}
	case *NamedType:
		switch t2 := type2.(type) {
		case *NamedType:
			if t1.Name == t2.Name {
				var L1 = len(t1.Args)
				var L2 = len(t2.Args)
				if L1 != L2 {
					return false
				}
				var L = L1
				for i := 0; i < L; i += 1 {
					if !(TypeEqualWithoutContext(t1.Args[i], t2.Args[i])) {
						return false
					}
				}
				return true
			} else {
				return false
			}
		default:
			return false
		}
	case *AnonymousType:
		switch t2 := type2.(type) {
		case *AnonymousType:
			switch r1 := t1.Repr.(type) {
			case Unit:
				switch t2.Repr.(type) {
				case Unit:
					return true
				default:
					return false
				}
			case Tuple:
				switch r2 := t2.Repr.(type) {
				case Tuple:
					var L1 = len(r1.Elements)
					var L2 = len(r2.Elements)
					if L1 == L2 {
						var L = L1
						for i := 0; i < L; i += 1 {
							if !(TypeEqualWithoutContext(r1.Elements[i], r2.Elements[i])) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Record:
				switch r2 := t2.Repr.(type) {
				case Record:
					var L1 = len(r1.Fields)
					var L2 = len(r2.Fields)
					if L1 == L2 {
						for name, f1 := range r1.Fields {
							var f2, exists = r2.Fields[name]
							if !exists || !(TypeEqualWithoutContext(f1.Type, f2.Type)) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Func:
				switch r2 := t2.Repr.(type) {
				case Func:
					if !(TypeEqualWithoutContext(r1.Input, r2.Input)) {
						return false
					}
					if !(TypeEqualWithoutContext(r1.Output, r2.Output)) {
						return false
					}
					return true
				default:
					return false
				}
			default:
				panic("impossible branch")
			}
		default:
			return false
		}
	default:
		panic("impossible branch")
	}
}

