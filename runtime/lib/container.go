package lib

import (
	"fmt"
	"strconv"
	"math"
	"math/big"
	"encoding/json"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
)


var ContainerFunctions = map[string] Value {
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
				return BoolFrom(h.Call(f, item).(SumValue))
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
			return BoolFrom(h.Call(f, item).(SumValue))
		}))
	},
	"seq-every": func(input Seq, f Value, h InteropContext) SumValue {
		return ToBool(SeqEvery(input, func(item Value) bool {
			return BoolFrom(h.Call(f, item).(SumValue))
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
		return arr.SliceView(l.(uint), r.(uint)).CopyAsSlice()
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
		}).CopyAsSlice()
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
		return arr.CopyAsSlice()
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
	// TODO: trim, trim-left, trim-left, {has,trim}-prefix, {has,trim}-suffix
	"new-map-str": func(v Value) Map {
		var str_arr = ArrayFrom(v)
		var m = NewStrMap()
		for i := uint(0); i < str_arr.Length; i += 1 {
			var key, value = Tuple2From(str_arr.GetItem(i).(ProductValue))
			var inserted, override = m.Inserted(key.(String), value)
			if override {
				var key_desc = strconv.Quote(GoStringFromString(key.(String)))
				panic("duplicate map key " + key_desc)
			}
			m = inserted
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
}
