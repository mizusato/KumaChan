package checker

import (
	"strings"
	"kumachan/loader"
)


type TypeDescContext struct {
	ParamNames     [] string
	InferredNames  [] string
	InferredTypes  map[uint] Type
	// TODO: current module name --- omit the common part of the module name
}

func DescribeTypeWithParams(type_ Type, params ([] string)) string {
	return DescribeType(type_, TypeDescContext {
		ParamNames: params,
	})
}

func DescribeType(type_ Type, ctx TypeDescContext) string {
	switch t := type_.(type) {
	case *NeverType:
		return NeverTypeName
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
	case *AnonymousType:
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
