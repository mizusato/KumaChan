package loader

import (
	"fmt"
	"kumachan/stdlib"
	"kumachan/transformer/ast"
)


type MaybeSymbol interface { MaybeSymbol() }

func (Symbol) MaybeSymbol() {}
type Symbol struct {
	ModuleName  string
	SymbolName  string
}

func (sym Symbol) String() string {
	if sym.ModuleName == "" {
		return sym.SymbolName
	} else {
		return fmt.Sprintf("%s::%s", sym.ModuleName, sym.SymbolName)
	}
}

func NewSymbol (mod string, sym string) Symbol {
	return Symbol {
		ModuleName: mod,
		SymbolName: sym,
	}
}

func Id2String(id ast.Identifier) string {
	return string(id.Name)
}

func (mod *Module) SymbolFromName(name ast.Identifier) Symbol {
	var sym_name = Id2String(name)
	var _, exists = __PreloadCoreSymbolSet[sym_name]
	if exists {
		return NewSymbol(stdlib.Core, sym_name)
	} else {
		return NewSymbol(Id2String(mod.Node.Name), sym_name)
	}
}

func (mod *Module) SymbolFromRef(ref ast.Ref) MaybeSymbol {
	var ref_mod = Id2String(ref.Module)
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return NewSymbol (
			Id2String(corresponding.Node.Name),
			Id2String(ref.Id),
		)
	} else {
		if ref_mod == "" {
			var sym_name = Id2String(ref.Id)
			if ref.Specific {
				// ::Module <=> Module::Module
				return NewSymbol(sym_name, sym_name)
			} else {
				var _, exists = __PreloadCoreSymbolSet[sym_name]
				if exists {
					// Core::Int, Core::Float, Core::Effect, ...
					return NewSymbol(stdlib.Core, sym_name)
				} else {
					// a, b, c
					return NewSymbol("", sym_name)
				}
			}
		} else {
			return nil
		}
	}
}

func (mod *Module) TypeSymbolFromRef(ref ast.Ref) MaybeSymbol {
	var self = Id2String(mod.Node.Name)
	var maybe_sym = mod.SymbolFromRef(ref)
	var sym, ok = maybe_sym.(Symbol)
	if ok {
		if sym.ModuleName == "" {
			return NewSymbol(self, sym.SymbolName)
		} else {
			return sym
		}
	} else {
		return nil
	}
}


var __PreloadCoreSymbols = stdlib.GetCoreTypes()
var __PreloadCoreSymbolSet = func() map[string] bool {
	var set = make(map[string] bool)
	for _, name := range __PreloadCoreSymbols {
		set[name] = true
	}
	return set
} ()
func IsPreloadCoreSymbol (sym Symbol) bool {
	var _, exists = __PreloadCoreSymbolSet[sym.SymbolName]
	if exists {
		if sym.ModuleName != stdlib.Core {
			panic("invalid symbol: preload symbol must belong to CoreModule")
		}
		return true
	} else {
		return false
	}
}
