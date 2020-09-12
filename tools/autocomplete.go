package tools

import (
	"strings"
	"kumachan/parser/syntax"
	"kumachan/loader"
	"kumachan/parser/ast"
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
	Text  string   `json:"text"`
	Type  string   `json:"type"`
}

func AutoComplete(req AutoCompleteRequest, ctx ServerContext) AutoCompleteResponse {
	var get_search_text = func() string {
		var raw_text = req.PrecedingText
		var ranges = IdentifierRegexp.FindAllStringIndex(raw_text, -1)
		if len(ranges) == 0 {
			return ""
		}
		var last = ranges[len(ranges)-1]
		var lo = last[0]
		var hi = last[1]
		if hi != len(raw_text) {
			return ""
		}
		return raw_text[lo:hi]
	}
	var text = get_search_text()
	if text == "" {
		return AutoCompleteResponse{}
	}
	var suggestions = make([] AutoCompleteSuggestion, 0)
	var suggested_function_names = make(map[string] bool)
	var process_statement = func(stmt ast.VariousStatement, imported bool) {
		switch s := stmt.Statement.(type) {
		case ast.DeclFunction:
			if imported && !(s.Public) {
				return
			}
			if len(s.Name.Name) == 0 { panic("something went wrong") }
			var first_char = s.Name.Name[0]
			if first_char < 128 && first_char != rune(text[0]) {
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
	if len(text) >= 2 {
		var mod, _, err =
			loader.LoadEntryWithCache(req.CurrentPath, ctx.LoaderCache)
		if err == nil {
			for _, stmt := range mod.Node.Statements {
				process_statement(stmt, false)
			}
			for _, imp := range mod.ImpMap {
				for _, stmt := range imp.Node.Statements {
					process_statement(stmt, true)
				}
			}
		}
	}
	for _, kw := range Keywords {
		if strings.HasPrefix(kw, text) {
			suggestions = append(suggestions, AutoCompleteSuggestion {
				Text: kw,
				Type: "keyword",
			})
		}
	}
	return AutoCompleteResponse { Suggestions: suggestions }
}

