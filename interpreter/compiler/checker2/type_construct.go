package checker2

import (
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

type RawType struct {
	Type  typsys.Type
}

func newSpecialType(which string) (RawType, bool) {
	switch which {
	case typsys.TypeNameUnknown:
		return RawType { Type: &typsys.UnknownType {} }, true
	case typsys.TypeNameUnit:
		return RawType { Type: typsys.UnitType {} }, true
	case typsys.TypeNameTop:
		return RawType { Type: typsys.TopType {} }, true
	case typsys.TypeNameBottom:
		return RawType { Type: typsys.BottomType {} }, true
	default:
		return RawType {}, false
	}
}

func newParameterType(which string, params ([] typsys.Parameter)) (RawType, bool) {
	for i := range params {
		var p = &(params[i])
		if which == p.Name {
			return RawType { typsys.ParameterType { Parameter: p } }, true
		}
	}
	return RawType {}, false
}

func newType(t ast.VariousType, ctx TypeConsContext) (RawType, *source.Error) {
	switch T := t.Type.(type) {
	case ast.TypeRef:
		var n = typeNameFromTypeRef(T, ctx.Module)
		if n.ModuleName == "" {
			var item_name = n.ItemName
			var special, is_special = newSpecialType(item_name)
			if is_special {
				var num_args = uint(len(T.TypeArgs))
				if num_args > 0 {
					return RawType {}, source.MakeError(T.Location, E_TypeWrongParameterQuantity{
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
			return RawType {}, source.MakeError(def.Location, E_TypeNotFound {
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
		var num_args = uint(len(T.TypeArgs))
		if !(least_arity <= num_args && num_args <= arity) {
			return RawType {}, source.MakeError(T.Location, E_TypeWrongParameterQuantity {
				Which: n_desc,
				Given: num_args,
				Least: least_arity,
				Total: arity,
			})
		}
		var args = make([] typsys.Type, arity)
		for i := uint(0); i < arity; i += 1 {
			var arg typsys.Type
			if i < num_args {
				var raw, err = newType(T.TypeArgs[i], ctx)
				if err != nil { return RawType {}, err }
				arg = raw.Type
			} else {
				arg = def.Parameters[i].Default
			}
			if arg == nil { panic("something went wrong") }
		}
		var ret = &typsys.NestedType {
			Content: typsys.Ref {
				Def:  def.TypeDef,
				Args: args,
			},
		}
		return RawType { Type: ret }, nil
	case ast.TypeLiteral:
		return newTypeFromRepr(T.Repr.Repr, ctx)
	default:
		panic("impossible branch")
	}
}

func newTypeFromRepr(r ast.Repr, ctx TypeConsContext) (RawType, *source.Error) {
	switch R := r.(type) {
	case ast.ReprTuple:
		var num_elements = uint(len(R.Elements))
		if num_elements == 0 {
			return RawType { Type: typsys.UnitType {} }, nil
		} else {
			var elements = make([] typsys.Type, num_elements)
			for i, t := range R.Elements {
				var raw, err = newType(t, ctx)
				if err != nil { return RawType {}, err }
				elements[i] = raw.Type
			}
			var tuple = typsys.Tuple { Elements: elements }
			var ret = &typsys.NestedType { Content: tuple }
			return RawType { Type: ret }, nil
		}
	case ast.ReprRecord:
		var fields = make([] typsys.Field, len(R.Fields))
		var index_map = make(map[string] uint)
		for i, field := range R.Fields {
			var index = uint(i)
			var field_name = ast.Id2String(field.Name)
			var _, exists = index_map[field_name]
			if exists {
				return RawType {}, source.MakeError(field.Name.Location,
					E_TypeDuplicateField { Which: field_name })
			}
			index_map[field_name] = index
			var raw, err = newType(field.Type, ctx)
			if err != nil { return RawType {}, err }
			fields[i] = typsys.Field {
				Attr: attr.FieldAttr {
					Attrs: attr.Attrs {
						Location: field.Location,
						Section:  nil,
						Doc:      ast.GetDocContent(field.Docs),
					},
				},
				Name: field_name,
				Type: raw.Type,
			}
		}
		var record = typsys.Record {
			FieldIndexMap: index_map,
			Fields:        fields,
		}
		var ret = &typsys.NestedType { Content: record }
		return RawType { Type: ret }, nil
	case ast.ReprFunc:
		var input, err1 = newType(R.Input, ctx)
		if err1 != nil { return RawType {}, err1 }
		var output, err2 = newType(R.Output, ctx)
		if err2 != nil { return RawType {}, err2 }
		var lambda = typsys.Lambda {
			Input:  input.Type,
			Output: output.Type,
		}
		var ret = &typsys.NestedType { Content: lambda }
		return RawType { Type: ret }, nil
	default:
		panic("impossible branch")
	}
}


