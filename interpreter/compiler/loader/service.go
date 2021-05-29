package loader

import (
	"os"
	"fmt"
	"errors"
	"io/ioutil"
	"path/filepath"
	"kumachan/standalone/rpc"
	"kumachan/stdlib"
	"kumachan/interpreter/base/parser/ast"
)


const ServiceTagPrefix = "service-"
const ServiceArgumentTag = (ServiceTagPrefix + "argument")
const ServiceMethodTag = (ServiceTagPrefix + "method")
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
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err2.Message().String())
		panic(wrap(errors.New(err2.Desc().StringPlain())))
	}
	return root
})()

func DecorateServiceModule(root ast.Root, manifest Manifest, ctx Context) (ast.Root, ModuleServiceInfo, *Error) {
	if manifest.Kind != ModuleKind_Service {
		return root, ModuleServiceInfo{}, nil
	}
	var throw = func(err_msg string) (ast.Root, ModuleServiceInfo, *Error) {
		return ast.Root{}, ModuleServiceInfo{}, &Error {
			Context:  ctx,
			Concrete: E_InvalidService { Reason: err_msg },
		}
	}
	var has_tag = func(target string, tags ([] ast.Tag)) bool {
		for _, tag := range tags {
			if ast.GetTagContent(tag) == target {
				return true
			}
		}
		return false
	}
	var id = rpc.ServiceIdentifier {
		Vendor:  manifest.Vendor,
		Project: manifest.Project,
		Name:    manifest.Config.Service.Name,
		Version: manifest.Config.Service.Version,
	}
	var methods = make([] ast.DeclFunction, 0)
	var method_occurred = make(map[string] bool)
	var arg_type ast.DeclType
	var arg_type_found bool
	for _, s := range root.Statements {
		switch decl := s.Statement.(type) {
		case ast.DeclFunction:
			if has_tag(ServiceMethodTag, decl.Tags) {
				var name = ast.Id2String(decl.Name)
				if method_occurred[name] {
					return throw(fmt.Sprintf("duplicate method: %s", name))
				}
				method_occurred[name] = true
				methods = append(methods, decl)
				if len(decl.Params) > 0 {
					return throw(fmt.Sprintf("invalid method: %s", name))
				}
			}
		case ast.DeclType:
			if has_tag(ServiceArgumentTag, decl.Tags) {
				var name = ast.Id2String(decl.Name)
				if arg_type_found {
					return throw(fmt.Sprintf("duplicate argument type: %s", name))
				}
				arg_type_found = true
				arg_type = decl
				if len(decl.Params) > 0 {
					return throw(fmt.Sprintf("invalid argument type: %s", name))
				}
			}
		}
	}
	if !(arg_type_found) {
		return throw("argument type not defined")
	}
	var draft = root
	var statements = make([] ast.VariousStatement, 0)
	for _, s := range __ServiceTemplate.Statements {
		switch decl := s.Statement.(type) {
		case ast.DeclFunction:
			var name = ast.Id2String(decl.Name)
			if name == stdlib.ServiceCreateFunction {
				decl.Body = ast.VariousBody {
					Node: decl.Name.Node,
					Body: ast.ServiceCreateFuncBody {},
				}
				s.Statement = decl
			}
		case ast.DeclConst:
			var name = ast.Id2String(decl.Name)
			if name == stdlib.ServiceIdentifierConst {
				decl.Value = ast.VariousConstValue {
					Node:       decl.Name.Node,
					ConstValue: ast.PredefinedValue {
						Value: id,
					},
				}
				s.Statement = decl
			}
		case ast.DeclType:
			var name = ast.Id2String(decl.Name)
			if name == stdlib.ServiceArgumentType {
				decl.TypeDef.TypeDef = ast.BoxedType {
					Node:  decl.TypeDef.Node,
					Inner: ast.VariousType {
						Node: decl.TypeDef.Node,
						Type: ast.TypeRef {
							Node:     decl.TypeDef.Node,
							Id:       arg_type.Name,
							TypeArgs: [] ast.VariousType {},
						},
					},
				}
				s.Statement = decl
			} else if name == stdlib.ServiceMethodsType {
				var fields = make([] ast.Field, len(methods))
				for i, method := range methods {
					var field = ast.Field {
						Node: method.Name.Node,
						Name: method.Name,
						Type: ast.VariousType {
							Node: method.Repr.Node,
							Type: ast.TypeLiteral {
								Node: method.Repr.Node,
								Repr: ast.VariousRepr {
									Node: method.Repr.Node,
									Repr: method.Repr,
								},
							},
						},
					}
					fields[i] = field
				}
				decl.TypeDef.TypeDef = ast.ImplicitType {
					Node: decl.TypeDef.Node,
					Repr: ast.ReprRecord {
						Node:   decl.TypeDef.Node,
						Fields: fields,
					},
				}
				s.Statement = decl
			}
		}
		statements = append(statements, s)
	}
	for _, s := range root.Statements {
		switch decl := s.Statement.(type) {
		case ast.DeclFunction:
			if has_tag(ServiceMethodTag, decl.Tags) {
				var name = ast.Id2String(decl.Name)
				decl.Body = ast.VariousBody {
					Node: decl.Name.Node,
					Body: ast.ServiceMethodFuncBody { Name: name },
				}
				s.Statement = decl
			}
		}
		statements = append(statements, s)
	}
	draft.Statements = statements
	var method_names = make([] string, len(methods))
	for i, decl := range methods { method_names[i] = ast.Id2String(decl.Name) }
	var arg_type_name = ast.Id2String(arg_type.Name)
	var info = ModuleServiceInfo {
		IsService:          true,
		ServiceIdentifier:  id,
		ServiceArgTypeName: arg_type_name,
		ServiceMethodNames: method_names,
	}
	return draft, info, nil
}

