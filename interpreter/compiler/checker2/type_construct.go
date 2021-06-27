package checker2

import (
	"encoding/json"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
)


type TypeConsContext struct {
	Module    *loader.Module
	TypeReg   TypeRegistry
	AliasReg  AliasRegistry
	ParamVec  [] typsys.Parameter
}
func (ctx TypeConsContext) ResolveGlobalName(n name.TypeName) (TypeDef, string, bool) {
	var alias, is_alias = ctx.AliasReg[n.Name]
	if is_alias {
		n = name.TypeName { Name: alias.To }
	}
	var desc = DescribeNameWithPossibleAlias(n.Name, alias.To)
	var def, exists = ctx.TypeReg[n]
	return def, desc, exists
}

func newSpecialType(which string) (typsys.Type, bool) {
	switch which {
	case typsys.TypeNameUnknown:
		return &typsys.UnknownType {}, true
	case typsys.TypeNameUnit:
		return typsys.UnitType {}, true
	case typsys.TypeNameTop:
		return typsys.TopType {}, true
	case typsys.TypeNameBottom:
		return typsys.BottomType {}, true
	default:
		return nil, false
	}
}

func newParameterType(which string, params ([] typsys.Parameter)) (typsys.Type, bool) {
	for i := range params {
		var p = &(params[i])
		if which == p.Name {
			return typsys.ParameterType { Parameter: p }, true
		}
	}
	return nil, false
}

func newType(t ast.VariousType, ctx TypeConsContext) (typsys.Type, *source.Error) {
	switch T := t.Type.(type) {
	case ast.TypeRef:
		var num_args = uint(len(T.TypeArgs))
		if num_args > MaxTypeParameters {
			return nil, source.MakeError(T.Location,
				E_TooManyTypeParameters {})
		}
		var n = name.TypeName { Name: NameFrom(T.Module, T.Item, ctx.Module) }
		if n.ModuleName == "" {
			var item_name = n.ItemName
			var special, is_special = newSpecialType(item_name)
			if is_special {
				var num_args = uint(len(T.TypeArgs))
				if num_args > 0 {
					return nil, source.MakeError(T.Location, E_TypeWrongParameterQuantity{
						Which: item_name,
						Given: num_args,
						Least: 0,
						Total: 0,
					})
				}
				return special, nil
			}
			var param, is_param = newParameterType(item_name, ctx.ParamVec)
			if is_param {
				return param, nil
			}
		}
		var def, n_desc, exists = ctx.ResolveGlobalName(n)
		if !(exists) {
			return nil, source.MakeError(T.Location, E_TypeNotFound {
				Which: n_desc,
			})
		}
		var arity = uint(len(def.Parameters))
		var least_arity = arity
		if arity > 0 {
			for i := (arity - 1); i >= 0; i -= 1 {
				if def.Parameters[i].Default != nil {
					least_arity -= 1
				} else {
					break
				}
			}
		}
		if !(least_arity <= num_args && num_args <= arity) {
			return nil, source.MakeError(T.Location,
				E_TypeWrongParameterQuantity {
					Which: n_desc,
					Given: num_args,
					Least: least_arity,
					Total: arity,
				})
		}
		var args = make([] typsys.Type, arity)
		var err = def.ForEachParameter(func(i uint, p *typsys.Parameter) *source.Error {
			var arg, err = (func() (typsys.Type, *source.Error) {
				if i < num_args {
					var specified, err = newType(T.TypeArgs[i], ctx)
					if err != nil { return nil, err }
					return specified, nil
				} else {
					return p.Default, nil
				}
			})()
			if err != nil { return err }
			args[i] = arg
			return nil
		})
		if err != nil { return nil, err }
		return &typsys.NestedType {
			Content: typsys.Ref {
				Def:  def.TypeDef,
				Args: args,
			},
		}, nil
	case ast.TypeLiteral:
		return newTypeFromRepr(T.Repr.Repr, ctx)
	default:
		panic("impossible branch")
	}
}

func newTypeFromRepr(r ast.Repr, ctx TypeConsContext) (typsys.Type, *source.Error) {
	switch R := r.(type) {
	case ast.ReprTuple:
		// TODO: quantity limit
		var num_elements = uint(len(R.Elements))
		if num_elements == 0 {
			return typsys.UnitType {}, nil
		} else {
			var elements = make([] typsys.Type, num_elements)
			for i, t := range R.Elements {
				var element, err = newType(t, ctx)
				if err != nil { return nil, err }
				elements[i] = element
			}
			var tuple = typsys.Tuple { Elements: elements }
			return &typsys.NestedType { Content: tuple }, nil
		}
	case ast.ReprRecord:
		// TODO: quantity limit
		var fields = make([] typsys.Field, len(R.Fields))
		var index_map = make(map[string] uint)
		for i, field := range R.Fields {
			var index = uint(i)
			var field_name = ast.Id2String(field.Name)
			var _, exists = index_map[field_name]
			if exists {
				return nil, source.MakeError(field.Name.Location,
					E_TypeDuplicateField { Which: field_name })
			}
			index_map[field_name] = index
			var field_type, err = newType(field.Type, ctx)
			if err != nil { return nil, err }
			var meta attr.FieldMetadata
			var meta_text = ast.GetMetadataContent(field.Meta)
			var meta_err = json.Unmarshal(([] byte)(meta_text), &meta)
			if meta_err != nil {
				return nil, source.MakeError(field.Meta.Location,
					E_InvalidMetadata { Reason: meta_err.Error() })
			}
			fields[i] = typsys.Field {
				Attr: attr.FieldAttrs {
					Attrs: attr.Attrs {
						Location: field.Location,
						Section:  nil,
						Doc:      ast.GetDocContent(field.Docs),
					},
					Metadata: meta,
				},
				Name: field_name,
				Type: field_type,
			}
		}
		var record = typsys.Record {
			FieldIndexMap: index_map,
			Fields:        fields,
		}
		return &typsys.NestedType { Content: record }, nil
	case ast.ReprFunc:
		var input, err1 = newType(R.Input, ctx)
		if err1 != nil { return nil, err1 }
		var output, err2 = newType(R.Output, ctx)
		if err2 != nil { return nil, err2 }
		var lambda = typsys.Lambda {
			Input:  input,
			Output: output,
		}
		return &typsys.NestedType { Content: lambda }, nil
	default:
		panic("impossible branch")
	}
}


