package container

import (
	"reflect"
	"strings"
	"unicode/utf8"
	. "kumachan/interpreter/base"
)


type Encoding  int
const (
	UTF8  Encoding  =  iota
)

func StringForceDecode(bytes ([] byte), e Encoding) string {
	switch e {
	case UTF8:
		var buf strings.Builder
		for len(bytes) > 0 {
			var char, size = utf8.DecodeRune(bytes)
			buf.WriteRune(char)
			bytes = bytes[size:]
		}
		return buf.String()
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringDecode(bytes ([] byte), e Encoding) (string, bool) {
	switch e {
	case UTF8:
		var buf strings.Builder
		for len(bytes) > 0 {
			var char, size = utf8.DecodeRune(bytes)
			if char == utf8.RuneError && size == 1 {
				// Note: An error should be thrown when input is invalid
				//       to ensure this function to be invertible.
				return "", false
			}
			buf.WriteRune(char)
			bytes = bytes[size:]
		}
		return buf.String(), true
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringEncode(str string, e Encoding) ([] byte) {
	switch e {
	case UTF8:
		return ([] byte)(str)
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringCompare(a string, b string) Ordering {
	if a < b {
		return Smaller
	} else if a > b {
		return Bigger
	} else {
		return Equal
	}
}

func StringConcat(l List) string {
	var buf strings.Builder
	l.ForEach(func(i uint, item Value) {
		buf.WriteString(item.(string))
	})
	return buf.String()
}

func StringHasSubstring(operand string, sub string) bool {
	return (strings.Index(operand, sub) != -1)
}

type StringIterator struct {
	Operand  string
}
func (it *StringIterator) GetItemType() reflect.Type {
	return reflect.TypeOf(rune(0))
}
func (it *StringIterator) Next() (Value, Seq, bool) {
	if it == nil || it.Operand == "" {
		return nil, nil, false
	}
	var op = it.Operand
	for _, char := range op {
		var rest = op[utf8.RuneLen(char):]
		return char, &StringIterator { rest }, true
	}
	panic("impossible branch")
}

type StringSplitIterator struct {
	Operand    string
	Separator  string
}
func (it *StringSplitIterator) GetItemType() reflect.Type {
	return reflect.TypeOf("")
}
func (it *StringSplitIterator) Next() (Value, Seq, bool) {
	if it == nil || it.Operand == "" {
		return nil, nil, false
	}
	var op = it.Operand
	var sep = it.Separator
	for i, _ := range op {
		if strings.HasPrefix(op[i:], sep) {
			var item = op[:i]
			var next_op = op[i+len(sep):]
			var rest = &StringSplitIterator {
				Operand:   next_op,
				Separator: sep,
			}
			return item, rest, true
		}
	}
	return op, EmptySeq { ItemType: it.GetItemType() }, true
}
func StringSplit(str string, sep string) Seq {
	return &StringSplitIterator {
		Operand:   str,
		Separator: sep,
	}
}

func StringJoin(seq Seq, sep string) string {
	var buf strings.Builder
	var index = uint(0)
	for v,rest,ok := seq.Next(); ok; v,rest,ok = rest.Next() {
		if index > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(v.(string))
		index += 1
	}
	return buf.String()
}

