package checker2

import (
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/stdlib"
)


const MaxTypeParameters = 8
const Discarded = "_" // TODO: (blank: within AST)

var coreTypes = (func() (map[string] struct{}) {
	var set = make(map[string] struct{})
	var list = stdlib.CoreTypeNames()
	for _, name := range list {
		set[name] = struct{}{}
	}
	return set
})()

var typeBadNames = [] string {
	Discarded,
	typsys.TypeNameUnit,
	typsys.TypeNameUnknown,
	typsys.TypeNameTop,
	typsys.TypeNameBottom,
}
func isValidTypeItemName(name string) bool {
	for _, bad := range typeBadNames {
		if name == bad {
			return false
		}
	}
	return true
}

var functionBadNames = append([] string {}, typeBadNames...)
func isValidFunctionItemName(name string) bool {
	for _, bad := range functionBadNames {
		if name == bad {
			return false
		}
	}
	return true
}


