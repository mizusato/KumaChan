package checker

import (
	"kumachan/loader"
	"kumachan/parser/ast"
	"strings"
)


type GenericType struct {
	Params     [] string
	Value      TypeVal
	Node       ast.Node
	CaseIndex  uint
}
type TypeVal interface { TypeVal() }

func (impl Union) TypeVal() {}
type Union struct {
	CaseTypes  [] CaseType
}
type CaseType struct {
	Name    loader.Symbol
	Params  [] uint
}
func (impl Boxed) TypeVal() {}
type Boxed struct {
	InnerType  Type
	AsIs       bool
	Protected  bool
	Opaque     bool
}
func (impl Native) TypeVal() {}
type Native struct {}


type Type interface { CheckerType() }

func (impl ParameterType) CheckerType() {}
type ParameterType struct {
	Index          uint
	BeingInferred  bool
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
func (impl WildcardRhsType) CheckerType() {}
type WildcardRhsType struct {}
const WildcardRhsTypeDesc = "?"


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


func GetCaseInfo(u Union, args []Type, sym loader.Symbol) (uint, []Type, bool) {
	for index, case_type := range u.CaseTypes {
		if case_type.Name == sym {
			var case_args = make([]Type, len(case_type.Params))
			for i, which_arg := range case_type.Params {
				case_args[i] = args[which_arg]
			}
			return uint(index), case_args, true
		}
	}
	return BadIndex, nil, false
}

type TypeDescContext struct {
	ParamNames     [] string
	UseInferred    bool
	InferredNames  [] string
	InferredTypes  map[uint] Type
}

func DescribeTypeWithParams(type_ Type, params []string) string {
	return DescribeType(type_, TypeDescContext {
		ParamNames:    params,
		UseInferred:   false,
	})
}

func DescribeType(type_ Type, ctx TypeDescContext) string {
	switch t := type_.(type) {
	case WildcardRhsType:
		return WildcardRhsTypeDesc
	case ParameterType:
		if ctx.UseInferred {
			var inferred_t, exists = ctx.InferredTypes[t.Index]
			if exists {
				return DescribeType(inferred_t, ctx)
			} else {
				return ctx.InferredNames[t.Index]
			}
		} else {
			return ctx.ParamNames[t.Index]
		}
	case NamedType:
		var buf strings.Builder
		if loader.IsPreloadCoreSymbol(t.Name) {
			buf.WriteString(t.Name.SymbolName)
		} else {
			buf.WriteString(t.Name.String())
		}
		if len(t.Args) > 0 {
			buf.WriteRune('[')
			for i, arg := range t.Args {
				buf.WriteString(DescribeType(arg, ctx))
				if i != len(t.Args)-1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteRune(']')
		}
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
			var i = 0
			for name, field := range r.Fields {
				buf.WriteString(name)
				buf.WriteString(": ")
				buf.WriteString(DescribeType(field.Type, ctx))
				if i != len(r.Fields)-1 {
					buf.WriteString(", ")
				}
				i += 1
			}
			buf.WriteString(" }")
			return buf.String()
		case Func:
			var buf strings.Builder
			buf.WriteString("(λ ")
			buf.WriteString(DescribeType(r.Input, ctx))
			buf.WriteString(" ")
			buf.WriteString(DescribeType(r.Output, ctx))
			buf.WriteString(")")
			return buf.String()
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func AreTypesEqualInSameCtx(type1 Type, type2 Type) bool {
	switch t1 := type1.(type) {
	case WildcardRhsType:
		return false
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
