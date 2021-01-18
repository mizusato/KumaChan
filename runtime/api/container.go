package api

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"math/big"
	"encoding/json"
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
)


var ContainerFunctions = map[string] Value {
	"chr": func(n uint) SumValue {
		if n <= 0x10FFFF && !(0xD800 <= n && n <= 0xDFFF) {
			return Just(uint32(n))
		} else {
			return Na()
		}
	},
	"chr!": func(n uint) uint32 {
		if n <= 0x10FFFF && !(0xD800 <= n && n <= 0xDFFF) {
			return uint32(n)
		} else {
			panic(fmt.Sprintf("invalid code point 0x%X", n))
		}
	},
	"range-iterate": func(l uint, r uint) Seq {
		return RangeSeq {
			Current: l,
			Bound:   r,
		}
	},
	"seq-next": func(seq Seq) SumValue {
		var item, rest, exists = seq.Next()
		if exists {
			return Just(ToTuple2(item, rest))
		} else {
			return Na()
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
	"seq-map?": func(input Seq, f Value, h InteropContext) Seq {
		// TODO: consider changing "map?" to a more clear name
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
	"array-at": func(v Value, index uint) SumValue {
		var arr = ArrayFrom(v)
		if index < arr.Length {
			return Just(arr.GetItem(index))
		} else {
			return Na()
		}
	},
	"array-at!": func(v Value, index uint) Value {
		var arr = ArrayFrom(v)
		if index < arr.Length {
			return arr.GetItem(index)
		} else {
			panic(fmt.Sprintf (
				"array index out of range (%d/%d)",
				index, arr.Length,
			))
		}
	},
	"array-length": func(v Value) Value {
		var arr = ArrayFrom(v)
		return arr.Length
	},
	"array-reverse": func(v Value) Value {
		var arr = ArrayFrom(v)
		return arr.Reversed()
	},
	"array-slice": func(v Value, range_ ProductValue) Value {
		var l, r = Tuple2From(range_)
		var arr = ArrayFrom(v)
		return arr.SliceView(l.(uint), r.(uint)).CopyAsSlice(arr.ItemType)
	},
	"array-slice-view": func(v Value, range_ ProductValue) Value {
		var l, r = Tuple2From(range_)
		var arr = ArrayFrom(v)
		return arr.SliceView(l.(uint), r.(uint))
	},
	"array-map": func(v Value, f Value, h InteropContext) Value {
		var arr = ArrayFrom(v)
		return arr.MapView(func(item Value) Value {
			return h.Call(f, item)
		}).CopyAsSlice(arr.ItemType)
	},
	"array-map-view": func(v Value, f Value, h InteropContext) Value {
		var arr = ArrayFrom(v)
		return arr.MapView(func(item Value) Value {
			return h.Call(f, item)
		})
	},
	"array-iterate": func(v Value) Seq {
		return ArrayFrom(v).Iterate()
	},
	"String from error": func(err error) String {
		return StringFromGoString(err.Error())
	},
	"String from Array": func(v Value) Value {
		var arr = ArrayFrom(v)
		return arr.CopyAsSlice(reflect.TypeOf(Char(0)))
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
	"String from Float": func(x float64) String {
		return StringFromGoString(fmt.Sprint(x))
	},
	"encode-utf8": func(str String) []byte {
		return StringEncode(str, UTF8)
	},
	"decode-utf8": func(bytes ([] byte)) SumValue {
		var str, ok = StringDecode(bytes, UTF8)
		if ok {
			return Just(str)
		} else {
			return Na()
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
		for len(s) > 0 {
			var r, _, rest, err = strconv.UnquoteChar(s, byte('"'))
			if err != nil {
				return Na()
			}
			buf = append(buf, Char(r))
			s = rest
		}
		return Just(buf)
	},
	"parse-float": func(str String) SumValue {
		var x, err = strconv.ParseFloat(GoStringFromString(str), 64)
		if err != nil {
			return Na()
		} else {
			if math.IsInf(x, 0) || math.IsNaN(x) {
				return Na()
			} else {
				return Just(x)
			}
		}
	},
	"str-concat": func(v Value) String {
		return StringConcat(ArrayFrom(v))
	},
	"str-find": func(str String, sub String) SumValue {
		var index, ok = StringFind(str, sub)
		if ok {
			return Just(index)
		} else {
			return Na()
		}
	},
	"str-split": StringSplit,
	"str-join": StringJoin,
	"trim": StringTrim,
	"trim-left": StringTrimLeft,
	"trim-right": StringTrimRight,
	"trim-prefix": StringTrimPrefix,
	"trim-suffix": StringTrimSuffix,
	"new-set": func(cmp_ Value, values_ Value, h InteropContext) Set {
		var values = ArrayFrom(values_)
		var cmp = Compare(func(a Value, b Value) Ordering {
			var t = h.Call(cmp_, &ValProd{Elements: [] Value { a, b }})
			return FromOrdering(t.(SumValue))
		})
		var set = NewSet(cmp)
		for i := uint(0); i < values.Length; i += 1 {
			var item = values.GetItem(i)
			var result, override = set.Inserted(item)
			if override {
				panic(fmt.Sprintf("duplicate set item: %s", Inspect(item)))
			}
			set = result
		}
		return set
	},
	"set-has": func(set Set, v Value) SumValue {
		var _, exists = set.Lookup(v)
		return ToBool(exists)
	},
	"create-map-str": func(v Value) Map {
		var str_arr = ArrayFrom(v)
		var m = NewStrMap()
		for i := uint(0); i < str_arr.Length; i += 1 {
			var key, value = Tuple2From(str_arr.GetItem(i).(ProductValue))
			var result, override = m.Inserted(key.(String), value)
			if override {
				var key_desc = strconv.Quote(GoStringFromString(key.(String)))
				panic("duplicate map key " + key_desc)
			}
			m = result
		}
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
			return Just(v)
		} else {
			return Na()
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
	"map-insert*": func(m Map, k Value, v Value) Map {
		var result, _ = m.Inserted(k, v)
		return result
	},
	"map-delete*": func(m Map, k Value) Map {
		var _, result, _ = m.Deleted(k)
		return result
	},
	"create-list": func(v Value, get_key Value, h InteropContext) List {
		return NewList(ArrayFrom(v), func(item Value) String {
			return h.Call(get_key, item).(String)
		})
	},
	"empty-list": func() List {
		return NewList(ArrayFrom([] Value {}), nil)
	},
	"list-iterate": func(l List) Seq {
		return ListIterator {
			List:      l,
			NextIndex: 0,
		}
	},
	"list-length": func(l List) uint {
		return l.Length()
	},
	"list-has": func(l List, k String) SumValue {
		return ToBool(l.Has(k))
	},
	"list-get": func(l List, k String) Value {
		return l.Get(k)
	},
	"list-update": func(l List, k String, f Value, h InteropContext) List {
		return l.Updated(k, func(v Value) Value {
			return h.Call(f, v)
		})
	},
	"list-delete": func(l List, k String) List {
		return l.Deleted(k)
	},
	"list-prepend": func(l List, entry ProductValue) List {
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Prepended(key, value)
	},
	"list-append": func(l List, entry ProductValue) List {
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Appended(key, value)
	},
	"list-insert-before": func(l List, pivot_ ProductValue, entry ProductValue) List {
		var pivot = pivot_.Elements[0].(String)
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Inserted(key, value, Before, pivot)
	},
	"list-insert-after": func(l List, pivot_ ProductValue, entry ProductValue) List {
		var pivot = pivot_.Elements[0].(String)
		var key = entry.Elements[0].(String)
		var value = entry.Elements[1]
		return l.Inserted(key, value, After, pivot)
	},
	"list-move-before": func(l List, key String, pivot_ ProductValue) List {
		var pivot = pivot_.Elements[0].(String)
		return l.Moved(key, Before, pivot)
	},
	"list-move-after": func(l List, key String, pivot_ ProductValue) List {
		var pivot = pivot_.Elements[0].(String)
		return l.Moved(key, After, pivot)
	},
	"list-move-up": func(l List, key String) List {
		l, _ = l.Adjusted(key, Up)
		return l
	},
	"list-move-down": func(l List, key String) List {
		l, _ = l.Adjusted(key, Down)
		return l
	},
	"list-swap": func(l List, key_a String, key_b String) List {
		return l.Swapped(key_a, key_b)
	},
}

