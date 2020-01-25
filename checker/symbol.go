package checker

import "kumachan/transformer/node"

type MaybeSymbol interface { MaybeSymbol() }

func (impl Symbol) MaybeSymbol() {}
type Symbol struct {
	ModuleName  string
	SymbolName  string
}

func Id2String(id *node.Identifier) string {
	return string(id.Name)
}

func SymbolFromId(id *node.Identifier, current_mod string) Symbol {
	return Symbol {
		ModuleName: current_mod,
		SymbolName: Id2String(id),
	}
}

func SymbolFromRef(ref *node.Ref, current_mod string) Symbol {
	var mod string
	if ref.Module.Name == nil {
		mod = current_mod
	} else {
		mod = Id2String(&ref.Module)
	}
	return Symbol {
		ModuleName: mod,
		SymbolName: Id2String(&ref.Id),
	}
}
