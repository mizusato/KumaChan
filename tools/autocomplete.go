package tools

import (
	"fmt"
	"strings"
	"kumachan/parser/syntax"
	"kumachan/loader"
	"kumachan/parser/ast"
	"kumachan/stdlib"
)


var IdentifierRegexp = syntax.GetIdentifierRegexp()
var Keywords = syntax.GetKeywordList()

type AutoCompleteRequest struct {
	PrecedingText  string   `json:"precedingText"`
	CurrentPath    string   `json:"currentPath"`
}

type AutoCompleteResponse struct {
	Suggestions  [] AutoCompleteSuggestion   `json:"suggestions"`
}

type AutoCompleteSuggestion struct {
	Text     string   `json:"text"`
	Type     string   `json:"type"`
	Display  string   `json:"displayText,omitempty"`
}

func AutoComplete(req AutoCompleteRequest, ctx ServerContext) AutoCompleteResponse {
	const double_colon = "::"
	var get_search_text = func() (string, string) {
		var raw_text = req.PrecedingText
		var ranges = IdentifierRegexp.FindAllStringIndex(raw_text, -1)
		if len(ranges) == 0 {
			return "", ""
		}
		var last = ranges[len(ranges)-1]
		var lo = last[0]
		var hi = last[1]
		if hi != len(raw_text) {
			if strings.HasSuffix(raw_text, double_colon) &&
				hi == (len(raw_text) - len(double_colon)) {
				return "", raw_text[lo:hi]
			} else {
				return "", ""
			}
		}
		var text = raw_text[lo:hi]
		var ldc = len(double_colon)
		if len(ranges) >= ldc {
			var preceding = ranges[len(ranges)-ldc]
			var p_lo = preceding[0]
			var p_hi = preceding[1]
			if (p_hi + ldc) == lo && raw_text[p_hi:lo] == double_colon {
				var text_mod = raw_text[p_lo:p_hi]
				return text, text_mod
			} else {
				return text, ""
			}
		} else {
			return text, ""
		}
	}
	var text, text_mod = get_search_text()
	if text == "" && text_mod == "" {
		return AutoCompleteResponse {}
	}
	var quick_check = func(id ast.Identifier) bool {
		if !(len(id.Name) > 0) { panic("something went wrong") }
		if len(text) > 0 {
			var first_char = id.Name[0]
			if first_char < 128 &&
				first_char != rune(text[0]) &&
				(first_char + ('a' - 'A')) != rune(text[0]) {
				return false
			}
		}
		return true
	}
	var suggestions = make([] AutoCompleteSuggestion, 0)
	var suggested_function_names = make(map[string] bool)
	var process_statement func(ast.Statement, string)
	var get_name_with_mod = func(name string, mod_prefix string) string {
		if mod_prefix == "" {
			return name
		} else {
			return fmt.Sprintf("%s::%s", mod_prefix, name)
		}
	}
	process_statement = func(stmt ast.Statement, mod_prefix string) {
		switch s := stmt.(type) {
		case ast.DeclType:
			switch v := s.TypeValue.TypeValue.(type) {
			case ast.UnionType:
				for _, decl := range v.Cases {
					process_statement(decl, mod_prefix)
				}
			}
			if !(text_mod == mod_prefix) {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = loader.Id2String(s.Name)
			var name_lower = strings.ToLower(name)
			if strings.HasPrefix(name, text) || strings.HasPrefix(name_lower, text) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    name,
					Type:    "type",
					Display: get_name_with_mod(name, mod_prefix),
				})
			}
		case ast.DeclConst:
			if !(text_mod == mod_prefix) {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = loader.Id2String(s.Name)
			var name_lower = strings.ToLower(name)
			if strings.HasPrefix(name, text) || strings.HasPrefix(name_lower, text) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    name,
					Type:    "constant",
					Display: get_name_with_mod(name, mod_prefix),
				})
			}
		case ast.DeclFunction:
			if !(text_mod == "") {
				return
			}
			if !(mod_prefix == "" || s.Public) {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = loader.Id2String(s.Name)
			if strings.HasPrefix(name, text) {
				if !(suggested_function_names[name]) {
					suggestions = append(suggestions, AutoCompleteSuggestion {
						Text: name,
						Type: "function",
					})
					suggested_function_names[name] = true
				}
			}
		}
	}
	if (text_mod == "" && len(text) >= 2) || text_mod != "" {
		var mod, idx, err =
			loader.LoadEntryWithCache(req.CurrentPath, ctx.LoaderCache)
		if err != nil { goto keywords }
		for _, item := range mod.Node.Statements {
			process_statement(item.Statement, "")
		}
		for mod_prefix, imp := range mod.ImpMap {
			for _, item := range imp.Node.Statements {
				process_statement(item.Statement, mod_prefix)
			}
		}
		if text_mod == "" {
			for mod_prefix, _ := range mod.ImpMap {
				if strings.HasPrefix(mod_prefix, text) {
					suggestions = append(suggestions, AutoCompleteSuggestion {
						Text: mod_prefix,
						Type: "import",
					})
				}
			}
			var core_mod = idx[stdlib.Core]
			for _, stmt := range core_mod.Node.Statements {
				switch s := stmt.Statement.(type) {
				case ast.DeclConst:
					var name = loader.Id2String(s.Name)
					if strings.HasPrefix(name, text) {
						suggestions = append(suggestions, AutoCompleteSuggestion {
							Text: name,
							Type: "constant",
						})
					}
				case ast.DeclType:
					var name = loader.Id2String(s.Name)
					if strings.HasPrefix(name, text) {
						suggestions = append(suggestions, AutoCompleteSuggestion{
							Text: name,
							Type: "type",
						})
					}
				}
			}
		}
	}
	keywords:
	if len(text) > 0 && text_mod == "" {
		for _, kw := range Keywords {
			if strings.HasPrefix(kw, text) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text: kw,
					Type: "keyword",
				})
			}
		}
	}
	return AutoCompleteResponse { Suggestions: suggestions }
}

