package lib

import (
	"fmt"
	"math/big"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
	"strconv"
	"encoding/json"
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
	"seq-map": func(input Seq, f Value, h MachineHandle) Seq {
		return MappedSeq {
			Input:  input,
			Mapper: func(item Value) Value {
				return h.Call(f, item)
			},
		}
	},
	"seq-filter": func(input Seq, f Value, h MachineHandle) Seq {
		return FilteredSeq {
			Input:  input,
			Filter: func(item Value) bool {
				return BoolFrom(h.Call(f, item).(SumValue))
			},
		}
	},
	"seq-scan": func(input Seq, init Value, f Value, h MachineHandle) Seq {
		return ScannedSeq {
			Previous: init,
			Rest:     input,
			Reducer: func(prev Value, cur Value) Value {
				return h.Call(f, ToTuple2(prev, cur))
			},
		}
	},
	"seq-reduce": func(input Seq, init Value, f Value, h MachineHandle) Value {
		return SeqReduce(input, init, func(prev Value, cur Value) Value {
			return h.Call(f, ToTuple2(prev, cur))
		})
	},
	"seq-some": func(input Seq, f Value, h MachineHandle) Value {
		return SeqSome(input, func(item Value) bool {
			return BoolFrom(h.Call(f, item).(SumValue))
		})
	},
	"seq-every": func(input Seq, f Value, h MachineHandle) Value {
		return SeqEvery(input, func(item Value) bool {
			return BoolFrom(h.Call(f, item).(SumValue))
		})
	},
	"seq-collect": func(seq Seq) Value {
		return SeqCollect(seq)
	},
	"array-get": func(av Value, index uint) Value {
		var arr = ArrayFrom(av)
		if index < arr.Length {
			return arr.GetItem(index)
		} else {
			panic(fmt.Sprintf (
				"array index out of range (%d/%d)",
				index, arr.Length,
			))
		}
	},
	"array-slice": func(av Value, range_ ProductValue) Value {
		var l, r = Tuple2From(range_)
		var arr = ArrayFrom(av)
		return arr.SliceView(l.(uint), r.(uint)).CopyAsSlice()
	},
	"array-slice-view": func(av Value, range_ ProductValue) Value {
		var l, r = Tuple2From(range_)
		var arr = ArrayFrom(av)
		return arr.SliceView(l.(uint), r.(uint))
	},
	"array-map": func(av Value, f Value, h MachineHandle) Value {
		var arr = ArrayFrom(av)
		return arr.MapView(func(item Value) Value {
			return h.Call(f, item)
		}).CopyAsSlice()
	},
	"array-map-view": func(av Value, f Value, h MachineHandle) Value {
		var arr = ArrayFrom(av)
		return arr.MapView(func(item Value) Value {
			return h.Call(f, item)
		})
	},
	"array-iterate": func(av Value) Seq {
		return ArrayFrom(av).Iterate()
	},
	"String from Int": func(n *big.Int) String {
		return String(n.String())
	},
	"String from Number": func(x uint) String {
		return String(fmt.Sprint(x))
	},
	"String from Float": func(x float64) String {
		return String(fmt.Sprint(x))
	},
	"String from Int8": func(n int8) String {
		return String(fmt.Sprint(n))
	},
	"String from Int16": func(n int16) String {
		return String(fmt.Sprint(n))
	},
	"String from Int32": func(n int32) String {
		return String(fmt.Sprint(n))
	},
	"String from Int64": func(n int64) String {
		return String(fmt.Sprint(n))
	},
	"String from Uint8": func(n uint8) String {
		return String(fmt.Sprint(n))
	},
	"String from Uint16": func(n uint16) String {
		return String(fmt.Sprint(n))
	},
	"String from Uint32": func(n uint32) String {
		return String(fmt.Sprint(n))
	},
	"String from Uint64": func(n uint64) String {
		return String(fmt.Sprint(n))
	},
	"encode-utf8": func(str String) []byte {
		return StringEncode(str, UTF8)
	},
	"decode-utf8": func(bytes []byte) SumValue {
		var str, ok = StringDecode(bytes, UTF8)
		if ok {
			return Just(str)
		} else {
			return Na()
		}
	},
	"quote": func(str String) String {
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
		return buf
	},
	"str-concat": func(av Value) String {
		return StringConcat(ArrayFrom(av))
	},
	"str-find": func(str String, sub String) SumValue {
		var index, ok = StringFind(str, sub)
		if ok {
			return Just(index)
		} else {
			return Na()
		}
	},
	"str-split": func(str String, sep String) Seq {
		return StringSplit(str, sep)
	},
	"str-join": func(seq Seq, sep String) String {
		return StringJoin(seq, sep)
	},
	// TODO: trim, trim-left, trim-left, {has,trim}-prefix, {has,trim}-suffix
	"new-map-str": func(v Value) Map {
		var str_arr = ArrayFrom(v)
		var m = NewStrMap()
		for i := uint(0); i < str_arr.Length; i += 1 {
			var key, value = Tuple2From(str_arr.GetItem(i).(ProductValue))
			m = m.Inserted(key.(String), value)
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
}
