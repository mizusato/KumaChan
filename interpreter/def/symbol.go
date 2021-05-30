package def

import "fmt"


type MaybeSymbol interface{ MaybeSymbol() }
func (Symbol) MaybeSymbol() {}
type Symbol struct {
	ModuleName  string
	SymbolName  string
}

func MakeSymbol(mod string, sym string) Symbol {
	return Symbol {
		ModuleName: mod,
		SymbolName: sym,
	}
}

func (sym Symbol) String() string {
	if sym.ModuleName == "" {
		return sym.SymbolName
	} else {
		return fmt.Sprintf("%s::%s", sym.ModuleName, sym.SymbolName)
	}
}

