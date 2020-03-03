package checker

import (
	"kumachan/transformer/node"
	"strings"
)


func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  [] rune
}


func CheckString(s node.StringLiteral, ctx ExprContext) (Expr, *ExprError) {
	var info = ctx.GetExprInfo(s.Node)
	return Expr {
		Type:  NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { s.Value },
	}, nil
}

func CheckText(text node.Text, ctx ExprContext) (Expr, *ExprContext) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(text.Node) }
	var template = text.Template
	var segments = make([]string, 0)
	var arity = 0
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
	var format = func(args []string) string {
		var buf strings.Builder
		for i, seg := range segments {
			buf.WriteString(seg)
			if i < arity {
				buf.WriteString(args[i])
			}
		}
		return buf.String()
	}
	var elements = make([]Type, arity)
	for i := 0; i < arity; i += 1 {
		elements[i] = NamedType { Name: __String, Args: make([]Type, 0) }
	}
	var t Type = AnonymousType { Func {
		Input:  AnonymousType { Tuple { elements } },
		Output: NamedType { Name: __String, Args: make([]Type, 0) },
	} }
	return Expr {
		Type:  t,
		Value: NativeFunction { format },
		Info:  info,
	}, nil
}
