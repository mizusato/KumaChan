package checker

import (
	"math"
	"math/big"
)

func AdaptInteger(expected_kind string, value *big.Int) (ExprVal, bool) {
	switch expected_kind {
	case "Int":
		return IntLiteral { value }, true
	case "Int64":
		if value.IsInt64() {
			return SmallIntLiteral {
				Kind:  "Int64",
				Value: uint64(value.Int64()),
			}, true
		} else {
			return nil, false
		}
	case "Uint64", "Qword":
		if value.IsUint64() {
			return SmallIntLiteral {
				Kind:  "Uint64",
				Value: value.Uint64(),
			}, true
		} else {
			return nil, false
		}
	case "Int32":
		if value.IsInt64() {
			var x = value.Int64()
			if math.MinInt32 <= x && x <= math.MaxInt32 {
				return SmallIntLiteral {
					Kind:  "Int32",
					Value: uint64(uint32(int32(x))),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Uint32", "Dword", "Char":
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint32 {
				return SmallIntLiteral {
					Kind:  "Uint32",
					Value: x,
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Int16":
		if value.IsInt64() {
			var x = value.Int64()
			if math.MinInt16 <= x && x <= math.MaxInt16 {
				return SmallIntLiteral {
					Kind:  "Int16",
					Value: uint64(uint16(int16(x))),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Uint16", "Word":
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint16 {
				return SmallIntLiteral {
					Kind:  "Uint16",
					Value: x,
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Int8":
		if value.IsInt64() {
			var x = value.Int64()
			if math.MinInt8 <= x && x <= math.MaxInt8 {
				return SmallIntLiteral {
					Kind:  "Int8",
					Value: uint64(uint8(int8(x))),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Uint8", "Byte":
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint8 {
				return SmallIntLiteral {
					Kind:  "Uint8",
					Value: x,
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case "Bit":
		if value.IsUint64() {
			var x = value.Uint64()
			if x == 0 || x == 1 {
				return SmallIntLiteral {
					Kind:  "Bit",
					Value: x,
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	default:
		panic("impossible branch")
	}
}
