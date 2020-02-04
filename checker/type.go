package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	"kumachan/native/types"
)

type GenericType struct {
	Arity     uint
	IsOpaque  bool
	Value     TypeVal
	Node      node.Node
	Order     uint
}

type TypeVal interface { TypeVal() }

func (impl UnionTypeVal) TypeVal() {}
type UnionTypeVal struct {
	SubTypes  [] loader.Symbol
}

func (impl SingleTypeVal) TypeVal() {}
type SingleTypeVal struct {
	Expr  TypeExpr
}

type TypeExpr interface { TypeExpr() }

func (impl ParameterType) TypeExpr() {}
type ParameterType struct {
	Index  uint
}

func (impl NamedType) TypeExpr() {}
type NamedType struct {
	Name  loader.Symbol
	Args  [] TypeExpr
}

func (impl AnonymousType) TypeExpr() {}
type AnonymousType struct {
	Repr  TypeRepr
}

type TypeRepr interface { TypeRepr() }

func (impl Unit) TypeRepr() {}
type Unit struct {}

func (impl Tuple) TypeRepr() {}
type Tuple struct {
	Elements  [] TypeExpr
}

func (impl Bundle) TypeRepr() {}
type Bundle struct {
	Fields  map[string] TypeExpr
}

func (impl Func) TypeRepr() {}
type Func struct {
	Input   TypeExpr
	Output  TypeExpr
}

func (impl NativeType) TypeRepr() {}
type NativeType struct {
	Id  types.NativeTypeId
}


func IsLocalType (type_ TypeExpr, mod string) bool {
	switch t := type_.(type) {
	case ParameterType:
		return false
	case NamedType:
		if t.Name.ModuleName == mod {
			return true
		} else {
			for _, arg := range t.Args {
				if IsLocalType(arg, mod) {
					return true
				}
			}
			return false
		}
	case AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return false
		case Tuple:
			for _, el := range r.Elements {
				if IsLocalType(el, mod) {
					return true
				}
			}
			return false
		case Bundle:
			for _, f := range r.Fields {
				if IsLocalType(f, mod) {
					return true
				}
			}
			return false
		case Func:
			if IsLocalType(r.Input, mod) {
				return true
			}
			if IsLocalType(r.Output, mod) {
				return true
			}
			return false
		case NativeType:
			return false
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func AreTypesRoughlyEqual (type1 TypeExpr, type2 TypeExpr) bool {
	switch t1 := type1.(type) {
	case ParameterType:
		return false
	case NamedType:
		switch t2 := type2.(type) {
		case NamedType:
			if t1.Name == t2.Name {
				var L1 = len(t1.Args)
				var L2 = len(t2.Args)
				if L1 != L2 { panic("type registration went wrong") }
				var L = L1
				for i := 0; i < L; i += 1 {
					if !(AreTypesRoughlyEqual(t1.Args[i], t2.Args[i])) {
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
	case AnonymousType:
		switch t2 := type2.(type) {
		case AnonymousType:
			switch r1 := t1.Repr.(type) {
			case Unit:
				switch _ := t2.Repr.(type) {
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
							if !(AreTypesRoughlyEqual(r1.Elements[i], r2.Elements[i])) {
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
							if !exists || !(AreTypesRoughlyEqual(f1, f2)) {
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
					if !(AreTypesRoughlyEqual(r1.Input, r2.Input)) {
						return false
					}
					if !(AreTypesRoughlyEqual(r1.Output, r2.Output)) {
						return false
					}
					return true
				default:
					return false
				}
			case NativeType:
				switch r2 := t2.Repr.(type) {
				case NativeType:
					return r1.Id == r2.Id
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
