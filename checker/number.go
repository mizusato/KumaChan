package checker

import (
	"kumachan/transformer/node"
	"math"
	"math/big"
	"strconv"
)


func (impl UntypedInteger) SemiExprVal() {}
type UntypedInteger struct {
	Value   *big.Int
}

func (impl IntLiteral) ExprVal() {}
type IntLiteral struct {
	Value  *big.Int
}

func (impl SmallIntLiteral) ExprVal() {}
type SmallIntLiteral struct {
	Kind   string
	Value  interface {}
}

func (impl FloatLiteral) ExprVal() {}
type FloatLiteral struct {
	Value  float64
}


func CheckInteger(i node.IntegerLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(i.Node)
	var chars = i.Value
	var abs_chars []rune
	if chars[0] == '-' {
		abs_chars = chars[1:]
	} else {
		abs_chars = chars
	}
	var has_base_prefix = false
	if len(abs_chars) >= 2 {
		var c1 = abs_chars[0]
		var c2 = abs_chars[1]
		if c1 == '0' && (c2 == 'x' || c2 == 'o' || c2 == 'b' || c2 == 'X' || c2 == 'O' || c2 == 'B') {
			has_base_prefix = true
		}
	}
	var str = string(chars)
	var value *big.Int
	var ok bool
	if has_base_prefix {
		value, ok = big.NewInt(0).SetString(str, 0)
	} else {
		// avoid "0" prefix to be recognized as octal with base 0
		value, ok = big.NewInt(0).SetString(str, 10)
	}
	if ok {
		return SemiExpr {
			Value: UntypedInteger { value },
			Info:  info,
		}, nil
	} else {
		return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_InvalidInteger { str },
		}
	}
}

func CheckFloat(f node.FloatLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(f.Node)
	var value, err = strconv.ParseFloat(string(f.Value), 64)
	if err != nil { panic("invalid float literal got from parser") }
	return LiftTyped(Expr {
		Type:  NamedType {
			Name: __Float,
			Args: make([]Type, 0),
		},
		Value: FloatLiteral { value },
		Info:  info,
	}), nil
}


func AssignIntegerTo(expected Type, integer UntypedInteger, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var _, err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	switch E := expected.(type) {
	case NamedType:
		var sym = E.Name
		var kind, exists = __IntegerTypeMap[sym]
		if exists {
			if len(E.Args) > 0 { panic("something went wrong") }
			var val, ok = AdaptInteger(kind, integer.Value)
			if ok {
				return Expr {
					Type:  expected,
					Info:  info,
					Value: val,
				}, nil
			} else {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_IntegerOverflow { kind },
				}
			}
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_IntegerAssignedToNonIntegerType {},
	}
}

func AdaptInteger(expected_kind string, value *big.Int) (ExprVal, bool) {
	switch expected_kind {
	case "Int":
		return IntLiteral { value }, true
	case "Int64":
		if value.IsInt64() {
			return SmallIntLiteral {
				Kind:  "Int64",
				Value: int64(value.Int64()),
			}, true
		} else {
			return nil, false
		}
	case "Uint64", "Qword":
		if value.IsUint64() {
			return SmallIntLiteral {
				Kind:  "Uint64",
				Value: uint64(value.Uint64()),
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
					Value: int32(x),
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
					Value: uint32(x),
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
					Value: int16(x),
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
					Value: uint16(x),
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
					Value: int8(x),
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
					Value: uint8(x),
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
					Value: (x == 1),
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
