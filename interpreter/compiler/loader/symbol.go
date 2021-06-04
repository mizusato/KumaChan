package loader

import (
	"kumachan/interpreter/def"
	"kumachan/interpreter/lang/textual/ast"
	"kumachan/stdlib"
)


const SelfModule = "self"

func (mod *Module) SymbolFromDeclName(name ast.Identifier) def.Symbol {
	var sym_name = ast.Id2String(name)
	return def.MakeSymbol(mod.Name, sym_name)
}

func (mod *Module) SymbolFromRef (
	ref_mod   string,
	name      string,
	use_core  bool,
) def.MaybeSymbol {
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return def.MakeSymbol(
			corresponding.Name,
			name,
		)
	} else {
		if ref_mod == "" {
			var sym_name = name
			if use_core {
				var _, exists = __CoreSymbolSet[sym_name]
				if exists {
					// core::Int, core::Real, core::Observable, ...
					return def.MakeSymbol(stdlib.Mod_core, sym_name)
				} else {
					return def.MakeSymbol("", sym_name)
				}
			} else {
				return def.MakeSymbol("", sym_name)
			}
		} else if ref_mod == SelfModule {
			var self = mod.Name
			return def.MakeSymbol(self, name)
		} else {
			return nil
		}
	}
}

func (mod *Module) TypeSymbolFromRef(ref_mod string, name string) def.MaybeSymbol {
	var self = mod.Name
	var maybe_sym = mod.SymbolFromRef(ref_mod, name, true)
	var sym, ok = maybe_sym.(def.Symbol)
	if ok {
		if sym.ModuleName == "" {
			return def.MakeSymbol(self, sym.SymbolName)
		} else {
			return sym
		}
	} else {
		return nil
	}
}

func (mod *Module) SymbolFromInlineRef(ref ast.InlineRef) def.MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	return mod.SymbolFromRef(ref_mod, name, false)
}

func (mod *Module) SymbolFromTypeRef(ref ast.TypeRef) def.MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	return mod.TypeSymbolFromRef(ref_mod, name)
}


var __CoreSymbols = stdlib.GetCoreScopedSymbols()
var __CoreSymbolSet = func() map[string] bool {
	var set = make(map[string] bool)
	for _, name := range __CoreSymbols {
		set[name] = true
	}
	return set
} ()

func IsCoreSymbol(sym def.Symbol) bool {
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

