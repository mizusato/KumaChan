package checker

import (
	"kumachan/loader"
	"kumachan/transformer/ast"
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
	ValueType  Type
	Point      ErrorPoint
}
func (impl TuplePattern) CheckerPattern() {}
type TuplePattern struct {
	Items  [] PatternItem
}
func (impl BundlePattern) CheckerPattern() {}
type BundlePattern struct {
	Items  [] PatternItem
}

type PatternItem struct {
	Name   string
	Index  uint
	Type   Type
	Point  ErrorPoint
}


func PatternFrom (
	p_node  ast.VariousPattern,
	input   Type,
	ctx     ExprContext,
) (Pattern, *ExprError) {
	var err_result = func(e ConcreteExprError) (Pattern, *ExprError) {
		return Pattern{}, &ExprError {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: e,
		}
	}
	switch p := p_node.Pattern.(type) {
	case ast.PatternTrivial:
		return Pattern {
			Point:    ctx.GetErrorPoint(p_node.Node),
			Concrete: TrivialPattern {
				ValueName: loader.Id2String(p.Name),
				ValueType: input,
				Point:     ctx.GetErrorPoint(p.Name.Node),
			},
		}, nil
	case ast.PatternTuple:
		switch tuple := UnboxTuple(input, ctx).(type) {
		case Tuple:
			var required = len(p.Names)
			var given = len(tuple.Elements)
			if given != required {
				return err_result(E_TupleSizeNotMatching {
					Required:  required,
					Given:     given,
					GivenType: ctx.DescribeType(AnonymousType { tuple }),
				})
			} else {
				var occurred = make(map[string]bool)
				var ignored = 0
				var items = make([]PatternItem, 0)
				for i, identifier := range p.Names {
					var name = loader.Id2String(identifier)
					if name == IgnoreMark {
						ignored += 1
					} else {
						var _, duplicate = occurred[name]
						if duplicate {
							return Pattern{}, &ExprError {
								Point:    ctx.GetErrorPoint(identifier.Node),
								Concrete: E_DuplicateBinding { name },
							}
						}
						occurred[name] = true
						items = append(items, PatternItem {
							Name:  loader.Id2String(identifier),
							Index: uint(i),
							Type:  tuple.Elements[i],
							Point: ctx.GetErrorPoint(identifier.Node),
						})
					}
				}
				if ignored == len(p.Names) {
					return err_result(E_EntireValueIgnored {})
				} else {
					return Pattern {
						Point:    ctx.GetErrorPoint(p_node.Node),
						Concrete: TuplePattern { items },
					}, nil
				}
			}
		case TR_NonTuple:
			return err_result(E_MatchingNonTupleType {})
		case TR_TupleButOpaque:
			return err_result(E_MatchingOpaqueTupleType {})
		default:
			panic("impossible branch")
		}
	case ast.PatternBundle:
		switch bundle := UnboxBundle(input, ctx).(type) {
		case Bundle:
			var occurred = make(map[string]bool)
			var items = make([]PatternItem, len(p.Names))
			for i, identifier := range p.Names {
				var name = loader.Id2String(identifier)
				var field, exists = bundle.Fields[name]
				if exists && name == IgnoreMark {
					// field should not be named using IgnoreMark;
					// IgnoreMark used in bundle pattern considered
					//   as "field does not exist" error
					panic("something went wrong")
				}
				if !exists {
					return Pattern{}, &ExprError {
						Point:    ctx.GetErrorPoint(identifier.Node),
						Concrete: E_FieldDoesNotExist {
							Field:  name,
							Target: ctx.DescribeType(input),
						},
					}
				}
				var _, duplicate = occurred[name]
				if duplicate {
					return Pattern{}, &ExprError {
						Point:    ctx.GetErrorPoint(identifier.Node),
						Concrete: E_DuplicateBinding { name },
					}
				}
				occurred[name] = true
				items[i] = PatternItem {
					Name:  loader.Id2String(identifier),
					Index: field.Index,
					Type:  field.Type,
					Point: ctx.GetErrorPoint(identifier.Node),
				}
			}
			return Pattern {
				Point:    ctx.GetErrorPoint(p_node.Node),
				Concrete: BundlePattern { items },
			}, nil
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


func (ctx ExprContext) WithShadowingPatternMatching(p Pattern) ExprContext {
	var added = make(map[string]Type)
	switch P := p.Concrete.(type) {
	case TrivialPattern:
		added[P.ValueName] = P.ValueType
	case TuplePattern:
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
	case BundlePattern:
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
	default:
		panic("impossible branch")
	}
	var new_ctx, _ = ctx.WithAddedLocalValues(added)
	return new_ctx
}

func (ctx ExprContext) WithPatternMatching(p Pattern) (ExprContext, *ExprError) {
	var err_result = func(e ConcreteExprError) (ExprContext, *ExprError) {
		return ExprContext{}, &ExprError {
			Point:    p.Point,
			Concrete: e,
		}
	}
	var check = func(added map[string]Type) (ExprContext, *ExprError) {
		var new_ctx, shadowed = ctx.WithAddedLocalValues(added)
		if shadowed != "" {
			return err_result(E_DuplicateBinding { shadowed })
		} else {
			return new_ctx, nil
		}
	}
	switch P := p.Concrete.(type) {
	case TrivialPattern:
		if P.ValueName == IgnoreMark {
			return err_result(E_EntireValueIgnored {})
		} else {
			var added = make(map[string]Type)
			added[P.ValueName] = P.ValueType
			return check(added)
		}
	case TuplePattern:
		var added = make(map[string]Type)
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
		return check(added)
	case BundlePattern:
		var added = make(map[string]Type)
		for _, item := range P.Items {
			added[item.Name] = item.Type
		}
		return check(added)
	default:
		panic("impossible branch")
	}
}
