package loader

import (
	"fmt"
	"kumachan/transformer/node"
)

const CoreModule = "Core"

type MaybeSymbol interface { MaybeSymbol() }

func (impl Symbol) MaybeSymbol() {}
type Symbol struct {
	ModuleName  string
	SymbolName  string
}

func (sym Symbol) String() string {
	return fmt.Sprintf("%s::%s", sym.ModuleName, sym.SymbolName)
}

func NewSymbol (mod string, sym string) Symbol {
	return Symbol {
		ModuleName: mod,
		SymbolName: sym,
	}
}

func Id2String(id node.Identifier) string {
	return string(id.Name)
}

func (mod *Module) SymbolFromName(name node.Identifier) Symbol {
	var sym_name = Id2String(name)
	var _, exists = __PreloadCoreSymbolSet[sym_name]
	if exists {
		return NewSymbol(CoreModule, sym_name)
	} else {
		return NewSymbol(Id2String(mod.Node.Name), sym_name)
	}
}

func (mod *Module) SymbolFromRef(ref node.Ref) MaybeSymbol {
	var self = Id2String(mod.Node.Name)
	var ref_mod = Id2String(ref.Module)
	var corresponding, exists = mod.ImpMap[ref_mod]
	if exists {
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
					return NewSymbol(CoreModule, sym_name)
				} else {
					// Self::SymbolName
					return NewSymbol(self, sym_name)
				}
			}
		} else {
			return nil
		}
	}
}


var __PreloadCoreSymbols = []string {
	"Bit", "Byte", "Word", "Dword", "Qword",
	"Int",
	"Bytes", "String", "Seq", "Array", "Heap", "Set", "Map",
	"Effect",
	"Int64", "Uint64", "Int32", "Uint32", "Int16", "Uint16", "Int8", "Uint8",
	"Char",
	"Float64", "Float",
	"Bool", "Yes", "No",
	"Maybe", "Just", "Nothing",
	"Result", "OK", "NG",
	"List",
}
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
		if sym.ModuleName != CoreModule {
			panic("invalid symbol: preload symbol must belong to CoreModule")
		}
		return true
	} else {
		return false
	}
}