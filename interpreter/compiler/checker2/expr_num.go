package checker2

import (
	"math"
	"unicode"
	"math/big"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/standalone/util"
)


type smallNumericTypeInfo struct {
	which  nominalType
	adapt  smallNumericTypeAdapter
}
type smallNumericTypeAdapter func(*big.Int)(interface{}, bool)

var smallNumericTypes = [] smallNumericTypeInfo {
	{ which: coreNormalFloat, adapt: normalFloatAdapt },
	{ which: coreFloat, adapt: normalFloatAdapt },
	{ which: coreQword, adapt: uintAdapt(math.MaxUint64, func(v *big.Int) interface{} {
		return v.Uint64()
	}) },
	{ which: coreDword, adapt: uintAdapt(math.MaxUint32, func(v *big.Int) interface{} {
		return uint32(v.Uint64())
	}) },
	{ which: coreChar, adapt: uintAdapt(unicode.MaxRune, func(v *big.Int) interface{} {
		return rune(v.Uint64())
	}) },
	{ which: coreWord, adapt: uintAdapt(math.MaxUint16, func(v *big.Int) interface{} {
		return uint16(v.Uint64())
	}) },
	{ which: coreByte, adapt: uintAdapt(math.MaxUint8, func(v *big.Int) interface{} {
		return byte(v.Uint64())
	}) },
}

func normalFloatAdapt(v *big.Int) (interface{}, bool) {
	var float_v, ok = util.IntegerToDouble(v)
	return float_v, ok
}
func intAdapt(min int64, max uint64, cast func(*big.Int)(interface{})) smallNumericTypeAdapter {
	return smallNumericTypeAdapter(func(value *big.Int) (interface{}, bool) {
		var min = big.NewInt(0).SetInt64(min)
		var max = big.NewInt(0).SetUint64(max)
		if (min.Cmp(value) <= 0) && (value.Cmp(max) <= 0) {
			return cast(value), true
		} else {
			return nil, false
		}
	})
}
func uintAdapt(max uint64, cast func(*big.Int)(interface{})) smallNumericTypeAdapter {
	return intAdapt(0, max, cast)
}


func checkChar(C ast.CharLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(C.Location, &s, ctx, nil)
		var value, ok = util.ParseRune(C.Value)
		if !(ok) {
			return cc.error(
				E_InvalidChar { Content: string(C.Value) })
		}
		return cc.assign(
			expected, coreChar(ctx.Types),
			checked.NumericLiteral { Value: value })
	})
}

func checkFloat(F ast.FloatLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(F.Location, &s, ctx, nil)
		var value, ok = util.ParseDouble(F.Value)
		if !(ok) {
			return cc.error(
				E_FloatOverflowUnderflow {})
		}
		if !(util.IsNormalFloat(value)) {
			panic("invalid float literal got from parser")
		}
		return cc.assign(expected, coreNormalFloat(ctx.Types),
			checked.NumericLiteral { Value: value })
	})
}

func checkInteger(I ast.IntegerLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(I.Location, &s, ctx, nil)
		var value, ok = util.WellBehavedParseInteger(I.Value)
		if !(ok) { panic("something went wrong") }
		var big_min_t = (func() typsys.Type {
			if util.IsNonNegative(value) {
				return coreNumber(ctx.Types)
			} else {
				return coreInteger(ctx.Types)
			}
		})()
		for _, t := range smallNumericTypes {
			if t.which.isEqualTo(expected, ctx.Types) {
				var adapted, ok = t.adapt(value)
				if ok {
					return cc.ok(expected,
						checked.NumericLiteral { Value: adapted })
				} else {
					var _, is_float = adapted.(float64)
					if is_float {
						return cc.error(
							E_IntegerNotRepresentableByFloatType {})
					} else {
						return cc.error(
							E_IntegerOverflowUnderflow {
								TypeName: cc.describeType(expected),
							})
					}
				}
			} else {
				// continue
			}
		}
		return cc.assign(
			expected, big_min_t,
			checked.NumericLiteral { Value: value })
	})
}


