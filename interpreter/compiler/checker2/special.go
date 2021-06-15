package checker2

import (
	"kumachan/interpreter/compiler/checker2/typsys"
	"strings"
)

const BlankIdentifier = "_"
const ExactSuffix = "*"
const CovariantPrefix = "+"
const ContravariantPrefix = "-"

var typeBadPrefixes = [] string {
	CovariantPrefix,
	ContravariantPrefix,
}
var typeBadSuffixes = [] string {
	ExactSuffix,
}
var typeBadFullNames = [] string {
	BlankIdentifier,
	typsys.TypeNameUnit,
	typsys.TypeNameUnknown,
	typsys.TypeNameTop,
	typsys.TypeNameBottom,
}
func CheckTypeName(name string) bool {
	for _, prefix := range typeBadPrefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	for _, suffix := range typeBadSuffixes {
		if strings.HasSuffix(name, suffix) {
			return false
		}
	}
	for _, full := range typeBadFullNames {
		if name == full {
			return false
		}
	}
	return true
}


