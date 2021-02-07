package checker

import (
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
)


type GenericType struct {
	Doc         string
	Tags        TypeTags
	Params      [] TypeParam
	Bounds      TypeBounds
	Defaults    map[uint] Type
	Definition  TypeDef
	Node        ast.Node
	CaseInfo    CaseInfo
}
type TypeParam struct {
	Name      string
	Variance  TypeVariance
}
func TypeParamsNames(params ([] TypeParam)) ([] string) {
	var draft = make([] string, len(params))
	for i, _ := range draft {
		draft[i] = params[i].Name
	}
	return draft
}


type TypeDef interface { CheckerTypeDef() }

func (impl *Enum) CheckerTypeDef() {}
type Enum struct {
	CaseTypes  [] CaseType
}
type CaseType struct {
	Name    loader.Symbol
	Params  [] uint
}
func (impl *Boxed) CheckerTypeDef() {}
type Boxed struct {
	InnerType  Type
	Implicit   bool
	// following properties are exclusive
	Weak       bool  // TODO: make it a standalone option
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
	Name  loader.Symbol
	Args  [] Type
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
func (impl Bundle) TypeRepr() {}
type Bundle struct {
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


func AreTypesEqualInSameCtx(type1 Type, type2 Type) bool {
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
				if L1 != L2 { panic("type registration went wrong") }
				var L = L1
				for i := 0; i < L; i += 1 {
					if !(AreTypesEqualInSameCtx(t1.Args[i], t2.Args[i])) {
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
							if !(AreTypesEqualInSameCtx(r1.Elements[i], r2.Elements[i])) {
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
			case Bundle:
				switch r2 := t2.Repr.(type) {
				case Bundle:
					var L1 = len(r1.Fields)
					var L2 = len(r2.Fields)
					if L1 == L2 {
						for name, f1 := range r1.Fields {
							var f2, exists = r2.Fields[name]
							if !exists || !(AreTypesEqualInSameCtx(f1.Type, f2.Type)) {
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
					if !(AreTypesEqualInSameCtx(r1.Input, r2.Input)) {
						return false
					}
					if !(AreTypesEqualInSameCtx(r1.Output, r2.Output)) {
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

