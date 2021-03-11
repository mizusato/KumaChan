package checker

import (
	"unsafe"
	"strings"
	"math/big"
	. "kumachan/util/error"
	"kumachan/stdlib"
	"kumachan/lang/parser/ast"
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
	var value = make([] uint32, len(s.First.Value))
	copy(value, CharSliceFromRuneSlice(s.First.Value))
	for _, part := range s.Parts {
		switch p := part.Part.(type) {
		case ast.StringText:
			value = append(value, CharSliceFromRuneSlice(p.Value)...)
		case ast.CharLiteral:
			var char, err = GetChar(p, ctx)
			if err != nil { return SemiExpr{}, err }
			value = append(value, char)
		}
	}
	var info = ctx.GetExprInfo(s.Node)
	return LiftTyped(Expr {
		Type:  &NamedType {
			Name: __String,
			Args: make([]Type, 0),
		},
		Info:  info,
		Value: StringLiteral { value },
	}), nil
}

func CheckFormatter(formatter ast.Formatter, ctx ExprContext) (SemiExpr, *ExprError) {
	var template = make([] uint32, len(formatter.First.Template))
	copy(template, CharSliceFromRuneSlice(formatter.First.Template))
	var is_raw_char = make(map[uint] bool)
	for _, part := range formatter.Parts {
		switch p := part.Part.(type) {
		case ast.FormatterText:
			template = append(template, CharSliceFromRuneSlice(p.Template)...)
		case ast.CharLiteral:
			var char, err = GetChar(p, ctx)
			if err != nil { return SemiExpr{}, err }
			is_raw_char[uint(len(template))] = true
			template = append(template, char)
		}
	}
	var info = ctx.GetExprInfo(formatter.Node)
	var segments = make([] [] uint32, 0)
	var arity uint = 0
	var buf strings.Builder
	for i, char := range template {
		if char == TextPlaceholder && !(is_raw_char[uint(i)]) {
			var seg = buf.String()
			buf.Reset()
			segments = append(segments, CharSliceFromRuneSlice([]rune(seg)))
			arity += 1
		} else {
			buf.WriteRune(rune(char))
		}
	}
	var last = buf.String()
	if last != "" {
		segments = append(segments, CharSliceFromRuneSlice([]rune(last)))
	}
	var elements = make([] Type, arity)
	for i := uint(0); i < arity; i += 1 {
		elements[i] = __T_String
	}
	var input Type
	if len(elements) == 0 {
		input = &AnonymousType { Unit {} }
	} else if len(elements) == 1 {
		input = elements[0]
	} else {
		input = &AnonymousType { Tuple { elements } }
	}
	var t Type = &AnonymousType { Func {
		Input:  input,
		Output: __T_String,
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
	var char_value, err = GetChar(char, ctx)
	if err != nil { return SemiExpr{}, err }
	return LiftTyped(Expr {
		Type:  __T_Char,
		Value: SmallIntLiteral {
			Value: char_value,
		},
		Info:  ctx.GetExprInfo(char.Node),
	}), nil
}

func GetChar(char ast.CharLiteral, ctx ExprContext) (uint32, *ExprError) {
	var raw = char.Value
	if len(raw) < 2 { panic("something went wrong") }
	var use_rune = func(r rune) (uint32, *ExprError) {
		return uint32(r), nil
	}
	var invalid = func() (uint32, *ExprError) {
		return ^(uint32(0)), &ExprError {
			Point:    ErrorPointFrom(char.Node),
			Concrete: E_InvalidCharacter { string(raw) },
		}
	}
	var c0 = raw[0]
	if c0 == '`' {
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
			return val.(SmallIntLiteral).Value.(uint32), nil
		default:
			return invalid()
		}
	} else {
		panic("something went wrong")
	}
}