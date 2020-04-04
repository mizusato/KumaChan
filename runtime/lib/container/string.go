package container

import (
	"kumachan/runtime/common"
	"unicode/utf8"
)

type String = []rune

type Encoding  int
const (
	UTF8  Encoding  =  iota
)

func StringFrom(bytes Bytes, e Encoding) (String, bool) {
	switch e {
	case UTF8:
		var str = make(String, 0, len(bytes) / 4)
		for len(bytes) > 0 {
			var char, size = utf8.DecodeRune(bytes)
			if char == utf8.RuneError { return nil, false }
			str = append(str, char)
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
		var chunk [4]byte
		for _, r := range str {
			var size = utf8.EncodeRune(chunk[:], r)
			for i := 0; i < size; i += 1 {
				buf = append(buf, chunk[i])
			}
		}
		return buf
	default:
		panic("unknown or unimplemented encoding")
	}
}

func StringCompare(a String, b String) common.Ordering {
	if len(a) <= len(b) {
		for i := 0; i < len(b); i += 1 {
			if i < len(a) {
				if a[i] < b[i] {
					return common.Smaller
				} else if a[i] > b[i] {
					return common.Bigger
				}
			} else {
				return common.Smaller
			}
		}
		return common.Equal
	} else {
		return StringCompare(b, a).Reversed()
	}
}
