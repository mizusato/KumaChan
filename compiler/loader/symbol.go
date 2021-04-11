package loader

import (
	"kumachan/lang"
	"kumachan/lang/parser/ast"
	"kumachan/stdlib"
)


const SelfModule = "self"

func (mod *Module) SymbolFromDeclName(name ast.Identifier) lang.Symbol {
	var sym_name = ast.Id2String(name)
	return lang.MakeSymbol(mod.Name, sym_name)
}

func (mod *Module) SymbolFromRef (
	ref_mod   string,
	name      string,
	specific  bool,
	use_core  bool,
) lang.MaybeSymbol {
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return lang.MakeSymbol(
			corresponding.Name,
			name,
		)
	} else {
		if ref_mod == "" {
			var sym_name = name
			if specific {
				// ::Module <=> Module::Module
				return lang.MakeSymbol(sym_name, sym_name)
			} else if use_core {
				var _, exists = __CoreSymbolSet[sym_name]
				if exists {
					// core::Int, core::Real, core::Observable, ...
					return lang.MakeSymbol(stdlib.Mod_core, sym_name)
				} else {
					return lang.MakeSymbol("", sym_name)
				}
			} else {
				return lang.MakeSymbol("", sym_name)
			}
		} else if ref_mod == SelfModule {
			var self = mod.Name
			return lang.MakeSymbol(self, name)
		} else {
			return nil
		}
	}
}

func (mod *Module) TypeSymbolFromRef(ref_mod string, name string, specific bool) lang.MaybeSymbol {
	var self = mod.Name
	var maybe_sym = mod.SymbolFromRef(ref_mod, name, specific, true)
	var sym, ok = maybe_sym.(lang.Symbol)
	if ok {
		if sym.ModuleName == "" {
			return lang.MakeSymbol(self, sym.SymbolName)
		} else {
			return sym
		}
	} else {
		return nil
	}
}

func (mod *Module) SymbolFromInlineRef(ref ast.InlineRef) lang.MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	var specific = ref.Specific
	return mod.SymbolFromRef(ref_mod, name, specific, false)
}

func (mod *Module) SymbolFromTypeRef(ref ast.TypeRef) lang.MaybeSymbol {
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

func IsCoreSymbol(sym lang.Symbol) bool {
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

