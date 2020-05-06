package loader

import (
	"fmt"
	"kumachan/stdlib"
	"kumachan/parser/ast"
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

func (mod *Module) SymbolFromDeclName(name ast.Identifier) Symbol {
	var sym_name = Id2String(name)
	return NewSymbol(mod.Name, sym_name)
}

func (mod *Module) SymbolFromRef(ref_mod string, name string, specific bool) MaybeSymbol {
	var self = mod.Name
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return NewSymbol (
			corresponding.Name,
			name,
		)
	} else {
		if ref_mod == "" {
			var sym_name = name
			if specific {
				// ::Module <=> Module::Module
				return NewSymbol(sym_name, sym_name)
			} else {
				var _, exists = __PreloadCoreSymbolSet[sym_name]
				if exists {
					// Core::Int, Core::Float, Core::Effect, ...
					return NewSymbol(stdlib.Core, sym_name)
				} else {
					// a, b, c, Self::Type1
					return NewSymbol("", sym_name)
				}
			}
		} else if ref_mod == self {
			return NewSymbol(self, name)
		} else {
			return nil
		}
	}
}

func (mod *Module) TypeSymbolFromRef(ref_mod string, name string, specific bool) MaybeSymbol {
	var self = mod.Name
	var maybe_sym = mod.SymbolFromRef(ref_mod, name, specific)
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

func (mod *Module) SymbolFromInlineRef(ref ast.InlineRef) MaybeSymbol {
	var ref_mod = Id2String(ref.Module)
	var name = Id2String(ref.Id)
	var specific = ref.Specific
	return mod.SymbolFromRef(ref_mod, name, specific)
}

func (mod *Module) SymbolFromTypeRef(ref ast.TypeRef) MaybeSymbol {
	var ref_mod = Id2String(ref.Module)
	var name = Id2String(ref.Id)
	var specific = ref.Specific
	return mod.TypeSymbolFromRef(ref_mod, name, specific)
}


var __PreloadCoreSymbols = stdlib.GetCoreScopedSymbols()
var __PreloadCoreSymbolSet = func() map[string] bool {
	var set = make(map[string] bool)
	for _, name := range __PreloadCoreSymbols {
		set[name] = true
	}
	return set
} ()
func IsPreloadCoreSymbol(sym Symbol) bool {
	if sym.ModuleName == stdlib.Core {
		var _, exists = __PreloadCoreSymbolSet[sym.SymbolName]
		if exists {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
