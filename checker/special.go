package checker

import "kumachan/loader"

/* should be consistent with stdlib/core.km */
var __Maybe = loader.NewSymbol(loader.CoreModule, "Maybe")
// var __Just uint = 0
var __Nothing uint = 1

func IsMaybeType (t Type) bool {
	switch T := t.(type) {
	case NamedType:
		return T.Name == __Maybe
	default:
		return false
	}
}

func UnitWithIndex (index uint, info ExprInfo) ExprVal {
	return Sum {
		Value: Expr {
			Type: AnonymousType { Unit {} },
			Value: UnitValue {},
			Info: info,
		},
		Index: index,
	}
}
