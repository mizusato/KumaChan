package checker2

import (
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
)



type TypeRegistry (map[name.TypeName] *typsys.TypeDef)

func RegisterTypes(idx loader.Index) (TypeRegistry, source.Errors) {
	panic("not implemented")  // TODO
}


