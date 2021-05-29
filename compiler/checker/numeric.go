package checker

import (
	"math"
	"math/big"
	"strconv"
	"kumachan/stdlib"
	"kumachan/misc/util"
	"kumachan/lang/parser/ast"
)


const MaxSafeIntegerToDouble = 9007199254740991
const MinSafeIntegerToDouble = -9007199254740991

func (impl UntypedInteger) SemiExprVal() {}
type UntypedInteger struct {
	Value   *big.Int
}

func (impl IntegerLiteral) ExprVal() {}
type IntegerLiteral struct {
	Value  *big.Int
}

func (impl SmallIntLiteral) ExprVal() {}
type SmallIntLiteral struct {
	Value  interface {}
}

func (impl FloatLiteral) ExprVal() {}
type FloatLiteral struct {
	Value  float64
}


func CheckInteger(i ast.IntegerLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(i.Node)
	var chars = i.Value
	var abs_chars ([] rune)
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

func CheckFloat(f ast.FloatLiteral, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(f.Node)
	var value, err = strconv.ParseFloat(string(f.Value), 64)
	if err != nil {
		return SemiExpr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_FloatOverflow {},
		}
	}
	if !(util.IsNormalFloat(value)) {
		panic("invalid float literal got from parser")
	}
	return LiftTyped(Expr {
		Type:  &NamedType {
			Name: __NormalFloat,
			Args: make([]Type, 0),
		},
		Value: FloatLiteral { value },
		Info:  info,
	}), nil
}


func AssignIntegerTo(expected Type, integer UntypedInteger, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var assign_default = func() (Expr, *ExprError) {
		var v = integer.Value
		var t Type
		if util.IsNonNegative(v) {
			t = __T_Number
		} else {
			t = __T_Integer
		}
		var default_expr = Expr {
			Type:  t,
			Value: IntegerLiteral { v },
			Info:  info,
		}
		return TypedAssignTo(expected, default_expr, ctx)
	}
	if expected == nil {
		return assign_default()
	}
	expected_certain, err := GetCertainType(expected, info.ErrorPoint, ctx)
	if err != nil {
		return assign_default()
	}
	switch E := expected_certain.(type) {
	case *NamedType:
		var sym = E.Name
		var kind, exists = __IntegerTypeMap[sym]
		if exists {
			if len(E.Args) > 0 { panic("something went wrong") }
			var val, ok = AdaptInteger(kind, integer.Value)
			if ok {
				return Expr {
					Type:  expected_certain,
					Value: val,
					Info:  info,
				}, nil
			} else {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_IntegerOverflow { kind },
				}
			}
		}
		if sym == __Float || sym == __NormalFloat {
			var v_big = integer.Value
			if v_big.IsInt64() {
				var v = v_big.Int64()
				if MinSafeIntegerToDouble <= v && v <= MaxSafeIntegerToDouble {
					return Expr {
						Type:  expected_certain,
						Value: FloatLiteral { Value: float64(v) },
						Info:  info,
					}, nil
				}
			}
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_IntegerNotRepresentableByFloatType {},
			}
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_IntegerAssignedToNonIntegerType {
			NonIntegerType: ctx.DescribeInferredType(expected),
		},
	}
}

func AdaptInteger(expected_kind string, value *big.Int) (ExprVal, bool) {
	switch expected_kind {
	case stdlib.Integer:
		return IntegerLiteral { value }, true
	case stdlib.Number:
		if util.IsNonNegative(value) {
			return IntegerLiteral { value }, true
		} else {
			return nil, false
		}
	case stdlib.Qword:
		if value.IsUint64() {
			return SmallIntLiteral {
				Value: uint64(value.Uint64()),
			}, true
		} else {
			return nil, false
		}
	case stdlib.Dword, stdlib.Char:
		// note on the Char type:
		//   assume unsigned here and
		//   convert to singed value later (with codepoint validation)
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint32 {
				return SmallIntLiteral {
					Value: uint32(x),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case stdlib.Word:
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint16 {
				return SmallIntLiteral {
					Value: uint16(x),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case stdlib.Byte:
		if value.IsUint64() {
			var x = value.Uint64()
			if x <= math.MaxUint8 {
				return SmallIntLiteral {
					Value: uint8(x),
				}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	case stdlib.Bit:
		if value.IsUint64() {
			var x = value.Uint64()
			if x == 0 || x == 1 {
				return SmallIntLiteral {
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
