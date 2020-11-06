package container

import (
	"reflect"
	"unicode/utf8"
	. "kumachan/error"
	. "kumachan/runtime/common"
)


type Encoding  int
const (
	UTF8  Encoding  =  iota
)

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
		var result = Equal
		var first_different = true
		for i := 0; i < len(b); i += 1 {
			if i < len(a) {
				if a[i] < b[i] {
					if first_different {
						result = Smaller
						first_different = false
					}
				}
				if a[i] > b[i] {
					if first_different {
						result = Bigger
						first_different = false
					}
				}
			} else {
				if first_different {
					// b starts with a and longer than a
					return Smaller
				} else {
					return result
				}
			}
		}
		return result
	} else {
		return StringCompare(b, a).Reversed()
	}
}

func StringFastCompare(a String, b String) Ordering {
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
		return StringFastCompare(b, a).Reversed()
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

func StringTrim(str String, char Char) String {
	return StringTrimRight(StringTrimLeft(str, char), char)
}

func StringTrimLeft(str String, char Char) String {
	for len(str) > 0 && str[0] == char {
		if len(str) >= 2 {
			str = str[1:]
		} else {
			str = String([] Char {})
		}
	}
	return str
}

func StringTrimRight(str String, char Char) String {
	for len(str) > 0 && str[len(str)-1] == char {
		str = str[:len(str)-1]
	}
	return str
}

func StringTrimPrefix(str String, prefix String) String {
	if len(prefix) <= len(str) {
		for i := 0; i < len(prefix); i += 1 {
			if prefix[i] != str[i] {
				return str
			}
		}
		if len(prefix) < len(str) {
			return str[len(prefix):]
		} else {
			return String([] Char {})
		}
	} else {
		return str
	}
}

func StringTrimSuffix(str String, suffix String) String {
	if len(str) > 0 && len(suffix) > 0 && len(suffix) <= len(str) {
		var j = len(suffix)-1
		for i := len(str)-1; i >= len(str)-len(suffix); i -= 1 {
			if suffix[j] != str[i] {
				return str
			}
			j -= 1
		}
		return str[:len(str)-len(suffix)]
	} else {
		return str
	}
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
