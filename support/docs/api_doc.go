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
	Mod   string
	Name  string
	Sec   [] string
}
type ApiItemKind string
const (
	TypeDecl  ApiItemKind = "type"
	ConstDecl ApiItemKind = "constant"
	FuncDecl  ApiItemKind = "function"
)

func GenerateApiDocs(idx checker.Index) ApiDocIndex {
	var result = make(ApiDocIndex)
	for _, mod := range idx {
		var buf strings.Builder
		var mod_name = mod.Name
		buf.WriteString(string(block("title", text("text", mod_name))))
		var reg = mod.Context.Types
		var outline = make([] ApiItem, 0)
		var sections = make(map[string] ([] int))
		var outline_add_item = func(kind ApiItemKind, sym loader.Symbol, sec ([] string)) {
			var normalized_sec = make([] string, 0)
			for _, s := range sec {
				if s != "" {
					normalized_sec = append(normalized_sec, s)
				}
			}
			if len(normalized_sec) == 0 {
				normalized_sec = nil
			}
			var i = len(outline)
			outline = append(outline, ApiItem {
				Kind: kind,
				Id:   sym.String(),
				Mod:  sym.ModuleName,
				Name: sym.SymbolName,
				Sec:  normalized_sec,
			})
			for _, s := range sec {
				var index_list, exists = sections[s]
				if exists {
					sections[s] = append(index_list, i)
				} else {
					sections[s] = [] int { i }
				}
			}
		}
		var add_type = func(sym loader.Symbol, g *checker.GenericType) {
			outline_add_item(TypeDecl, sym, [] string { g.Section })
			var id = sym.String()
			if !(g.CaseInfo.IsCaseType) {
				var content = typeDecl(sym, g, reg, mod_name)
				var wrapped = blockWithId(id, "api toplevel", content)
				buf.WriteString(string(wrapped))
				buf.WriteString("\n")
			}
		}
		var add_const = func(name string, constant checker.CheckedConstant) {
			var sym = loader.MakeSymbol(mod_name, name)
			outline_add_item(ConstDecl, sym, [] string { constant.Section })
			var id = sym.String()
			var content = constDecl(sym, constant.Type, constant.Doc)
			var wrapped = blockWithId(id, "api toplevel", content)
			buf.WriteString(string(wrapped))
			buf.WriteString("\n")
		}
		var add_function = func(name string, group ([] checker.CheckedFunction)) {
			var sym = loader.MakeSymbol(mod_name, name)
			var sec = make([] string, 0)
			var sec_added = make(map[string] bool)
			for _, f := range group {
				var s = f.Section
				if sec_added[s] { continue }
				sec_added[s] = true
				sec = append(sec, s)
			}
			sort.Strings(sec)
			outline_add_item(FuncDecl, sym, sec)
			var id = sym.String()
			var filtered = make([] checker.CheckedFunction, 0)
			for _, f := range group {
				if !(f.IsSelfAlias) {
					filtered = append(filtered, f)
				}
			}
			if len(filtered) == 0 {
				return
			}
			var content = funcDecl(sym, filtered)
			var wrapped = blockWithId(id, "api toplevel", content)
			buf.WriteString(string(wrapped))
			buf.WriteString("\n")
		}
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
		var constants = make([] string, 0)
		for name, constant := range mod.Constants {
			if constant.Public {
				constants = append(constants, name)
			}
		}
		sort.Strings(constants)
		for _, name := range constants {
			var constant = mod.Constants[name]
			add_const(name, constant)
		}
		var functions = make([] string, 0)
		var function_groups = make(map[string] ([] checker.CheckedFunction))
		for name, group := range mod.Functions {
			var public_subgroup = make([] checker.CheckedFunction, 0)
			for _, f := range group {
				if f.Public {
					public_subgroup = append(public_subgroup, f)
				}
			}
			if len(public_subgroup) > 0 {
				functions = append(functions, name)
				function_groups[name] = public_subgroup
			}
		}
		sort.Strings(functions)
		for _, name := range functions {
			var group = function_groups[name]
			add_function(name, group)
		}
		result[mod_name] = ModuleApiDoc {
			Content: Html(buf.String()),
			Outline: outline,
		}
	}
	return result
}

func splitDescription(content string) (string, string) {
	var index = strings.Index(content, "\n\n")
	if index != -1 {
		var common = content[:index]
		var individual = ""
		var after = (index + 2)
		if after < len(content) {
			individual = content[after:]
		}
		return common, individual
	} else {
		return "", content
	}
}

func typeDecl(sym loader.Symbol, g *checker.GenericType, reg checker.TypeRegistry, mod string) Html {
	return block("type",
		block("header",
			keyword("type"), text("name", sym.SymbolName),
			typeParams(g.Params, g.Defaults, g.Bounds, mod)),
		block("definition", typeDef(g, reg, mod)),
		description(g.Doc))
}

func constDecl(sym loader.Symbol, t checker.Type, doc string) Html {
	return block("constant",
		block("header", keyword("const"), text("name", sym.SymbolName),
			keyword(":"), inline("type", typeExpr(t, nil, sym.ModuleName))),
		description(doc))
}

func funcDecl(sym loader.Symbol, group ([] checker.CheckedFunction)) Html {
	var group_contents = make([] Html, len(group))
	var buf strings.Builder
	for i, f := range group {
		group_contents[i] = funcOverload(sym, f)
		var common, _ = splitDescription(f.Doc)
		buf.WriteString(common)
		buf.WriteRune('\n')
	}
	var common_desc = buf.String()
	return block("function",
		block("header", keyword("function"), text("name", sym.SymbolName)),
		block("group", group_contents...),
		description(common_desc))
}

func funcOverload(sym loader.Symbol, f checker.CheckedFunction) Html {
	var no_defaults = make(map[uint] checker.Type)
	var params = f.Params
	var mod = sym.ModuleName
	var _, desc = splitDescription(f.Doc)
	return block("overload",
		block("header",
			funcAlias(mod, f.AliasList),
			funcTypeParams (
				typeParams(params, no_defaults, f.Bounds, mod),
				(len(f.Bounds.Super) + len(f.Bounds.Sub)) == 0,
			),
			funcImplicit(f.RawImplicit, params, mod),
			inline("type", typeExpr(f.Type, params, mod))),
		description(desc),
	)
}

func funcAlias(mod string, alias_list ([] string)) Html {
	if len(alias_list) == 0 {
		return Html("")
	}
	var alias_contents = make([] Html, len(alias_list))
	for i, alias := range alias_list {
		var id = (loader.MakeSymbol(mod, alias)).String()
		alias_contents[i] = inlineWithId(id, "alias-item", text("name", alias))
	}
	return inline("alias", keyword("("),
		inline("alias-list", join(alias_contents, keyword(","))),
		keyword(")"))
}

func funcTypeParams(params Html, omit_by_default bool) Html {
	if params == Html("") {
		return Html("")
	}
	if omit_by_default {
		return inline("func-type-params omit", keyword("generic"), params)
	} else {
		return inline("func-type-params", keyword("generic"), params)
	}
}

func funcImplicit (
	types  ([] checker.Type),
	params ([] checker.TypeParam),
	mod    string,
) Html {
	if len(types) == 0 {
		return Html("")
	}
	var type_contents = make([] Html, len(types))
	for i, t := range types {
		type_contents[i] = typeExpr(t, params, mod)
	}
	return inline("implicit",
		keyword("["), inline("types", type_contents...), keyword("]"))
}

func description(content string) Html {
	var lines = strings.Split(content, "\n")
	var c_lines = make([] Html, len(lines))
	for i, line := range lines {
		c_lines[i] = blockWithSpacePreserved("line", escape(line))
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
					keyword("⊂"), typeExpr(super, params, mod))
			} else {
				var sub, has_sub = bounds.Sub[uint(i)]
				if has_sub {
					return inline("bound-sub",
						keyword("⊃"), typeExpr(sub, params, mod))
				}
			}
			return Html("")
		})()
		var c_param = inline("type-param",
			c_default, c_variance, c_name, c_bound)
		contents[i] = c_param
	}
	return inline("type-params",
		keyword("["),
		inline("type-param-list", join(contents, keyword(","))),
		keyword("]"))
}

func typeDef(g *checker.GenericType, reg checker.TypeRegistry, mod string) Html {
	switch d := g.Definition.(type) {
	case *checker.Native:
		return keyword("native")
	case *checker.Boxed:
		var kind = (func() string {
			if d.Opaque    { return "opaque" }
			if d.Protected { return "protected" }
			if d.Implicit  { return "implicit" }
			return ""
		})()
		var c_kind = (func() Html {
			if kind != "" {
				return block("kind", modifier(kind))
			} else {
				return Html("")
			}
		})()
		var c_weak = (func() Html {
			if d.Weak {
				return block("weak", modifier("weak"))
			} else {
				return Html("")
			}
		})()
		switch T := d.InnerType.(type) {
		case *checker.AnonymousType:
			switch R := T.Repr.(type) {
			case checker.Unit:
				if kind == "" {
					return Html("")
				}
			case checker.Bundle:
				return block("boxed",
					c_kind,
					c_weak,
					blockBundle(R, g.FieldInfo, g.Params, mod))
			}
		}
		return block("boxed",
			c_kind,
			c_weak,
			block("inner", typeExpr(d.InnerType, g.Params, mod)))
	case *checker.Enum:
		var cases = make([] Html, len(d.CaseTypes))
		for i, item := range d.CaseTypes {
			cases[i] = blockWithId(item.Name.String(), "api",
				typeDecl(item.Name, reg[item.Name], reg, mod))
		}
		return block("enum",
			modifier("enum"),
			block("cases", cases...))
	default:
		panic("impossible branch")
	}
}

func blockBundle(b checker.Bundle, info (map[string] checker.FieldInfo), params ([] checker.TypeParam), mod string) Html {
	var fields = make([] Html, len(b.Fields))
	for name, f := range b.Fields {
		var t = typeExpr(f.Type, params, mod)
		var desc = (func() Html {
			if info == nil { return Html("") }
			var field_info, exists = info[name]
			if !(exists) { return Html("") }
			return description(field_info.Doc)
		})()
		fields[f.Index] = block("block-field-item",
			block("block-field", text("name", name), keyword(":"), t),
			block("field-desc", desc))
	}
	return block("inner block-bundle", join(fields, Html("")))
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
				elements[i] = inline("element", typeExpr(el, params, mod))
			}
			return inline("tuple",
				keyword("("),
				inline("element-list", join(elements, keyword(","))),
				keyword(")"))
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
			switch I := R.Input.(type) {
			case *checker.AnonymousType:
				switch I.Repr.(type) {
				case checker.Func:
					return inline("func",
						inline("in", keyword("("), in, keyword(")")),
						keyword("→"),
						inline("out", out))
				}
			}
			switch O := R.Output.(type) {
			case *checker.AnonymousType:
				switch O.Repr.(type) {
				case checker.Func:
					return inline("func",
						inline("in", in),
						keyword("→"),
						inline("out",  keyword("("), out, keyword(")")))
				}
			}
			return inline("func",
				inline("in", in),
				keyword("→"),
				inline("out", out))
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

func inlineWithId(id string, class string, content... Html) Html {
	return Html(fmt.Sprintf("<span id=\"%s\" class=\"%s\">%s</span>",
		escape(id), escape(class), join(content, Html(""))))
}

func block(class string, content... Html) Html {
	return Html(fmt.Sprintf("<div class=\"%s\">%s</div>",
		escape(class), join(content, Html(""))))
}

func blockWithSpacePreserved(class string, content... Html) Html {
	return Html(fmt.Sprintf("<div class=\"%s\"><pre>%s</pre></div>",
		escape(class), join(content, Html(""))))
}

func blockWithId(id string, class string, content... Html) Html {
	return Html(fmt.Sprintf("<div id=\"%s\" class=\"%s\">%s</div>",
		escape(id), escape(class), join(content, Html(""))))
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

