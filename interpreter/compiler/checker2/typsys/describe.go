package typsys

import (
	"fmt"
	"strings"
)


const TypeNameUnknown = "unknown"
const TypeNameUnit = "unit"
const TypeNameTop = "any"
const TypeNameBottom = "never"

func DescribeType(t Type, s *InferringState) string {
	switch T := t.(type) {
	case *UnknownType:
		return TypeNameUnknown
	case UnitType:
		return TypeNameUnit
	case TopType:
		return TypeNameTop
	case BottomType:
		return TypeNameBottom
	case ParameterType:
		if s != nil {
			var ps, exists = s.mapping[T.Parameter]
			if exists {
				var name = T.Parameter.Name
				var op = ps.status.OperatorString()
				var current = DescribeType(ps.currentInferred, nil)
				return fmt.Sprintf("[%s(%s)%s]", name, op, current)
			} else {
				return T.Parameter.Name
			}
		} else {
			return T.Parameter.Name
		}
	case *NestedType:
		switch N := T.Content.(type) {
		case Ref:
			return fmt.Sprintf("%s[%s]", N.Def.Name, DescribeTypeVec(N.Args, s))
		case Tuple:
			return fmt.Sprintf("(%s)", DescribeTypeVec(N.Elements, s))
		case Record:
			return fmt.Sprintf("{%s}", DescribeFieldVec(N.Fields, s))
		case Lambda:
			var input = DescribeType(N.Input, s)
			var output = DescribeType(N.Output, s)
			return fmt.Sprintf("&(%s)=>(%s)", input, output)
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func DescribeTypeVec(types ([] Type), s *InferringState) string {
	var mapped = make([] string, len(types))
	for i, t := range types {
		mapped[i] = DescribeType(t, s)
	}
	return strings.Join(mapped, ",")
}

func DescribeFieldVec(fields ([] Field), s *InferringState) string {
	var mapped = make([] string, len(fields))
	for i, f := range fields {
		mapped[i] = fmt.Sprintf("%s:%s", f.Name, DescribeType(f.Type, s))
	}
	return strings.Join(mapped, ",")
}


