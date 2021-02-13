package loader

import (
	"kumachan/compiler/loader/parser/ast"
	"kumachan/stdlib"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"errors"
)


const ServiceTagPrefix = "service-"
const ServiceTemplateFileName = "rpc_service_template.km"

var __ServiceTemplate = (func() ast.Root {
	var wrap = func(err error) error {
		return fmt.Errorf("failed to load rpc service template file: %w", err)
	}
	var path = filepath.Join(stdlib.GetDirectoryPath(), ServiceTemplateFileName)
	var content, err1 = ioutil.ReadFile(path)
	if err1 != nil { panic(wrap(err1)) }
	var source = SourceFile {
		Path:    path,
		Content: content,
	}
	var root, err2 = source.GetAST()
	if err2 != nil { panic(wrap(errors.New(err2.Desc().StringPlain()))) }
	return root
})()

func DecorateServiceModule(root ast.Root, manifest Manifest) (ast.Root, *Error) {
	if manifest.Kind != "service" {
		return root, nil
	}
	var draft = root
	var statements = make([] ast.VariousStatement, 0)
	for _, s := range __ServiceTemplate.Statements {
		statements = append(statements, s)
	}
	for _, s := range root.Statements {
		statements = append(statements, s)
	}
	draft.Statements = statements
	return draft, nil
}

