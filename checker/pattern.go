package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)


type MaybePattern interface { CheckerMaybePattern() }
func (impl Pattern) CheckerMaybePattern() {}
type Pattern struct {
	Point     ErrorPoint
	Concrete  ConcretePattern
}
type ConcretePattern interface { CheckerPattern() }
func (impl TrivialPattern) CheckerPattern() {}
type TrivialPattern struct {
	ValueName  string
	Point      ErrorPoint
}
func (impl TuplePattern) CheckerPattern() {}
type TuplePattern struct {
	ValueNames  [] string
	Points      [] ErrorPoint
}
func (impl BundlePattern) CheckerPattern() {}
type BundlePattern struct {
	ValueNames  [] string
	Points      [] ErrorPoint
}


func PatternFrom(p_node node.VariousPattern, ctx ExprContext) Pattern {
	switch p := p_node.Pattern.(type) {
	case node.PatternTrivial:
		return Pattern {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: TrivialPattern {
				ValueName: loader.Id2String(p.Name),
				Point:     ctx.GetErrorPoint(p.Name.Node),
			},
		}
	case node.PatternTuple:
		var names = make([]string, len(p.Names))
		var points = make([]ErrorPoint, len(p.Names))
		for i, identifier := range p.Names {
			names[i] = loader.Id2String(identifier)
			points[i] = ctx.GetErrorPoint(p.Names[i].Node)
		}
		return Pattern {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: TuplePattern {
				ValueNames: names,
				Points:     points,
			},
		}
	case node.PatternBundle:
		var names = make([]string, len(p.Names))
		var points = make([]ErrorPoint, len(p.Names))
		for i, identifier := range p.Names {
			names[i] = loader.Id2String(identifier)
			points[i] = ctx.GetErrorPoint(p.Names[i].Node)
		}
		return Pattern{
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: BundlePattern {
				ValueNames: names,
				Points:     points,
			},
		}
	default:
		panic("impossible branch")
	}
}

func MaybePatternFrom(p node.MaybePattern, ctx ExprContext) MaybePattern {
	switch p_node := p.(type) {
	case node.VariousPattern:
		return PatternFrom(p_node, ctx)
	default:
		return nil
	}
}


func (ctx ExprContext) WithPatternMatching (
	input    Type,
	pattern  Pattern,
	strict   bool,
) (ExprContext, *ExprError) {
	var err_result = func(e ConcreteExprError) (ExprContext, *ExprError) {
		return ExprContext{}, &ExprError {
			Point:    pattern.Point,
			Concrete: e,
		}
	}
	var check = func(added map[string]Type) (ExprContext, *ExprError) {
		var new_ctx, shadowed = ctx.WithAddedLocalValues(added)
		if shadowed != "" && !strict {
			return err_result(E_DuplicateBinding {shadowed })
		} else {
			return new_ctx, nil
		}
	}
	switch p := pattern.Concrete.(type) {
	case TrivialPattern:
		if p.ValueName == IgnoreMark {
			if strict {
				return err_result(E_EntireValueIgnored {})
			} else {
				return ctx, nil
			}
		} else {
			var added = make(map[string]Type)
			added[p.ValueName] = input
			return check(added)
		}
	case TuplePattern:
		switch tuple := UnboxTuple(input, ctx).(type) {
		case Tuple:
			var required = len(p.ValueNames)
			var given = len(tuple.Elements)
			if given != required {
				return err_result(E_TupleSizeNotMatching {
					Required:  required,
					Given:     given,
					GivenType: ctx.DescribeType(AnonymousType { tuple }),
				})
			} else {
				var added = make(map[string]Type)
				var ignored = 0
				for i, name := range p.ValueNames {
					if name == IgnoreMark {
						ignored += 1
					} else {
						var _, exists = added[name]
						if exists {
							return ExprContext{}, &ExprError {
								Point:    p.Points[i],
								Concrete: E_DuplicateBinding { name },
							}
						}
						added[name] = tuple.Elements[i]
					}
				}
				if ignored == len(p.ValueNames) {
					return err_result(E_EntireValueIgnored {})
				} else {
					return check(added)
				}
			}
		case TR_NonTuple:
			return err_result(E_MatchingNonTupleType {})
		case TR_TupleButOpaque:
			return err_result(E_MatchingOpaqueTupleType {})
		default:
			panic("impossible branch")
		}
	case BundlePattern:
		switch bundle := UnboxBundle(input, ctx).(type) {
		case Bundle:
			var added = make(map[string]Type)
			for i, name := range p.ValueNames {
				if name == IgnoreMark { panic("something went wrong") }
				var field_type, exists = bundle.Fields[name]
				if !exists {
					return ExprContext{}, &ExprError{
						Point:    p.Points[i],
						Concrete: E_FieldDoesNotExist {
							Field:  name,
							Target: ctx.DescribeType(input),
						},
					}
				}
				_, exists = added[name]
				if exists {
					return ExprContext{}, &ExprError {
						Point:    p.Points[i],
						Concrete: E_DuplicateBinding { name },
					}
				}
				added[name] = field_type
			}
			return check(added)
		case BR_NonBundle:
			return err_result(E_MatchingNonBundleType {})
		case BR_BundleButOpaque:
			return err_result(E_MatchingOpaqueBundleType {})
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}
