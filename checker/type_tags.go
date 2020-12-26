package checker

import (
	"fmt"
	"strings"
	"kumachan/parser/ast"
	"kumachan/parser/syntax"
)


type TypeTags struct {
	DataConfig  TypeDataConfig
}

type TypeDataConfig struct {
	Name     string
	Version  string
}

type TypeTagParsingError struct {
	Tag   ast.Tag
	Info  string
}

func ParseTypeTags(ast_tags ([] ast.Tag)) (TypeTags, *TypeTagParsingError) {
	var tags TypeTags
	for _, ast_tag := range ast_tags {
		var raw = string(ast_tag.RawContent)
		raw = strings.TrimRight(raw, "\r")
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.Trim(raw, " ")
		var t = strings.Split(raw, ":")
		if len(t) != 2 {
			return TypeTags{}, &TypeTagParsingError {
				Tag:  ast_tag,
				Info: "wrong format",
			}
		}
		var kind = t[0]
		if kind == "data" {
			var t = strings.Split(t[1], ",")
			for _, item := range t {
				item = strings.Trim(item, " ")
				var t = strings.Split(item, "=")
				if len(t) == 2 {
					var key = t[0]
					var val = t[1]
					if key == "name" {
						if syntax.GetIdentifierFullRegexp().MatchString(val) {
							tags.DataConfig.Name = val
						} else {
							return TypeTags{}, &TypeTagParsingError {
								Tag:  ast_tag,
								Info: fmt.Sprintf("invalid value for item 'name': %s", val),
							}
						}
					} else if key == "ver" {
						if syntax.GetIdentifierFullRegexp().MatchString(val) {
							tags.DataConfig.Version = val
						} else {
							return TypeTags{}, &TypeTagParsingError {
								Tag:  ast_tag,
								Info: fmt.Sprintf("invalid value for item 'ver': %s", val),
							}
						}
					} else {
						return TypeTags{}, &TypeTagParsingError {
							Tag:  ast_tag,
							Info: fmt.Sprintf("unknown data config key: %s", key),
						}
					}
				} else {
					return TypeTags{}, &TypeTagParsingError {
						Tag:  ast_tag,
						Info: fmt.Sprintf("invalid data config item: %s", item),
					}
				}
			}
			if tags.DataConfig.Name == "" {
				return TypeTags{}, &TypeTagParsingError {
					Tag:  ast_tag,
					Info: fmt.Sprintf("invalid data config: 'name' not set"),
				}
			}
			if tags.DataConfig.Version == "" {
				return TypeTags{}, &TypeTagParsingError {
					Tag:  ast_tag,
					Info: fmt.Sprintf("invalid data config: 'ver' not set"),
				}
			}
		} else {
			return TypeTags{}, &TypeTagParsingError {
				Tag:  ast_tag,
				Info: fmt.Sprintf("invalid type tag kind: %s", kind),
			}
		}
	}
	return tags, nil
}