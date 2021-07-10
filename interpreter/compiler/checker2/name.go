package checker2

import (
	"fmt"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/name"
	"kumachan/stdlib"
)


func NameFrom(id_mod ast.Identifier, id_item ast.Identifier, mi *ModuleInfo) name.Name {
	var ref_mod = ast.Id2String(id_mod)
	var ref_item = ast.Id2String(id_item)
	if ref_mod == "" {
		var _, is_core_type = coreTypes[ref_item]
		if is_core_type {
			// TODO: do not handle core too early (postpone to ResolveXXX)
			return name.MakeName(stdlib.Mod_core, ref_item)
		} else {
			return name.MakeName("", ref_item)
		}
	} else if ref_mod == SelfModule {
		return name.MakeName(mi.ModName, ref_item)
	} else {
		var imported, found = mi.ModImported[ref_mod]
		if found {
			return name.MakeName(imported.ModName, ref_item)
		} else {
			// TODO: should be an *source.Error
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


