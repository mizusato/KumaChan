package container

import (
	. "kumachan/lang"
	"reflect"
)

type Bytes = []byte

type ByteOrder int
const (
	BigEndian  ByteOrder  =  iota
	LittleEndian
)

func BytesFromBitSeq(seq Seq) Bytes {
	var buf = make(Bytes, 0)
	var written_bits = uint(8)
	for v,r,exists := seq.Next(); exists; v,r,exists = r.Next() {
		if written_bits == 8 {
			buf = append(buf, 0)
			written_bits = 0
		}
		if v.(bool) {
			buf[len(buf)-1] |= (1 << written_bits)
			written_bits += 1
		} else {
			written_bits += 1
		}
	}
	return buf
}

func BytesFromByteSeq(seq Seq) Bytes {
	var buf = make(Bytes, 0)
	for v,r,exists := seq.Next(); exists; v,r,exists = r.Next() {
		buf = append(buf, FromByte(v))
	}
	return buf
}

func BytesFromWordSeq(seq Seq, ord ByteOrder) Bytes {
	var buf = make(Bytes, 0)
	var chunk [2]byte
	for v,r,exists := seq.Next(); exists; v,r,exists = r.Next() {
		chunk = EncodeWord(FromWord(v), ord)
		buf = append(buf, chunk[0], chunk[1])
	}
	return buf
}

func BytesFromDwordSeq(seq Seq, ord ByteOrder) Bytes {
	var buf = make(Bytes, 0)
	var chunk [4]byte
	for v,r,exists := seq.Next(); exists; v,r,exists = r.Next() {
		chunk = EncodeDword(FromDword(v), ord)
		buf = append(buf, chunk[0], chunk[1], chunk[2], chunk[3])
	}
	return buf
}

func BytesFromQwordSeq(seq Seq, ord ByteOrder) Bytes {
	var buf = make(Bytes, 0)
	var chunk [8]byte
	for v,r,exists := seq.Next(); exists; v,r,exists = r.Next() {
		chunk = EncodeQword(FromQword(v), ord)
		for i := 0; i < 8; i += 1 {
			buf = append(buf, chunk[i])
		}
	}
	return buf
}

func BitArray(bytes Bytes) Array {
	return Array {
		Length: uint(len(bytes)) << 3,
		GetItem: func(i uint) Value {
			var n = i >> 3
			var offset = uint8(1) << (i & 7)
			return ToBool(1 == (uint8(bytes[n]) & offset))
		},
		ItemType: reflect.TypeOf(true),
	}
}

func ByteArray(bytes Bytes) Array {
	return Array {
		Length:  uint(len(bytes)),
		GetItem: func(i uint) Value {
			return bytes[i]
		},
		ItemType: reflect.TypeOf(uint8(0)),
	}
}

func WordArray(bytes Bytes, ord ByteOrder) Array {
	var L = uint(len(bytes))
	var length uint
	if L % 2 == 0 {
		length = (L / 2)
	} else {
		length = (L / 2) + 1
	}
	return Array {
		Length: length,
		GetItem: func(i uint) Value {
			var p = 2*i
			var chunk [2]byte
			var size uint = 2
			for (p + size) >= length { size -= 1 }
			copy(chunk[:], bytes[p: p+size])
			return uint16(DecodeWord(chunk, ord))
		},
		ItemType: reflect.TypeOf(uint16(0)),
	}
}

func DwordArray(bytes Bytes, ord ByteOrder) Array {
	var L = uint(len(bytes))
	var length uint
	if L % 4 == 0 {
		length = (L / 4)
	} else {
		length = (L / 4) + 1
	}
	return Array {
		Length: length,
		GetItem: func(i uint) Value {
			var p = 4*i
			var chunk [4]byte
			var size uint = 4
			for (p + size) >= length { size -= 1 }
			copy(chunk[:], bytes[p: p+size])
			return uint32(DecodeDword(chunk, ord))
		},
		ItemType: reflect.TypeOf(uint32(0)),
	}
}

func QwordArray(bytes Bytes, ord ByteOrder) Array {
	var L = uint(len(bytes))
	var length uint
	if L % 8 == 0 {
		length = (L / 8)
	} else {
		length = (L / 8) + 1
	}
	return Array {
		Length: length,
		GetItem: func(i uint) Value {
			var p = 8*i
			var chunk [8]byte
			var size uint = 8
			for (p + size) >= length { size -= 1 }
			copy(chunk[:], bytes[p: p+size])
			return uint64(DecodeQword(chunk, ord))
		},
		ItemType: reflect.TypeOf(uint64(0)),
	}
}

func DecodeWord(chunk [2]byte, ord ByteOrder) uint16 {
	switch ord {
	case BigEndian:
		return (uint16(chunk[0]) << 8) | uint16(chunk[1])
	case LittleEndian:
		return (uint16(chunk[1]) << 8) | uint16(chunk[0])
	default:
		panic("impossible branch")
	}
}

func EncodeWord(w uint16, ord ByteOrder) [2]byte {
	var chunk [2]byte
	switch ord {
	case BigEndian:
		chunk[0] = byte(w >> 8)
		chunk[1] = byte(w)
	case LittleEndian:
		chunk[1] = byte(w >> 8)
		chunk[0] = byte(w)
	default:
		panic("impossible branch")
	}
	return chunk
}

func DecodeDword(chunk [4]byte, ord ByteOrder) uint32 {
	var dw uint32 = 0
	switch ord {
	case BigEndian:
		for i := uint(0); i < 4; i += 1 {
			dw |= uint32(chunk[3-i]) << (i << 3)
		}
	case LittleEndian:
		for i := uint(0); i < 4; i += 1 {
			dw |= uint32(chunk[i]) << (i << 3)
		}
	default:
		panic("impossible branch")
	}
	return dw
}

func EncodeDword(dw uint32, ord ByteOrder) [4]byte {
	var chunk [4]byte
	switch ord {
	case BigEndian:
		for i := uint(0); i < 4; i += 1 {
			chunk[3-i] = byte(dw >> (i << 3))
		}
	case LittleEndian:
		for i := uint(0); i < 4; i += 1 {
			chunk[i] = byte(dw >> (i << 3))
		}
	default:
		panic("impossible branch")
	}
	return chunk
}

func DecodeQword(chunk [8]byte, ord ByteOrder) uint64 {
	var qw uint64 = 0
	switch ord {
	case BigEndian:
		for i := uint(0); i < 8; i += 1 {
			qw |= uint64(chunk[7-i]) << (i << 3)
		}
	case LittleEndian:
		for i := uint(0); i < 8; i += 1 {
			qw |= uint64(chunk[i]) << (i << 3)
		}
	default:
		panic("impossible branch")
	}
	return qw
}

func EncodeQword(qw uint64, ord ByteOrder) [8]byte {
	var chunk [8]byte
	switch ord {
	case BigEndian:
		for i := uint(0); i < 8; i += 1 {
			chunk[7-i] = byte(qw >> (i << 3))
		}
	case LittleEndian:
		for i := uint(0); i < 8; i += 1 {
			chunk[i] = byte(qw >> (i << 3))
		}
	default:
		panic("impossible branch")
	}
	return chunk
}
