package loader

import (
	"fmt"
	"kumachan/transformer/node"
)

type MaybeSymbol interface { MaybeSymbol() }

func (impl Symbol) MaybeSymbol() {}
type Symbol struct {
	ModuleName  string
	SymbolName  string
}

func (sym Symbol) String() string {
	return fmt.Sprintf("%s::%s", sym.ModuleName, sym.SymbolName)
}

func Id2String(id node.Identifier) string {
	return string(id.Name)
}

func (mod *Module) SymbolFromName(name node.Identifier) Symbol {
	return Symbol {
		ModuleName: Id2String(mod.Node.Name),
		SymbolName: Id2String(name),
	}
}

func (mod *Module) SymbolFromRef(ref node.Ref) MaybeSymbol {
	var ref_mod = Id2String(ref.Module)
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		return Symbol {
			ModuleName: Id2String(corresponding.Node.Name),
			SymbolName: Id2String(ref.Id),
		}
	} else {
		return nil
	}
}