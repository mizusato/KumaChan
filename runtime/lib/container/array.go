package container

import . "kumachan/runtime/common"

type Array struct {
	Length   uint
	GetItem  func(uint) Value
}

func ArrayFromSeq(seq Seq) Array {
	var item, rest, exists = seq.Next()
	if exists {
		var inline_acc = make([]uint64, 0)
		var variant_acc = make([]Value, 0)
		for exists {
			switch plain := item.(type) {
			case PlainValue:
				if plain.Pointer == nil && len(variant_acc) == 0 {
					inline_acc = append(inline_acc, plain.Inline)
					continue
				}
			}
			if len(inline_acc) > 0 {
				for _, I := range inline_acc {
					variant_acc = append(variant_acc, PlainValue { Inline: I })
				}
				variant_acc = append(variant_acc, item)
				inline_acc = inline_acc[:0]
			} else {
				variant_acc = append(variant_acc, item)
			}
			item, rest, exists = rest.Next()
		}
		if len(inline_acc) > 0 {
			return Array {
				Length:  uint(len(inline_acc)),
				GetItem: func(i uint) Value {
					return PlainValue { Inline: inline_acc[i] }
				},
			}
		} else {
			return Array {
				Length: uint(len(variant_acc)),
				GetItem: func(i uint) Value {
					return variant_acc[i]
				},
			}
		}
	} else {
		return Array { Length: 0 }
	}
}

type ArrayIterator struct {
	Array      Array
	NextIndex  uint
}

type ArrayReversedIterator struct {
	Array      Array
	NextIndex  uint
}

func (it ArrayIterator) Next() (Value, Seq, bool) {
	var array = it.Array
	var next = it.NextIndex
	if next < array.Length {
		return array.GetItem(next), ArrayIterator {
			Array:     array,
			NextIndex: (next + 1),
		}, true
	} else {
		return nil, nil, false
	}
}

func (it ArrayReversedIterator) Next() (Value, Seq, bool) {
	var array = it.Array
	var next = it.NextIndex
	if next >= 0 {
		return array.GetItem(next), ArrayIterator {
			Array:     array,
			NextIndex: (next - 1),
		}, true
	} else {
		return nil, nil, false
	}
}

func (array Array) Iterate() Seq {
	return ArrayIterator {
		Array:     array,
		NextIndex: 0,
	}
}

func (array Array) IterateReversed() Seq {
	return ArrayReversedIterator {
		Array:     array,
		NextIndex: (array.Length - 1),
	}
}

func (array Array) Map(f func(Value)Value) Array {
	return Array {
		Length:  array.Length,
		GetItem: func(i uint) Value {
			return f(array.GetItem(i))
		},
	}
}
