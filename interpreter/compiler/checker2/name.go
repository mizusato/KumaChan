package checker2

import (
	"fmt"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
)


func NameFrom (
	mod_node   ast.ModuleName,
	item_node  ast.Identifier,
	mi         *ModuleInfo,
) (name.Name, *source.Error) {
	var mod_node_not_empty = mod_node.NotEmpty
	var mod = ast.Id2String(mod_node.Identifier)
	var item = ast.Id2String(item_node)
	if mod == "" {
		if mod_node_not_empty {
			return name.MakeName(CoreModule, item), nil
		} else {
			return name.MakeName("", item), nil
		}
	} else if mod == SelfModule {
		return name.MakeName(mi.ModName, item), nil
	} else {
		var imported, found = mi.ModImported[mod]
		if found {
			return name.MakeName(imported.ModName, item), nil
		} else {
			return name.Name {}, source.MakeError(mod_node.Location,
				E_ModuleNotFound {
					ModuleName: mod,
				})
		}
	}
}

func DescribeNameWithPossibleAlias(n name.Name, to name.Name) string {
	if to != (name.Name {}) {
		return fmt.Sprintf("%s (aka %s)", n, to)
	} else {
		return n.String()
	}
}


