package checker

import (
	"strings"
	"kumachan/compiler/loader/parser/ast"
)


func DocStringFromRaw(raw ([] ast.Doc)) string {
	var buf strings.Builder
	for _, line := range raw {
		var t = string(line.RawContent)
		t = strings.TrimPrefix(t, "///")
		t = strings.TrimPrefix(t, " ")
		t = strings.TrimRight(t, " \r")
		buf.WriteString(t)
		buf.WriteRune('\n')
	}
	return buf.String()
}

