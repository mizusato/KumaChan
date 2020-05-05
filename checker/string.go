package checker

import (
	"unsafe"
	"strings"
	"math/big"
	. "kumachan/error"
	"kumachan/stdlib"
	"kumachan/transformer/ast"
)


func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  [] uint32
}

func (impl StringFormatter) ExprVal() {}
type StringFormatter struct {
	Segments  [] [] uint32
	Arity     uint
}

func CharSliceFromRuneSlice(runes ([] rune)) ([] uint32) {
	return *(*([] uint32))(unsafe.Pointer(&runes))
}


func CheckString(s ast.StringLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(s.Node)
	return LiftTyped(Expr {
		Type:  NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { CharSliceFromRuneSlice(s.Value) },
	}), nil
}

func CheckText(text ast.Text, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(text.Node)
	var template = text.Template
	var segments = make([] [] uint32, 0)
	var arity uint = 0
	var buf strings.Builder
	for _, char := range template {
		if char == TextPlaceholder {
			var seg = buf.String()
			buf.Reset()
			segments = append(segments, CharSliceFromRuneSlice([]rune(seg)))
			arity += 1
		} else {
			buf.WriteRune(char)
		}
	}
	var last = buf.String()
	if last != "" {
		segments = append(segments, CharSliceFromRuneSlice([]rune(last)))
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

func CheckChar(char ast.CharLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var raw = char.Value
	if len(raw) < 2 { panic("something went wrong") }
	var use_rune = func(r rune) (SemiExpr, *ExprError) {
		return LiftTyped(Expr {
			Type:  __T_Char,
			Value: SmallIntLiteral {
				Value: uint32(r),
			},
			Info:  ctx.GetExprInfo(char.Node),
		}), nil
	}
	var invalid = func() (SemiExpr, *ExprError) {
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(char.Node),
			Concrete: E_InvalidCharacter { string(raw) },
		}
	}
	var c0 = raw[0]
	if c0 == '^' {
		return use_rune(raw[1])
	} else if c0 == '\\' {
		var c1 = raw[1]
		switch c1 {
		case 'n':
			return use_rune('\n')
		case 'r':
			return use_rune('\r')
		case 't':
			return use_rune('\t')
		case 'e':
			return use_rune('\033')
		case 'b':
			return use_rune('\b')
		case 'a':
			return use_rune('\a')
		case 'f':
			return use_rune('\f')
		case 'u':
			var code_point_raw = string(raw[2:])
			var n, ok1 = big.NewInt(0).SetString(code_point_raw, 16)
			if !ok1 { return invalid() }
			var val, ok2 = AdaptInteger(stdlib.Char, n)
			if !ok2 { return invalid() }
			return LiftTyped(Expr {
				Type:  __T_Char,
				Value: val,
				Info:  ctx.GetExprInfo(char.Node),
			}), nil
		default:
			return invalid()
		}
	} else {
		panic("something went wrong")
	}
}
