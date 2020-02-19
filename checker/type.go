package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	"strings"
)


type GenericType struct {
	Arity       uint
	IsOpaque    bool
	Value       TypeVal
	Node        node.Node
	UnionIndex  uint
}
type TypeVal interface { TypeVal() }

func (impl UnionTypeVal) TypeVal() {}
type UnionTypeVal struct {
	SubTypes  [] loader.Symbol
}
func (impl SingleTypeVal) TypeVal() {}
type SingleTypeVal struct {
	InnerType  Type
}
func (impl NativeTypeVal) TypeVal() {}
type NativeTypeVal struct {}


type Type interface { CheckerType() }

func (impl ParameterType) CheckerType() {}
type ParameterType struct {
	Index  uint
}
func (impl NamedType) CheckerType() {}
type NamedType struct {
	Name  loader.Symbol
	Args  [] Type
}
func (impl AnonymousType) CheckerType() {}
type AnonymousType struct {
	Repr  TypeRepr
}


type TypeRepr interface { TypeRepr() }

func (impl Unit) TypeRepr() {}
type Unit struct {}
func (impl Tuple) TypeRepr() {}
type Tuple struct {
	Elements  [] Type
}
func (impl Bundle) TypeRepr() {}
type Bundle struct {
	Fields  map[string] Type
	Index   map[string] uint
}
func (impl Func) TypeRepr() {}
type Func struct {
	Input   Type
	Output  Type
}


func DescribeType(type_ Type, ctx TypeContext) string {
	switch t := type_.(type) {
	case ParameterType:
		return ctx.Params[t.Index]
	case NamedType:
		var buf strings.Builder
		buf.WriteString(t.Name.String())
		buf.WriteRune('[')
		for i, arg := range t.Args {
			buf.WriteString(DescribeType(arg, ctx))
			if i != len(t.Args)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteRune(']')
		return buf.String()
	case AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return "()"
		case Tuple:
			var buf strings.Builder
			buf.WriteRune('(')
			for i, el := range r.Elements {
				buf.WriteString(DescribeType(el, ctx))
				if i != len(r.Elements)-1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteRune(')')
			return buf.String()
		case Bundle:
			var buf strings.Builder
			buf.WriteString("{ ")
			for name, field := range r.Fields {
				buf.WriteString(name)
				buf.WriteString(": ")
				buf.WriteString(DescribeType(field, ctx))
			}
			buf.WriteString(" }")
			return buf.String()
		case Func:
			var buf strings.Builder
			buf.WriteRune('λ')
			buf.WriteString(DescribeType(r.Input, ctx))
			buf.WriteString(" -> ")
			buf.WriteString(DescribeType(r.Output, ctx))
			return buf.String()
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func IsLocalType (type_ Type, mod string) bool {
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
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func AreTypesOverloadUnsafe (type1 Type, type2 Type) bool {
	// Are type1 and type2 equal in the context of function overloading
	switch t1 := type1.(type) {
	case ParameterType:
		return true  // rough comparison
	case NamedType:
		switch t2 := type2.(type) {
		case NamedType:
			if t1.Name == t2.Name {
				var L1 = len(t1.Args)
				var L2 = len(t2.Args)
				if L1 != L2 { panic("type registration went wrong") }
				var L = L1
				for i := 0; i < L; i += 1 {
					if !(AreTypesOverloadUnsafe(t1.Args[i], t2.Args[i])) {
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
							if !(AreTypesOverloadUnsafe(r1.Elements[i], r2.Elements[i])) {
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
							if !exists || !(AreTypesOverloadUnsafe(f1, f2)) {
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
					if !(AreTypesOverloadUnsafe(r1.Input, r2.Input)) {
						return false
					}
					if !(AreTypesOverloadUnsafe(r1.Output, r2.Output)) {
						return false
					}
					return true
				default:
					return true
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

func AreTypesEqualInSameCtx (type1 Type, type2 Type) bool {
	switch t1 := type1.(type) {
	case ParameterType:
		switch t2 := type2.(type) {
		case ParameterType:
			return t1.Index == t2.Index
		default:
			return false
		}
	case NamedType:
		switch t2 := type2.(type) {
		case NamedType:
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
	case AnonymousType:
		switch t2 := type2.(type) {
		case AnonymousType:
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
							if !exists || !(AreTypesEqualInSameCtx(f1, f2)) {
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
					return true
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

