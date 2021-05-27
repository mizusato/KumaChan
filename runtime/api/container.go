package api

import (
	"fmt"
	"strconv"
	"strings"
	"math/big"
	"unicode/utf8"
	"encoding/json"
	"kumachan/stdlib"
	"kumachan/misc/util"
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
)


func Chr(n_ *big.Int) (rune, bool) {
	if !(n_.IsUint64()) {
		return -1, false
	}
	var n = n_.Uint64()
	if n <= 0x10FFFF && !(0xD800 <= n && n <= 0xDFFF) {
		return rune(n), true
	} else {
		return -1, false
	}
}

var ContainerFunctions = map[string] Value {
	"=Char": func(a rune, b rune) bool {
		return (a == b)
	},
	"chr": func(n *big.Int) EnumValue {
		var char, ok = Chr(n)
		if ok {
			return Some(char)
		} else {
			return None()
		}
	},
	"chr!": func(n *big.Int) rune {
		var char, ok = Chr(n)
		if ok {
			return char
		} else {
			panic(fmt.Sprintf("invalid code point 0x%X", n))
		}
	},
	"seq-range-inclusive": func(l *big.Int, r *big.Int) Seq {
		if r.Cmp(l) < 0 {
			panic("invalid sequence: lower bound bigger than upper bound")
		}
		return IntervalSeq {
			Current: l,
			Bound:   big.NewInt(0).Add(r, big.NewInt(1)),
		}
	},
	"seq-range-count": func(start *big.Int, n *big.Int) Seq {
		return IntervalSeq {
			Current: start,
			Bound:   big.NewInt(0).Add(start, n),
		}
	},
	"seq-shift": func(seq Seq) EnumValue {
		var item, rest, exists = seq.Next()
		if exists {
			return Some(Tuple(item, rest))
		} else {
			return None()
		}
	},
	"seq-nil": func() Seq {
		return EmptySeq { ItemType: ValueReflectType() }
	},
	"seq-cons": func(head Value, tail Seq) Seq {
		return ConsSeq {
			Head: head,
			Tail: tail,
		}
	},
	"seq-map": func(input Seq, f Value, h InteropContext) Seq {
		return MappedSeq {
			Input:  input,
			Mapper: func(item Value) Value {
				return h.Call(f, item)
			},
		}
	},
	"seq-filter-map": func(input Seq, f Value, h InteropContext) Seq {
		return OptMappedSeq {
			Input:      input,
			MapFilter: func(item Value) Value {
				return h.Call(f, item)
			},
		}
	},
	"seq-filter": func(input Seq, f Value, h InteropContext) Seq {
		return FilteredSeq {
			Input:  input,
			Filter: func(item Value) bool {
				return FromBool(h.Call(f, item).(EnumValue))
			},
		}
	},
	"seq-flat-map": func(input Seq, f Value, h InteropContext) Seq {
		return FlatMappedSeq {
			Input:   input,
			Mapper:  func(item Value) Value {
				return h.Call(f, item)
			},
			Current: EmptySeq { ItemType: input.GetItemType() },
		}
	},
	"seq-scan": func(input Seq, init Value, f Value, h InteropContext) Seq {
		return ScannedSeq {
			Previous: init,
			Rest:     input,
			Reducer: func(prev Value, cur Value) Value {
				return h.Call(f, Tuple(prev, cur))
			},
		}
	},
	"seq-reduce": func(input Seq, init Value, f Value, h InteropContext) Value {
		return SeqReduce(input, init, func(prev Value, cur Value) Value {
			return h.Call(f, Tuple(prev, cur))
		})
	},
	"seq-some": func(input Seq, f Value, h InteropContext) EnumValue {
		return ToBool(SeqSome(input, func(item Value) bool {
			return FromBool(h.Call(f, item).(EnumValue))
		}))
	},
	"seq-every": func(input Seq, f Value, h InteropContext) EnumValue {
		return ToBool(SeqEvery(input, func(item Value) bool {
			return FromBool(h.Call(f, item).(EnumValue))
		}))
	},
	"seq-chunk": func(input Seq, size *big.Int) Value {
		return ChunkedSeq {
			ChunkSize: util.GetUintNumber(size),
			Remaining: input,
		}
	},
	"seq-collect": func(seq Seq) Value {
		return SeqCollect(seq)
	},
	"list-length": func(v Value) Value {
		var arr = ListFrom(v)
		return arr.Length()
	},
	"list-reverse": func(v Value) Value {
		var arr = ListFrom(v)
		return arr.Reversed()
	},
	"list-iterate": func(v Value) Seq {
		return ListFrom(v).Iterate()
	},
	"list-shift": func(v Value) EnumValue {
		var item, rest, ok = ListFrom(v).Shifted()
		if ok {
			return Some(Tuple(item, rest))
		} else {
			return None()
		}
	},
	"list-pop": func(v Value) EnumValue {
		var item, rest, ok = ListFrom(v).Popped()
		if ok {
			return Some(Tuple(item, rest))
		} else {
			return None()
		}
	},
	"list-unshift": func(v Value) List {
		return ListFrom(v).Unshifted()
	},
	"list-unpop": func(v Value) List {
		return ListFrom(v).Unpopped()
	},
	"Seq from String": func(s string) Seq {
		return &StringIterator { Operand: s }
	},
	"String from error": func(err error) string {
		return err.Error()
	},
	"String from List": func(v Value) Value {
		var arr = ListFrom(v)
		return arr.CopyAsString()
	},
	"String from Char": func(char rune) string {
		return string([] rune { char })
	},
	"String from Integer": func(n *big.Int) string {
		return n.String()
	},
	"String from Bool": func(p EnumValue) string {
		if FromBool(p) {
			return stdlib.Yes
		} else {
			return stdlib.No
		}
	},
	"String from Float": func(x float64) string {
		return fmt.Sprint(x)
	},
	"String from Complex": func(x complex128) string {
		return fmt.Sprint(x)
	},
	"encode-utf8": func(str string) ([] byte) {
		return StringEncode(str, UTF8)
	},
	"decode-utf8": func(bytes ([] byte)) EnumValue {
		var str, ok = StringDecode(bytes, UTF8)
		if ok {
			return Some(str)
		} else {
			return None()
		}
	},
	"force-decode-utf8": func(bytes ([] byte)) string {
		return StringForceDecode(bytes, UTF8)
	},
	"quote": func(str string) string {
		var buf = make([] rune, 0, len(str)+2)
		buf = append(buf, '"')
		for _, char := range str {
			switch char {
			case '\\', '"':
				buf = append(buf, '\\')
				buf = append(buf, char)
			case '\n':
				buf = append(buf, '\\', 'n')
			case '\r':
				buf = append(buf, '\\', 'r')
			case '\t':
				buf = append(buf, '\\', 't')
			default:
				if strconv.IsPrint(char) {
					buf = append(buf, char)
				} else {
					var bin, err = json.Marshal(string([] rune { char }))
					if err != nil { panic("something went wrong") }
					var escaped = ([] rune)(string(bin[1:len(bin)-1]))
					buf = append(buf, escaped...)
				}
			}
		}
		buf = append(buf, '"')
		return string(buf)
	},
	"unquote": func(s string) EnumValue {
		var buf strings.Builder
		if !(len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"') {
			return None()
		}
		s = s[1:len(s)-1]
		for len(s) > 0 {
			var r, _, rest, err = strconv.UnquoteChar(s, byte('"'))
			if err != nil {
				return None()
			}
			buf.WriteRune(r)
			s = rest
		}
		var unquoted = buf.String()
		return Some(unquoted)
	},
	"parse-float": func(str string) EnumValue {
		var x, err = strconv.ParseFloat(str, 64)
		if err != nil || !(util.IsNormalFloat(x)) {
			return None()
		} else {
			return Some(x)
		}
	},
	"str-concat": func(v Value) string {
		return StringConcat(ListFrom(v))
	},
	"str-contains": func(operand string, sub string) EnumValue {
		return ToBool(StringHasSubstring(operand, sub))
	},
	"str-length": func(str string) *big.Int {
		return big.NewInt(int64(len(str)))
	},
	"str-shift": func(str string) EnumValue {
		if len(str) > 0 {
			for _, char := range str {
				var rest = str[utf8.RuneLen(char):]
				return Some(Tuple(char, rest))
			}
			panic("impossible branch")
		} else {
			return None()
		}
	},
	"str-shift-prefix": func(str string, prefix string) EnumValue {
		if strings.HasPrefix(str, prefix) {
			return Some(str[len(prefix):])
		} else {
			return None()
		}
	},
	"str-split": StringSplit,
	"str-join": StringJoin,
	"trim": strings.Trim,
	"trim-left": strings.TrimLeft,
	"trim-right": strings.TrimRight,
	"trim-prefix": strings.TrimPrefix,
	"trim-suffix": strings.TrimSuffix,
	"has-prefix": func(str string, prefix string) EnumValue {
		return ToBool(strings.HasPrefix(str, prefix))
	},
	"has-suffix": func(str string, suffix string) EnumValue {
		return ToBool(strings.HasSuffix(str, suffix))
	},
	"new-set": func(cmp_ Value, values_ Value, h InteropContext) Set {
		var values = ListFrom(values_)
		var cmp = Compare(func(a Value, b Value) Ordering {
			var t = h.Call(cmp_, Tuple(a, b))
			return FromOrdering(t.(EnumValue))
		})
		var set = NewSet(cmp)
		values.ForEach(func(i uint, item Value) {
			var result, override = set.Inserted(item)
			if override {
				panic(fmt.Sprintf("duplicate set item: %s", Inspect(item)))
			}
			set = result
		})
		return set
	},
	"set-has": func(set Set, v Value) EnumValue {
		var _, exists = set.Lookup(v)
		return ToBool(exists)
	},
	"create-map-str": func(v Value) Map {
		var entries = ListFrom(v)
		var m = NewMapOfStringKey()
		entries.ForEach(func(i uint, item Value) {
			var key_, value = Tuple2From(item.(TupleValue))
			var key = key_.(string)
			var result, override = m.Inserted(key, value)
			if override {
				var key_desc = strconv.Quote(key)
				panic("duplicate map key " + key_desc)
			}
			m = result
		})
		return m
	},
	"map-entries": func(m Map) ([] TupleValue) {
		var entries = make([] TupleValue, 0, m.Size())
		m.AVL.Walk(func(v Value) {
			var entry = v.(MapEntry)
			entries = append(entries, Tuple(entry.Key, entry.Value))
		})
		return entries
	},
	"map-get": func(m Map, k Value) EnumValue {
		var v, exists = m.Lookup(k)
		if exists {
			return Some(v)
		} else {
			return None()
		}
	},
	"map-get!": func(m Map, k Value) Value {
		var v, exists = m.Lookup(k)
		if exists {
			return v
		} else {
			panic(fmt.Sprintf("accessing absent key %s of a map", Inspect(k)))
		}
	},
	"map-insert": func(m Map, entry_ Value) EnumValue {
		var entry = entry_.(TupleValue)
		var k = entry.Elements[0]
		var v = entry.Elements[1]
		var result, override = m.Inserted(k, v)
		if !(override) {
			return Some(result)
		} else {
			return None()
		}
	},
	"map-insert*": func(m Map, entry_ Value) Map {
		var entry = entry_.(TupleValue)
		var k = entry.Elements[0]
		var v = entry.Elements[1]
		var result, _ = m.Inserted(k, v)
		return result
	},
	"map-delete": func(m Map, k Value) EnumValue {
		var deleted, rest, ok = m.Deleted(k)
		if ok {
			return Some(Tuple(deleted, rest))
		} else {
			return None()
		}
	},
	"map-delete*": func(m Map, k Value) Map {
		var _, rest, _ = m.Deleted(k)
		return rest
	},
	"create-flex": func(v Value, get_key Value, h InteropContext) FlexList {
		return NewFlexList(ListFrom(v), func(item Value) string {
			return h.Call(get_key, item).(string)
		})
	},
	"create-flex-empty": func() FlexList {
		return NewFlexList(ListFrom([] Value {}), nil)
	},
	"flex-iterate": func(l FlexList) Seq {
		return FlexListIterator {
			List:      l,
			NextIndex: 0,
		}
	},
	"flex-length": func(l FlexList) *big.Int {
		return util.GetNumberUint(l.Length())
	},
	"flex-has": func(l FlexList, k string) EnumValue {
		return ToBool(l.Has(k))
	},
	"flex-get": func(l FlexList, k string) Value {
		return l.Get(k)
	},
	"flex-update": func(l FlexList, k string, f Value, h InteropContext) FlexList {
		return l.Updated(k, func(v Value) Value {
			return h.Call(f, v)
		})
	},
	"flex-delete": func(l FlexList, k string) FlexList {
		return l.Deleted(k)
	},
	"flex-prepend": func(l FlexList, entry TupleValue) FlexList {
		var key = entry.Elements[0].(string)
		var value = entry.Elements[1]
		return l.Prepended(key, value)
	},
	"flex-append": func(l FlexList, entry TupleValue) FlexList {
		var key = entry.Elements[0].(string)
		var value = entry.Elements[1]
		return l.Appended(key, value)
	},
	"flex-insert-before": func(l FlexList, pivot_ TupleValue, entry TupleValue) FlexList {
		var pivot = pivot_.Elements[0].(string)
		var key = entry.Elements[0].(string)
		var value = entry.Elements[1]
		return l.Inserted(key, value, Before, pivot)
	},
	"flex-insert-after": func(l FlexList, pivot_ TupleValue, entry TupleValue) FlexList {
		var pivot = pivot_.Elements[0].(string)
		var key = entry.Elements[0].(string)
		var value = entry.Elements[1]
		return l.Inserted(key, value, After, pivot)
	},
	"flex-move-before": func(l FlexList, key string, pivot_ TupleValue) FlexList {
		var pivot = pivot_.Elements[0].(string)
		return l.Moved(key, Before, pivot)
	},
	"flex-move-after": func(l FlexList, key string, pivot_ TupleValue) FlexList {
		var pivot = pivot_.Elements[0].(string)
		return l.Moved(key, After, pivot)
	},
	"flex-move-up": func(l FlexList, key string) FlexList {
		l, _ = l.Adjusted(key, Up)
		return l
	},
	"flex-move-down": func(l FlexList, key string) FlexList {
		l, _ = l.Adjusted(key, Down)
		return l
	},
	"flex-swap": func(l FlexList, key_a string, key_b string) FlexList {
		return l.Swapped(key_a, key_b)
	},
}

