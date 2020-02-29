package checker

import "kumachan/loader"

const IgnoreMarker = "_"
/* should be consistent with stdlib/core.km */
var __Maybe = loader.NewSymbol(loader.CoreModule, "Maybe")
// var __Just uint = 0
var __Nothing uint = 1
var __Float = loader.NewSymbol(loader.CoreModule, "Float")
var __String = loader.NewSymbol(loader.CoreModule, "String")
var __Array = loader.NewSymbol(loader.CoreModule, "Array")

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
