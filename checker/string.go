package checker

import (
	"kumachan/transformer/ast"
	"strings"
)


func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  [] rune
}

func (impl StringFormatter) ExprVal() {}
type StringFormatter struct {
	Segments  [] []rune
	Arity     uint
}


func CheckString(s ast.StringLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
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

func CheckText(text ast.Text, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(text.Node)
	var template = text.Template
	var segments = make([] []rune, 0)
	var arity uint = 0
	var buf strings.Builder
	for _, char := range template {
		if char == TextPlaceholder {
			var seg = buf.String()
			buf.Reset()
			segments = append(segments, []rune(seg))
			arity += 1
		} else {
			buf.WriteRune(char)
		}
	}
	var last = buf.String()
	if last != "" {
		segments = append(segments, []rune(last))
	}
	var elements = make([]Type, arity)
	for i := uint(0); i < arity; i += 1 {
		elements[i] = NamedType { Name: __String, Args: make([]Type, 0) }
	}
	var input Type
	if len(elements) == 0 {
		input = AnonymousType { Unit {} }
	} else if len(elements) == 1 {
		input = elements[0]
	} else {
		input = AnonymousType { Tuple { elements } }
	}
	var t Type = AnonymousType { Func {
		Input:  input,
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
