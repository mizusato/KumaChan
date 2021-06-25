package checker2

import (
	"kumachan/interpreter/compiler/checker2/typsys"
)


const Discarded = "_" // TODO: (blank: within AST)

var typeBadNames = [] string {
	Discarded,
	typsys.TypeNameUnit,
	typsys.TypeNameUnknown,
	typsys.TypeNameTop,
	typsys.TypeNameBottom,
}
func CheckTypeName(name string) bool {
	for _, full := range typeBadNames {
		if name == full {
			return false
		}
	}
	return true
}


