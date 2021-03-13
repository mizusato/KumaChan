package loader

import (
	"fmt"
	"kumachan/stdlib"
	"kumachan/lang/parser/ast"
)


const SelfModule = "Self"

type Symbol struct {
	ModuleName  string
	SymbolName  string
}
type MaybeSymbol interface { MaybeSymbol() }
func (Symbol) MaybeSymbol() {}

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


func (mod *Module) SymbolFromDeclName(name ast.Identifier) Symbol {
	var sym_name = ast.Id2String(name)
	return MakeSymbol(mod.Name, sym_name)
}

func (mod *Module) SymbolFromRef(ref_mod string, name string, specific bool) MaybeSymbol {
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return MakeSymbol(
			corresponding.Name,
			name,
		)
	} else {
		if ref_mod == "" {
			var sym_name = name
			if specific {
				// ::Module <=> Module::Module
				return MakeSymbol(sym_name, sym_name)
			} else {
				var _, exists = __CoreSymbolSet[sym_name]
				if exists {
					// core::Int, core::Float, core::Effect, ...
					return MakeSymbol(stdlib.Mod_core, sym_name)
				} else {
					// a, b, c, Self::Type1
					return MakeSymbol("", sym_name)
				}
			}
		} else if ref_mod == SelfModule {
			var self = mod.Name
			return MakeSymbol(self, name)
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
			return MakeSymbol(self, sym.SymbolName)
		} else {
			return sym
		}
	} else {
		return nil
	}
}

func (mod *Module) SymbolFromInlineRef(ref ast.InlineRef) MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	var specific = ref.Specific
	return mod.SymbolFromRef(ref_mod, name, specific)
}

func (mod *Module) SymbolFromTypeRef(ref ast.TypeRef) MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	var specific = ref.Specific
	return mod.TypeSymbolFromRef(ref_mod, name, specific)
}


var __CoreSymbols = stdlib.GetCoreScopedSymbols()
var __CoreSymbolSet = func() map[string] bool {
	var set = make(map[string] bool)
	for _, name := range __CoreSymbols {
		set[name] = true
	}
	return set
} ()

func IsCoreSymbol(sym Symbol) bool {
	if sym.ModuleName == stdlib.Mod_core {
		var _, exists = __CoreSymbolSet[sym.SymbolName]
		if exists {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

