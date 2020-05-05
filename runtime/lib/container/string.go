package container

import (
	"reflect"
	"unsafe"
	"strings"
	"unicode/utf8"
	. "kumachan/error"
	. "kumachan/runtime/common"
)


type Char = uint32
type String = ([] Char)

type Encoding  int
const (
	UTF8  Encoding  =  iota
)

func StringFromGoString(bytes string) String {
	var str = make(String, 0, len(bytes) / 4)
	for len(bytes) > 0 {
		var char, size = utf8.DecodeRuneInString(bytes)
		str = append(str, Char(char))
		bytes = bytes[size:]
	}
	return str
}

func StringFromRuneSlice(runes ([] rune)) String {
	return *(*([] Char))(unsafe.Pointer(&runes))
}

func StringForceDecode(bytes Bytes, e Encoding) String {
	switch e {
	case UTF8:
		var str = make(String, 0, len(bytes) / 4)
		for len(bytes) > 0 {
			var char, size = utf8.DecodeRune(bytes)
			str = append(str, Char(char))
			bytes = bytes[size:]
		}
		return str
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringDecode(bytes Bytes, e Encoding) (String, bool) {
	switch e {
	case UTF8:
		var str = make(String, 0, len(bytes) / 4)
		for len(bytes) > 0 {
			var char, size = utf8.DecodeRune(bytes)
			if char == utf8.RuneError && size == 1 {
				// Note: An error should be thrown when input is invalid
				//       to ensure this function to be invertible.
				return nil, false
			}
			str = append(str, Char(char))
			bytes = bytes[size:]
		}
		return str, true
	default:
		panic("unknown or unimplemented encoding")
	}
}

func GoStringFromString(str String) string {
	var buf strings.Builder
	for _, char := range str {
		buf.WriteRune(rune(char))
	}
	return buf.String()
}

func StringEncode(str String, e Encoding) Bytes {
	switch e {
	case UTF8:
		var buf = make(Bytes, 0, len(str))
		var chunk ([4] byte)
		for _, r := range str {
			var size = utf8.EncodeRune(chunk[:], rune(r))
			for i := 0; i < size; i += 1 {
				buf = append(buf, chunk[i])
			}
		}
		return buf
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringCompare(a String, b String) Ordering {
	if len(a) <= len(b) {
		for i := 0; i < len(b); i += 1 {
			if i < len(a) {
				if a[i] < b[i] {
					return Smaller
				} else if a[i] > b[i] {
					return Bigger
				}
			} else {
				return Smaller
			}
		}
		return Equal
	} else {
		return StringCompare(b, a).Reversed()
	}
}

func StringConcat(arr Array) String {
	var buf = make(String, 0)
	for i := uint(0); i < arr.Length; i += 1 {
		var item = (arr.GetItem(i)).(String)
		buf = append(buf, item...)
	}
	return buf
}

func StringFind(str String, sub String) (uint, bool) {
	var L = uint(len(str))
	var M = uint(len(sub))
	outer: for i := uint(0); i < L; i += 1 {
		var ok = true
		for j := uint(0); j < M; j += 1 {
			var k = i + j
			if !(k < L) {
				ok = false
				break outer
			}
			if str[k] != sub[j] {
				ok = false
				break
			}
		}
		if ok {
			return i, true
		}
	}
	return (^uint(0)), false
}

type StringSplitIterator struct {
	Operand    String
	Separator  String
	LastIndex  uint
}
func (it *StringSplitIterator) GetItemType() reflect.Type {
	return reflect.TypeOf(StringFromGoString(""))
}
func (it *StringSplitIterator) Next() (Value, Seq, bool) {
	if it == nil { return nil, nil, false }
	var L = uint(len(it.Operand))
	var M = uint(len(it.Separator))
	outer: for i := it.LastIndex; i < L; i += 1 {
		var ok = true
		for j := uint(0); j < M; j += 1 {
			var k = i + j
			if !(k < L) {
				ok = false
				break outer
			}
			if it.Operand[k] != it.Separator[j] {
				ok = false
				break
			}
		}
		if ok {
			var item = append(String{}, it.Operand[it.LastIndex:i]...)
			var rest = &StringSplitIterator {
				Operand:   it.Operand,
				Separator: it.Separator,
				LastIndex: (i + M),
			}
			return item, rest, ok
		}
	}
	var item = append(String{}, it.Operand[it.LastIndex:]...)
	return item, nil, true
}
func (_ *StringSplitIterator) Inspect(_ func(Value)ErrorMessage) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "[seq string-split-iterator]")
	return msg
}
func StringSplit(str String, sep String) Seq {
	return &StringSplitIterator {
		Operand:   str,
		Separator: sep,
		LastIndex: 0,
	}
}

func StringJoin(seq Seq, sep String) String {
	var buf = make(String, 0)
	var index = uint(0)
	for v,rest,ok := seq.Next(); ok; v,rest,ok = rest.Next() {
		if index > 0 {
			buf = append(buf, sep...)
		}
		buf = append(buf, v.(String)...)
		index += 1
	}
	return buf
}
