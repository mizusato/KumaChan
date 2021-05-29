package checker

import (
	"strings"
	"kumachan/interpreter/base"
	"kumachan/interpreter/compiler/loader"
)


type TypeDescContext struct {
	ParamNames     [] string
	InferredNames  [] string
	InferredTypes  map[uint] Type
	CurrentModule  string
}

func getModuleNameCommonPrefixLength(a string, b string) uint {
	var L = uint(0)
	var A = ([] rune)(a)
	var B = ([] rune)(b)
	var S uint
	if len(A) > len(B) { S = uint(len(B)) } else { S = uint(len(A)) }
	for i := uint(0); i < S; i += 1 {
		if A[i] == B[i] {
			L += 1
		} else {
			break
		}
	}
	for L > 0 && A[L-1] != '.' && A[L-1] != ':' {
		L -= 1
	}
	return uint(len(string(A[:L])))
}

func GetClearModuleName(mod string, current_mod string) string {
	if mod == current_mod {
		return ""
	}
	var L = getModuleNameCommonPrefixLength(mod, current_mod)
	var clear string
	if L < uint(len(mod)) {
		clear = mod[L:]
	} else {
		clear = ""
	}
	if strings.HasSuffix(clear, ":" + loader.DefaultVersion) {
		clear = strings.TrimSuffix(clear, ":" + loader.DefaultVersion)
	}
	return clear
}

func DescribeTypeWithParams(type_ Type, params ([] string), mod string) string {
	return DescribeType(type_, TypeDescContext {
		ParamNames:    params,
		CurrentModule: mod,
	})
}

func DescribeType(type_ Type, ctx TypeDescContext) string {
	switch t := type_.(type) {
	case *NeverType:
		return NeverTypeName
	case *AnyType:
		return AnyTypeName
	case *ParameterType:
		if t.BeingInferred {
			var inferred_t, exists = ctx.InferredTypes[t.Index]
			if exists {
				return DescribeType(inferred_t, ctx)
			} else {
				return ctx.InferredNames[t.Index]
			}
		} else {
			return ctx.ParamNames[t.Index]
		}
	case *NamedType:
		var buf strings.Builder
		if loader.IsCoreSymbol(t.Name) {
			buf.WriteString(t.Name.SymbolName)
		} else {
			var mod = t.Name.ModuleName
			var clear = GetClearModuleName(mod, ctx.CurrentModule)
			var clear_sym = base.MakeSymbol(clear, t.Name.SymbolName)
			buf.WriteString(clear_sym.String())
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
	case *AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return "unit"
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
		case Record:
			var field_types = make([] Type, len(r.Fields))
			var field_names = make([] string, len(r.Fields))
			for name, field := range r.Fields {
				field_types[field.Index] = field.Type
				field_names[field.Index] = name
			}
			var buf strings.Builder
			buf.WriteString("{ ")
			for i := 0; i < len(r.Fields); i += 1 {
				buf.WriteString(field_names[i])
				buf.WriteString(": ")
				buf.WriteString(DescribeType(field_types[i], ctx))
				if i != len(r.Fields)-1 {
					buf.WriteString(", ")
				}
			}
			buf.WriteString(" }")
			return buf.String()
		case Func:
			var input_desc = DescribeType(r.Input, ctx)
			var need_wrap =
				!(strings.HasPrefix(input_desc, "(") ||
				strings.HasPrefix(input_desc, "{"))
			var buf strings.Builder
			buf.WriteString("λ")
			if need_wrap { buf.WriteString("(") }
			buf.WriteString(input_desc)
			if need_wrap { buf.WriteString(")") }
			buf.WriteString(" => ")
			buf.WriteString(DescribeType(r.Output, ctx))
			return buf.String()
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}
