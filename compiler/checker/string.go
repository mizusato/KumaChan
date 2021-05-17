package checker

import (
	"strings"
	"math/big"
	. "kumachan/misc/util/error"
	"kumachan/stdlib"
	"kumachan/lang/parser/ast"
)


func (impl StringLiteral) ExprVal() {}
type StringLiteral struct {
	Value  string
}

func (impl StringFormatter) ExprVal() {}
type StringFormatter struct {
	Segments  [] string
	Arity     uint
}


func CheckString(s ast.StringLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var buf = make([] rune, len(s.First.Value))
	copy(buf, s.First.Value)
	for _, part := range s.Parts {
		switch p := part.Part.(type) {
		case ast.StringText:
			buf = append(buf, p.Value...)
		case ast.CharLiteral:
			var char, err = GetChar(p, ctx)
			if err != nil { return SemiExpr{}, err }
			buf = append(buf, char)
		}
	}
	var value = string(buf)
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
	var template = make([] rune, len(formatter.First.Template))
	copy(template, formatter.First.Template)
	var is_raw_char = make(map[uint] bool)
	for _, part := range formatter.Parts {
		switch p := part.Part.(type) {
		case ast.FormatterText:
			template = append(template, p.Template...)
		case ast.CharLiteral:
			var char, err = GetChar(p, ctx)
			if err != nil { return SemiExpr{}, err }
			is_raw_char[uint(len(template))] = true
			template = append(template, char)
		}
	}
	var info = ctx.GetExprInfo(formatter.Node)
	var segments = make([] string, 0)
	var arity uint = 0
	var buf strings.Builder
	for i, char := range template {
		if char == TextPlaceholder && !(is_raw_char[uint(i)]) {
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

func GetChar(char ast.CharLiteral, ctx ExprContext) (rune, *ExprError) {
	var raw = char.Value
	if len(raw) < 2 { panic("something went wrong") }
	var use_rune = func(r rune) (rune, *ExprError) {
		return r, nil
	}
	var invalid = func() (rune, *ExprError) {
		return -1, &ExprError {
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
			// note: due to `rune` is an alias of `int32`,
			//       we should convert `uint32` to `int32` here
			// TODO: validate codepoint
			return rune(val.(SmallIntLiteral).Value.(uint32)), nil
		default:
			return invalid()
		}
	} else {
		panic("something went wrong")
	}
}