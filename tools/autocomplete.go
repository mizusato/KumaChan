package tools

import (
	"kumachan/parser/syntax"
	"strings"
)


var IdentifierRegexp = syntax.GetIdentifierRegexp()
var Keywords = syntax.GetKeywordList()

type AutoCompleteRequest struct {
	PrecedingText  string           `json:"precedingText"`
	CurrentPath    string           `json:"currentPath"`
	SiblingDirty   [] DirtyBuffer   `json:"siblingDirtyBufferContents"`
}

type AutoCompleteResponse struct {
	Suggestions  [] AutoCompleteSuggestion   `json:"suggestions"`
}

type AutoCompleteSuggestion struct {
	Text  string   `json:"text"`
	Type  string   `json:"type"`
}

func AutoComplete(req AutoCompleteRequest) AutoCompleteResponse {
	var raw_text = req.PrecedingText
	var ranges = IdentifierRegexp.FindAllStringIndex(raw_text, -1)
	if len(ranges) == 0 {
		return AutoCompleteResponse {}
	}
	var last = ranges[len(ranges)-1]
	var lo = last[0]
	var hi = last[1]
	if hi != len(raw_text) {
		return AutoCompleteResponse {}
	}
	var text = raw_text[lo:hi]
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

