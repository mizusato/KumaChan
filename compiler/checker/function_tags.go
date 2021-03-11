package checker

import (
	"fmt"
	"strings"
	"kumachan/compiler/loader"
	"kumachan/lang/parser/ast"
	"kumachan/lang/parser/syntax"
)


type FunctionTags struct {
	AliasList  [] string
	FunctionServiceConfig
}
type FunctionServiceConfig struct {
	IsServiceMethod  bool
}

type FunctionTagParsingError struct {
	Tag   ast.Tag
	Info  string
}

func ParseFunctionTags(ast_tags ([] ast.Tag)) (FunctionTags, *FunctionTagParsingError) {
	var tags = FunctionTags { AliasList: [] string {} }
	var occurred_alias = make(map[string] bool)
	for _, ast_tag := range ast_tags {
		var raw = ast.GetTagContent(ast_tag)
		if strings.HasPrefix(raw, loader.ServiceTagPrefix) {
			if raw == loader.ServiceMethodTag {
				tags.IsServiceMethod = true
				continue
			}
		}
		var t = strings.Split(raw, ":")
		if len(t) != 2 {
			return FunctionTags{}, &FunctionTagParsingError {
				Tag:  ast_tag,
				Info: "wrong format",
			}
		}
		var kind = t[0]
		if kind == "alias" {
			var t = strings.Split(t[1], ",")
			for _, item := range t {
				item = strings.Trim(item, " ")
				var id_regex = syntax.GetIdentifierFullRegexp()
				if id_regex.MatchString(item) {
					if occurred_alias[item] {
						return FunctionTags{}, &FunctionTagParsingError {
							Tag:  ast_tag,
							Info: fmt.Sprintf("duplicate function alias: %s", item),
						}
					}
					occurred_alias[item] = true
					tags.AliasList = append(tags.AliasList, item)
				} else {
					return FunctionTags{}, &FunctionTagParsingError {
						Tag:  ast_tag,
						Info: fmt.Sprintf("invalid function alias: %s", item),
					}
				}
			}
		} else {
			return FunctionTags{}, &FunctionTagParsingError {
				Tag:  ast_tag,
				Info: fmt.Sprintf("invalid function tag kind: %s", kind),
			}
		}
	}
	return tags, nil
}
