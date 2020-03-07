package checker

import (
	"kumachan/transformer/node"
	"strings"
)


func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  [] rune
}

func (impl StringFormatter) ExprVal() {}
type StringFormatter struct {
	Segments  [] string
	Arity     uint
}
func GetFormatterFunction(sf *StringFormatter) func([]string)string {
	return func(args []string) string {
		var buf strings.Builder
		for i, seg := range sf.Segments {
			buf.WriteString(seg)
			if uint(i) < sf.Arity {
				buf.WriteString(args[i])
			}
		}
		return buf.String()
	}
}


func CheckString(s node.StringLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(s.Node)
	return LiftTyped(Expr {
		Type:  NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { s.Value },
	}), nil
}

func CheckText(text node.Text, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(text.Node)
	var template = text.Template
	var segments = make([]string, 0)
	var arity uint = 0
	var buf strings.Builder
	for _, char := range template {
		if char == TextPlaceholder {
			var seg = buf.String()
			buf.Reset()
			segments = append(segments, seg)
			arity += 1
		} else {
			buf.WriteRune(char)
		}
	}
	var last = buf.String()
	if last != "" {
		segments = append(segments, last)
	}
	var elements = make([]Type, arity)
	for i := uint(0); i < arity; i += 1 {
		elements[i] = NamedType { Name: __String, Args: make([]Type, 0) }
	}
	var t Type = AnonymousType { Func {
		Input:  AnonymousType { Tuple { elements } },
		Output: NamedType { Name: __String, Args: make([]Type, 0) },
	} }
	return LiftTyped(Expr {
		Type:  t,
		Value: StringFormatter {
			Segments: segments,
			Arity:    arity,
		},
		Info:  info,
	}), nil
}
