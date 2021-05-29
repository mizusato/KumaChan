package loader

import (
	"kumachan/interpreter/base"
	"kumachan/interpreter/base/parser/ast"
	"kumachan/stdlib"
)


const SelfModule = "self"

func (mod *Module) SymbolFromDeclName(name ast.Identifier) base.Symbol {
	var sym_name = ast.Id2String(name)
	return base.MakeSymbol(mod.Name, sym_name)
}

func (mod *Module) SymbolFromRef (
	ref_mod   string,
	name      string,
	use_core  bool,
) base.MaybeSymbol {
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
		if ref_mod == "" { panic("something went wrong") }
		return base.MakeSymbol(
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
					return base.MakeSymbol(stdlib.Mod_core, sym_name)
				} else {
					return base.MakeSymbol("", sym_name)
				}
			} else {
				return base.MakeSymbol("", sym_name)
			}
		} else if ref_mod == SelfModule {
			var self = mod.Name
			return base.MakeSymbol(self, name)
		} else {
			return nil
		}
	}
}

func (mod *Module) TypeSymbolFromRef(ref_mod string, name string) base.MaybeSymbol {
	var self = mod.Name
	var maybe_sym = mod.SymbolFromRef(ref_mod, name, true)
	var sym, ok = maybe_sym.(base.Symbol)
	if ok {
		if sym.ModuleName == "" {
			return base.MakeSymbol(self, sym.SymbolName)
		} else {
			return sym
		}
	} else {
		return nil
	}
}

func (mod *Module) SymbolFromInlineRef(ref ast.InlineRef) base.MaybeSymbol {
	var ref_mod = ast.Id2String(ref.Module)
	var name = ast.Id2String(ref.Id)
	return mod.SymbolFromRef(ref_mod, name, false)
}

func (mod *Module) SymbolFromTypeRef(ref ast.TypeRef) base.MaybeSymbol {
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

func IsCoreSymbol(sym base.Symbol) bool {
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

