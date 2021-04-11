package api

import (
	"fmt"
	"math"
	"strconv"
	"math/big"
	"encoding/json"
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
	"kumachan/stdlib"
)


var ContainerFunctions = map[string] Value {
	"chr": func(n uint) SumValue {
		if n <= 0x10FFFF && !(0xD800 <= n && n <= 0xDFFF) {
			return Some(uint32(n))
		} else {
			return None()
		}
	},
	"chr!": func(n uint) uint32 {
		if n <= 0x10FFFF && !(0xD800 <= n && n <= 0xDFFF) {
			return uint32(n)
		} else {
			panic(fmt.Sprintf("invalid code point 0x%X", n))
		}
	},
	"seq-range-inclusive": func(l uint, r uint) Seq {
		if r < l { panic("invalid sequence: lower bound bigger than upper bound") }
		return IntervalSeq {
			Current: l,
			Bound:   (r + 1),
		}
	},
	"seq-range-count": func(start uint, n uint) Seq {
		return IntervalSeq {
			Current: start,
			Bound:   (start + n),
		}
	},
	"seq-next": func(seq Seq) SumValue {
		var item, rest, exists = seq.Next()
		if exists {
			return Some(ToTuple2(item, rest))
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
				return FromBool(h.Call(f, item).(SumValue))
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
				return h.Call(f, ToTuple2(prev, cur))
			},
		}
	},
	"seq-reduce": func(input Seq, init Value, f Value, h InteropContext) Value {
		return SeqReduce(input, init, func(prev Value, cur Value) Value {
			return h.Call(f, ToTuple2(prev, cur))
		})
	},
	"seq-some": func(input Seq, f Value, h InteropContext) SumValue {
		return ToBool(SeqSome(input, func(item Value) bool {
			return FromBool(h.Call(f, item).(SumValue))
		}))
	},
	"seq-every": func(input Seq, f Value, h InteropContext) SumValue {
		return ToBool(SeqEvery(input, func(item Value) bool {
			return FromBool(h.Call(f, item).(SumValue))
		}))
	},
	"seq-chunk": func(input Seq, size uint) Value {
		return ChunkedSeq {
			ChunkSize: size,
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
	"list-shift": func(v Value) SumValue {
		var item, rest, ok = ListFrom(v).Shifted()
		if ok {
			return Some(&ValProd { Elements: [] Value { item, rest } })
		} else {
			return None()
		}
	},
	"list-pop": func(v Value) SumValue {
		var item, rest, ok = ListFrom(v).Popped()
		if ok {
			return Some(&ValProd { Elements: [] Value { item, rest } })
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
	"String from error": func(err error) String {
		return StringFromGoString(err.Error())
	},
	"String from List": func(v Value) Value {
		var arr = ListFrom(v)
		return arr.CopyAsString()
	},
	"String from Char": func(char Char) String {
		return [] Char { char }
	},
	"String from Int": func(n *big.Int) String {
		return StringFromGoString(n.String())
	},
	"String from Number": func(x uint) String {
		return StringFromGoString(fmt.Sprint(x))
	},
	"String from Bool": func(p SumValue) String {
		if FromBool(p) {
			return StringFromGoString(stdlib.Yes)
		} else {
			return StringFromGoString(stdlib.No)
		}
	},
	"String from Float": func(x float64) String {
		return StringFromGoString(fmt.Sprint(x))
	},
	"encode-utf8": func(str String) []byte {
		return StringEncode(str, UTF8)
	},
	"decode-utf8": func(bytes ([] byte)) SumValue {
		var str, ok = StringDecode(bytes, UTF8)
		if ok {
			return Some(str)
		} else {
			return None()
		}
	},
	"force-decode-utf8": func(bytes ([] byte)) String {
		return StringForceDecode(bytes, UTF8)
	},
	"quote": func(str String) String {
		var buf = make([] Char, 0, len(str)+2)
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
				if strconv.IsPrint(rune(char)) {
					buf = append(buf, char)
				} else {
					var bin, err = json.Marshal(string([] rune { rune(char) }))
					if err != nil { panic("something went wrong") }
					var escaped = StringFromGoString(string(bin[1:len(bin)-1]))
					buf = append(buf, escaped...)
				}
			}
		}
		buf = append(buf, '"')
		return buf
	},
	"unquote": func(str String) SumValue {
		var buf = make([] Char, 0)
		var s = GoStringFromString(str)
		if !(len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"') {
			return None()
		}
		s = s[1:len(s)-1]
		for len(s) > 0 {
			var r, _, rest, err = strconv.UnquoteChar(s, byte('"'))
			if err != nil {
				return None()
			}
			buf = append(buf, Char(r))
			s = rest
		}
		return Some(buf)
	},
	"parse-real": func(str String) SumValue {
		var x, err = strconv.ParseFloat(GoStringFromString(str), 64)
		if err != nil {
			return None()
		} else {
			if math.IsInf(x, 0) || math.IsNaN(x) {
				return None()
			} else {
				return Some(x)
			}
		}
	},
	"substr": func(str String, interval ProductValue) String {
		var l = interval.Elements[0].(uint)
		var r = interval.Elements[1].(uint)
		return StringCopy(StringSliceView(str, l, r))
	},
	"substr-view": func(str String, interval ProductValue) String {
		var l = interval.Elements[0].(uint)
		var r = interval.Elements[1].(uint)
		return StringSliceView(str, l ,r)
	},
	"str-concat": func(v Value) String {
		return StringConcat(ListFrom(v))
	},
	"str-shift": func(str String) SumValue {
		if len(str) > 0 {
			return Some(&ValProd { Elements: [] Value { str[0], str[1:] } })
		} else {
			return None()
		}
	},
	"str-shift-prefix": func(str String, prefix String) SumValue {
		if StringHasPrefix(str, prefix) {
			return Some(str[len(prefix):])
		} else {
			return None()
		}
	},
	"str-find": func(str String, sub String) SumValue {
		var index, ok = StringFind(str, sub)
		if ok {
			return Some(index)
		} else {
			return None()
		}
	},
	"str-split": StringSplit,
	"str-join": StringJoin,
	"trim": StringTrim,
	"trim-left": StringTrimLeft,
	"trim-right": StringTrimRight,
	"trim-prefix": StringTrimPrefix,
	"trim-suffix": StringTrimSuffix,
	"has-prefix": func(str String, prefix String) SumValue {
		return ToBool(StringHasPrefix(str, prefix))
	},
	"has-suffix": func(str String, suffix String) SumValue {
		return ToBool(StringHasSuffix(str, suffix))
	},
	"new-set": func(cmp_ Value, values_ Value, h InteropContext) Set {
		var values = ListFrom(values_)
		var cmp = Compare(func(a Value, b Value) Ordering {
			var t = h.Call(cmp_, &ValProd{Elements: [] Value { a, b }})
			return FromOrdering(t.(SumValue))
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
	"set-has": func(set Set, v Value) SumValue {
		var _, exists = set.Lookup(v)
		return ToBool(exists)
	},
	"create-map-str": func(v Value) Map {
		var entries = ListFrom(v)
		var m = NewMapOfStringKey()
		entries.ForEach(func(i uint, item Value) {
			var key, value = Tuple2From(item.(ProductValue))
			var result, override = m.Inserted(key.(String), value)
			if override {
				var key_desc = strconv.Quote(GoStringFromString(key.(String)))
				panic("duplicate map key " + key_desc)
			}
			m = result
		})
		return m
	},
	"map-entries": func(m Map) ([] ProductValue) {
		var entries = make([] ProductValue, 0, m.Size())
		m.AVL.Walk(func(v Value) {
			var entry = v.(MapEntry)
			entries = append(entries, &ValProd {
				Elements: [] Value { entry.Key, entry.Value },
			})
		})
		return entries
	},
	"map-get": func(m Map, k Value) SumValue {
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
	"map-insert": func(m Map, entry_ Value) SumValue {
		var entry = entry_.(ProductValue)
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
		var entry = entry_.(ProductValue)
		var k = entry.Elements[0]
		var v = entry.Elements[1]
		var result, _ = m.Inserted(k, v)
		return result
	},
	"map-delete": func(m Map, k Value) SumValue {
		var deleted, rest, ok = m.Deleted(k)
		if ok {
			return Some(&ValProd { Elements: [] Value { deleted, rest } })
		} else {
			return None()
		}
	},
	"map-delete*": func(m Map, k Value) Map {
		var _, rest, _ = m.Deleted(k)
		return rest
	},
	"create-list": func(v Value, get_key Value, h InteropContext) FlexList {
		return NewFlexList(ListFrom(v), func(item Value) String {
			return h.Call(get_key, item).(String)
		})
	},
	"empty-list": func() FlexList {
		return NewFlexList(ListFrom([] Value {}), nil)
	},
	"flex-iterate": func(l FlexList) Seq {
		return FlexListIterator {
			List:      l,
			NextIndex: 0,
		}
	},
	"flex-length": func(l FlexList) uint {
		return l.Length()
	},
	"flex-has": func(l FlexList, k String) SumValue {
		return ToBool(l.Has(k))
	},
	"flex-get": func(l FlexList, k String) Value {
		return l.Get(k)
	},
	"flex-update": func(l FlexList, k String, f Value, h InteropContext) FlexList {
		return l.Updated(k, func(v Value) Value {
			return h.Call(f, v)
		})
	},
	"flex-delete": func(l FlexList, k String) FlexList {
		return l.Deleted(k)
	},
	"flex-prepend": func(l FlexList, entry ProductValue) FlexList {
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Prepended(key, value)
	},
	"flex-append": func(l FlexList, entry ProductValue) FlexList {
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Appended(key, value)
	},
	"flex-insert-before": func(l FlexList, pivot_ ProductValue, entry ProductValue) FlexList {
		var pivot = pivot_.Elements[0].(String)
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Inserted(key, value, Before, pivot)
	},
	"flex-insert-after": func(l FlexList, pivot_ ProductValue, entry ProductValue) FlexList {
		var pivot = pivot_.Elements[0].(String)
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Inserted(key, value, After, pivot)
	},
	"flex-move-before": func(l FlexList, key String, pivot_ ProductValue) FlexList {
		var pivot = pivot_.Elements[0].(String)
		return l.Moved(key, Before, pivot)
	},
	"flex-move-after": func(l FlexList, key String, pivot_ ProductValue) FlexList {
		var pivot = pivot_.Elements[0].(String)
		return l.Moved(key, After, pivot)
	},
	"flex-move-up": func(l FlexList, key String) FlexList {
		l, _ = l.Adjusted(key, Up)
		return l
	},
	"flex-move-down": func(l FlexList, key String) FlexList {
		l, _ = l.Adjusted(key, Down)
		return l
	},
	"flex-swap": func(l FlexList, key_a String, key_b String) FlexList {
		return l.Swapped(key_a, key_b)
	},
}

