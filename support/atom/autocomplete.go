package atom

import (
	"os"
	"fmt"
	"sort"
	"strings"
	"path/filepath"
	"kumachan/interpreter/lang/textual/ast"
	"kumachan/interpreter/lang/textual/syntax"
	"kumachan/interpreter/compiler/loader"
	"kumachan/stdlib"
)


var IdentifierRegexp = syntax.GetIdentifierRegexp()
var Keywords = syntax.GetKeywordList()

type AutoCompleteRequest struct {
	PrecedingText  string     `json:"precedingText"`
	LocalBindings  [] string  `json:"localBindings"`
	CurrentPath    string     `json:"currentPath"`
}

type AutoCompleteResponse struct {
	Suggestions  [] AutoCompleteSuggestion   `json:"suggestions"`
}

type AutoCompleteSuggestion struct {
	Text     string   `json:"text"`
	Replace  string   `json:"replacementPrefix"`
	Type     string   `json:"type"`
	Display  string   `json:"displayText,omitempty"`
}

func AutoComplete(req AutoCompleteRequest, ctx LangServerContext) AutoCompleteResponse {
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
	var input, input_mod = get_search_text()
	if input == "" && input_mod == "" {
		return AutoCompleteResponse {}
	}
	var quick_check = func(id ast.Identifier) bool {
		if !(len(id.Name) > 0) { panic("something went wrong") }
		if len(input) > 0 {
			var first_char = id.Name[0]
			if first_char < 128 &&
				first_char != rune(input[0]) &&
				(first_char + ('a' - 'A')) != rune(input[0]) {
				return false
			}
		}
		return true
	}
	var get_name_with_mod = func(name string, mod_prefix string) string {
		if mod_prefix == "" {
			return name
		} else {
			return fmt.Sprintf("%s::%s", mod_prefix, name)
		}
	}
	var suggestions = make([] AutoCompleteSuggestion, 0)
	if len(input) > 0 && input_mod == "" {
		for _, binding := range req.LocalBindings {
			if strings.HasPrefix(binding, input) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    binding,
					Replace: input,
					Type:    "variable",
				})
			}
		}
	}
	var suggested_function_names = make(map[string] bool)
	var process_statement func(ast.Statement, string)
	process_statement = func(stmt ast.Statement, mod string) {
		switch s := stmt.(type) {
		case ast.DeclType:
			switch v := s.TypeDef.TypeDef.(type) {
			case ast.EnumType:
				for _, case_decl := range v.Cases {
					process_statement(case_decl, mod)
				}
			}
			if !(input_mod == mod) {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = ast.Id2String(s.Name)
			var name_lower = strings.ToLower(name)
			if strings.HasPrefix(name, input) || strings.HasPrefix(name_lower, input) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    name,
					Replace: input,
					Type:    "type",
					Display: get_name_with_mod(name, mod),
				})
			}
		case ast.DeclConst:
			if (mod != "" && !(s.Public)) {
				return
			}
			if input_mod != mod {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = ast.Id2String(s.Name)
			var name_lower = strings.ToLower(name)
			if strings.HasPrefix(name, input) || strings.HasPrefix(name_lower, input) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    name,
					Replace: input,
					Type:    "constant",
					Display: get_name_with_mod(name, mod),
				})
			}
		case ast.DeclFunction:
			if (mod != "" && !(s.Public)) {
				return
			}
			if (input_mod != "" && input_mod != mod) {
				return
			}
			if !(quick_check(s.Name)) {
				return
			}
			var name = ast.Id2String(s.Name)
			if strings.HasPrefix(name, input) {
				if !(suggested_function_names[name]) {
					suggestions = append(suggestions, AutoCompleteSuggestion {
						Text:    name,
						Replace: input,
						Type:    "function",
					})
					suggested_function_names[name] = true
				}
			}
		}
	}
	if (input_mod == "" && len(input) >= 2) || input_mod != "" {
		var dir = filepath.Dir(req.CurrentPath)
		var manifest_path = filepath.Join(dir, loader.ManifestFileName)
		var mod_path string
		var mf, err_mf = os.Open(manifest_path)
		if err_mf != nil {
			mod_path = req.CurrentPath
		} else {
			mod_path = dir
			_ = mf.Close()
		}
		var mod, idx, _, err = loader.LoadEntry(mod_path)
		if err != nil { goto keywords }
		for _, item := range mod.AST.Statements {
			process_statement(item.Statement, "")
		}
		for mod_prefix, imp := range mod.ImpMap {
			for _, item := range imp.AST.Statements {
				process_statement(item.Statement, mod_prefix)
			}
		}
		if input_mod == "" {
			for mod_prefix, _ := range mod.ImpMap {
				if strings.HasPrefix(mod_prefix, input) {
					suggestions = append(suggestions, AutoCompleteSuggestion {
						Text:    mod_prefix,
						Replace: input,
						Type:    "import",
					})
				}
			}
			var process_core_statement func(stmt ast.Statement)
			process_core_statement = func(stmt ast.Statement) {
				switch s := stmt.(type) {
				case ast.DeclConst:
					if !(s.Public) { return }
					var name = ast.Id2String(s.Name)
					if strings.HasPrefix(name, input) {
						suggestions = append(suggestions, AutoCompleteSuggestion {
							Text:    name,
							Replace: input,
							Type:    "constant",
						})
					}
				case ast.DeclType:
					switch v := s.TypeDef.TypeDef.(type) {
					case ast.EnumType:
						for _, case_decl := range v.Cases {
							process_core_statement(case_decl)
						}
					}
					var name = ast.Id2String(s.Name)
					if strings.HasPrefix(name, input) {
						suggestions = append(suggestions, AutoCompleteSuggestion{
							Text:    name,
							Replace: input,
							Type:    "type",
						})
					}
				}
			}
			var core_mod = idx[stdlib.Mod_core]
			for _, stmt := range core_mod.AST.Statements {
				process_core_statement(stmt.Statement)
			}
		}
	}
	keywords:
	if len(input) > 0 && input_mod == "" {
		for _, kw := range Keywords {
			if len(kw) > 1 && ('a' <= kw[0] && kw[0] <= 'z') &&
				strings.HasPrefix(kw, input) {
				suggestions = append(suggestions, AutoCompleteSuggestion {
					Text:    kw,
					Replace: input,
					Type:    "keyword",
				})
			}
		}
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		var a = suggestions[i]
		var b = suggestions[j]
		if a.Type == b.Type {
			return a.Text < b.Text
		} else if a.Type == "keyword" {
			return false
		} else if b.Type == "keyword" {
			return true
		} else if a.Type == "function" {
			return false
		} else if b.Type == "function" {
			return true
		} else {
			return a.Text < b.Text
		}
	})
	return AutoCompleteResponse { Suggestions: suggestions }
}

