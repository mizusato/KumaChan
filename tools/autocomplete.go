package tools

import (
	"strings"
	"kumachan/parser/syntax"
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
	//var mod, idx, err =
	//	loader.LoadEntryWithCache(req.CurrentPath, ctx.LoaderCache)
	//if err == nil {
	//
	//}
	var suggestions = make([] AutoCompleteSuggestion, 0)
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

