package docs

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"kumachan/compiler/loader"
	"kumachan/compiler/checker"
	"kumachan/stdlib"
)


type Html string

type ApiDocIndex  map[string] ModuleApiDoc

type ModuleApiDoc struct {
	Content  Html
	Outline  [] ApiItem
}
type ApiItem struct {
	Kind  ApiItemKind
	Id    string
}
type ApiItemKind string
const (
	ApiType      ApiItemKind  =  "type"
	ApiConstant  ApiItemKind  =  "constant"
	ApiFunction  ApiItemKind  =  "function"
)

func GenerateApiDocs(idx checker.Index) ApiDocIndex {
	var result = make(ApiDocIndex)
	for _, mod := range idx {
		var mod_name = mod.Name
		var reg = mod.Context.Types
		var buf strings.Builder
		var outline = make([] ApiItem, 0)
		var add_type = func(sym loader.Symbol, g *checker.GenericType) {
			outline = append(outline, ApiItem {
				Kind: ApiType,
				Id:   sym.String(),
			})
			if !(g.CaseInfo.IsCaseType) {
				var content = typeDecl(sym, g, reg, mod_name)
				var wrapped = block("api", content)
				buf.WriteString(string(wrapped))
				buf.WriteString("\n")
			}
		}
		{
			var types = make([] loader.Symbol, 0)
			for sym, _ := range reg {
				if sym.ModuleName == mod_name {
					types = append(types, sym)
				}
			}
			sortSymbols(types)
			for _, sym := range types {
				var g = reg[sym]
				add_type(sym, g)
			}
		}
		result[mod_name] = ModuleApiDoc {
			Content: Html(buf.String()),
			Outline: outline,
		}
	}
	return result
}

func typeDecl(sym loader.Symbol, g *checker.GenericType, reg checker.TypeRegistry, mod string) Html {
	return blockWithName(sym.String(), "type",
		inline("header",
			keyword("type"), text("name", sym.SymbolName),
			typeParams(g.Params, g.Defaults, g.Bounds, mod)),
		block("definition", typeDef(g, reg, mod)),
		description(g.Doc))
}

func description(content string) Html {
	var lines = strings.Split(content, "\n")
	var c_lines = make([] Html, len(lines))
	for i, line := range lines {
		c_lines[i] = block("line", escape(line))
	}
	return block("description", join(c_lines, Html("")))
}

func typeParams (
	params    ([] checker.TypeParam),
	defaults  (map[uint] checker.Type),
	bounds    checker.TypeBounds,
	mod       string,
) Html {
	if len(params) == 0 {
		return Html("")
	}
	var contents = make([] Html, len(params))
	for i, p := range params {
		var c_name = text("name", p.Name)
		var c_variance = text("variance", (func() string {
			switch p.Variance {
			case checker.Covariant:     return "+"
			case checker.Contravariant: return "-"
			default:                    return ""
			}
		})())
		var c_default = (func() Html {
			var t, has_default = defaults[uint(i)]
			if has_default {
				var no_params = make([] checker.TypeParam, 0)
				return inline("default",
					keyword("["), typeExpr(t, no_params, mod), keyword("]"))
			} else {
				return Html("")
			}
		})()
		var c_bound = (func() Html {
			var super, has_super = bounds.Super[uint(i)]
			if has_super {
				return inline("bound-super",
					keyword("<"), typeExpr(super, params, mod))
			} else {
				var sub, has_sub = bounds.Sub[uint(i)]
				if has_sub {
					return inline("bound-sub",
						keyword(">"), typeExpr(sub, params, mod))
				}
			}
			return Html("")
		})()
		var c_param = inline("type-param",
			c_default, c_variance, c_name, c_bound)
		contents[i] = c_param
	}
	return inline("list-type-param",
		keyword("["), join(contents, keyword(",")), keyword("]"))
}

func typeDef(g *checker.GenericType, reg checker.TypeRegistry, mod string) Html {
	switch d := g.Definition.(type) {
	case *checker.Native:
		return keyword("native")
	case *checker.Boxed:
		var kind = (func() string {
			if d.Opaque    { return "opaque" }
			if d.Protected { return "protected" }
			if d.Weak      { return "weak" }
			if d.Implicit  { return "implicit" }
			return ""
		})()
		switch T := d.InnerType.(type) {
		case *checker.AnonymousType:
			switch T.Repr.(type) {
			case checker.Unit:
				if kind == "" {
					return Html("")
				}
			}
		}
		var c_kind = (func() Html {
			if kind != "" {
				return block("kind", modifier(kind))
			} else {
				return block("kind")
			}
		})()
		return block("boxed",
			c_kind,
			typeExpr(d.InnerType, g.Params, mod))
	case *checker.Enum:
		var cases = make([] Html, len(d.CaseTypes))
		for i, item := range d.CaseTypes {
			cases[i] = typeDecl(item.Name, reg[item.Name], reg, mod)
		}
		return block("enum",
			modifier("enum"),
			block("cases", cases...))
	default:
		panic("impossible branch")
	}
}

func typeExpr(t checker.Type, params ([] checker.TypeParam), mod string) Html {
	switch T := t.(type) {
	case *checker.AnyType:
		return keyword("any")
	case *checker.NeverType:
		return keyword("never")
	case *checker.ParameterType:
		return text("type-parameter", params[T.Index].Name)
	case *checker.NamedType:
		var args = make([] Html, len(T.Args))
		for i, arg := range T.Args {
			args[i] = typeExpr(arg, params, mod)
		}
		var c_args = (func() Html {
			if len(args) == 0 {
				return Html("")
			} else {
				return inline("args",
					keyword("["), join(args, keyword(",")), keyword("]"))
			}
		})()
		return inline("type-named",
			inline("name", link(T.Name, mod)), c_args)
	case *checker.AnonymousType:
		switch R := T.Repr.(type) {
		case checker.Unit:
			return keyword("unit")
		case checker.Tuple:
			var elements = make([] Html, len(R.Elements))
			for i, el := range R.Elements {
				elements[i] = typeExpr(el, params, mod)
			}
			return inline("tuple",
				keyword("("), join(elements, keyword(",")), keyword(")"))
		case checker.Bundle:
			var fields = make([] Html, len(R.Fields))
			for name, f := range R.Fields {
				var t = typeExpr(f.Type, params, mod)
				fields[f.Index] = inline("field",
					text("name", name), keyword(":"), t)
			}
			return inline("bundle",
				keyword("{"), join(fields, keyword(",")), keyword("}"))
		case checker.Func:
			var in = typeExpr(R.Input, params, mod)
			var out = typeExpr(R.Output, params, mod)
			switch RT := R.Input.(type) {
			case *checker.AnonymousType:
				switch RT.Repr.(type) {
				case checker.Tuple:
					return inline("func", in, keyword("→"), out)
				}
			}
			return inline("func",
				keyword("("), in, keyword(")"), keyword("→"), out)
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}


func escape(content string) Html {
	return Html(html.EscapeString(content))
}

func join(items ([] Html), sep Html) Html {
	var buf strings.Builder
	for i, item := range items {
		if i != 0 {
			buf.WriteString(string(sep))
		}
		buf.WriteString(string(item))
	}
	return Html(buf.String())
}

func link(sym loader.Symbol, current_mod string) Html {
	var href = fmt.Sprintf("%s#%s", sym.ModuleName, sym.String())
	var c_module = (func() Html {
		var clear = checker.GetClearModuleName(sym.ModuleName, current_mod)
		if clear == "" || sym.ModuleName == stdlib.Core {
			return Html("")
		} else {
			return text("symbol-module", clear)
		}
	})()
	var c_name = text("symbol-name", sym.SymbolName)
	var c_symbol = (func() Html {
		if c_module == Html("") {
			return inline("symbol", c_name)
		} else {
			return inline("symbol", c_module, text("delimiter", "::"), c_name)
		}
	})()
	return Html(fmt.Sprintf("<a href=\"%s\" title=\"%s\">%s</a>",
		escape(href), escape(sym.String()), c_symbol))
}

func text(class string, content string) Html {
	return Html(fmt.Sprintf("<span class=\"%s\">%s</span>",
		escape(class), escape(content)))
}

func keyword(name string) Html {
	return text("keyword", name)
}

func modifier(name string) Html {
	return text("modifier", name)
}

func inline(class string, content... Html) Html {
	return Html(fmt.Sprintf("<span class=\"%s\">%s</span>",
		escape(class), join(content, Html(""))))
}

func block(class string, content... Html) Html {
	return Html(fmt.Sprintf("<div class=\"%s\">%s</div>",
		escape(class), join(content, Html(""))))
}

func blockWithName(name string, class string, content... Html) Html {
	return Html(fmt.Sprintf("<div name=\"%s\" class=\"%s\">%s</div>",
		escape(name), escape(class), join(content, Html(""))))
}


func sortSymbols(list ([] loader.Symbol)) {
	sort.Sort(SymbolList(list))
}

type SymbolList ([] loader.Symbol)
func (l SymbolList) Len() int {
	return len(l)
}
func (l SymbolList) Swap(i int, j int) {
	var t = l[i]
	l[i] = l[j]
	l[j] = t
}
func (l SymbolList) Less(i int, j int) bool {
	if l[i].ModuleName == l[j].ModuleName {
		return l[i].SymbolName < l[j].SymbolName
	} else {
		return l[i].ModuleName < l[i].ModuleName
	}
}

