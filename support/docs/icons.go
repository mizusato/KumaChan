package docs

import (
	"fmt"
	"kumachan/util"
	"kumachan/runtime/lib/ui/qt"
)


var icons = (func() (map[string] *qt.ImageData) {
	var result = make(map[string] *qt.ImageData)
	var names = [] string { "module", "type", "const", "func" }
	for _, name := range names {
		var file = fmt.Sprintf("icons/%s.png", name)
		result[name] = &qt.ImageData {
			Data:   util.ReadInterpreterResource(file),
			Format: qt.PNG,
		}
	}
	return result
})()

func apiKindToIcon(kind ApiItemKind) *qt.ImageData {
	switch kind {
	case TypeDecl:
		return icons["type"]
	case ConstDecl:
		return icons["const"]
	case FuncDecl:
		return icons["func"]
	default:
		return nil
	}
}
