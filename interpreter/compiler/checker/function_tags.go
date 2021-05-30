package checker

import (
	"fmt"
	"strings"
	"reflect"
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/parser/ast"
	"kumachan/interpreter/parser/syntax"
)


type FunctionTags struct {
	AliasList  [] string
	FunctionServiceConfig
	FunctionCompilationFlags
}
type FunctionServiceConfig struct {
	IsServiceMethod  bool
}
type FunctionCompilationFlags struct {
	ExplicitCall  bool   `flag:"explicit-call"`
}

type FunctionTagParsingError struct {
	Tag   ast.Tag
	Info  string
}

func ParseFunctionTags(ast_tags ([] ast.Tag)) (FunctionTags, *FunctionTagParsingError) {
	var tags = FunctionTags { AliasList: [] string {} }
	var occurred_alias = make(map[string] bool)
	var flags_rv = reflect.ValueOf(&(tags.FunctionCompilationFlags))
	var flags_t = flags_rv.Elem().Type()
	outer: for _, ast_tag := range ast_tags {
		var raw = ast.GetTagContent(ast_tag)
		if strings.HasPrefix(raw, loader.ServiceTagPrefix) {
			if raw == loader.ServiceMethodTag {
				tags.IsServiceMethod = true
				continue
			}
		}
		for i := 0; i < flags_t.NumField(); i += 1 {
			var flag_name = flags_t.Field(i).Tag.Get("flag")
			if flag_name == "" { panic("something went wrong") }
			if raw == flag_name {
				flags_rv.Elem().Field(i).SetBool(true)
				continue outer
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
