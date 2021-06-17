package checker2

import (
	"strings"
	"encoding/json"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
	"kumachan/stdlib"
)


type TypeRegistry (map[name.TypeName] TypeDefItem)

type TypeDefItem struct {
	*typsys.TypeDef
	AstNode  *ast.DeclType
}

var coreTypes = (func() (map[string] struct{}) {
	var set = make(map[string] struct{})
	var list = stdlib.CoreTypeNames()
	for _, name := range list {
		set[name] = struct{}{}
	}
	return set
})()

func TypeNameFromIdentifier(id ast.Identifier, mod *loader.Module) (name.TypeName, bool) {
	var n = ast.Id2String(id)
	if CheckTypeName(n) {
		return name.MakeTypeName(mod.Name, n), true
	} else {
		return name.TypeName {}, false
	}
}

func ParameterNameVarianceFromIdentifier(id ast.Identifier) (string, typsys.Variance, bool) {
	var n = ast.Id2String(id)
	var v = typsys.Invariant
	if strings.HasPrefix(n, CovariantPrefix) {
		n = strings.TrimPrefix(n, CovariantPrefix)
		v = typsys.Covariant
	} else if strings.HasPrefix(n, ContravariantPrefix) {
		n = strings.TrimPrefix(n, ContravariantPrefix)
		v = typsys.Contravariant
	}
	if CheckTypeName(n) {
		return n, v, true
	} else {
		return "", -1, false
	}
}

func TypeNameFromTypeRef(ref ast.TypeRef, mod *loader.Module) name.TypeName {
	return name.TypeName { Name: NameFrom(ref.Module, ref.Item, mod) }
}

func TypeNameListFrom(ref_list ([] ast.TypeRef), mod *loader.Module) ([] name.TypeName) {
	var list = make([] name.TypeName, len(ref_list))
	for i, ref := range ref_list {
		list[i] = TypeNameFromTypeRef(ref, mod)
	}
	return list
}

var __DefaultInit, defaultWrite = (func() (typsys.Type, func(typsys.Type)(typsys.Type)) {
	return nil, func(t typsys.Type) typsys.Type { return t }
})()
var __BoundInit, boundWrite = (func() (typsys.Bound, func(typsys.Bound)(typsys.Bound)) {
	return typsys.Bound {}, func(b typsys.Bound) typsys.Bound { return b }
})()
var __ImplInit, implWrite = (func() (([] typsys.DispatchTable), func([] typsys.DispatchTable)([] typsys.DispatchTable)) {
	return nil, func(d ([] typsys.DispatchTable)) ([] typsys.DispatchTable) { return d }
})()
var __ContentInit, contentWrite = (func() (typsys.TypeDefContent, func(typsys.TypeDefContent)(typsys.TypeDefContent)) {
	return nil, func(c typsys.TypeDefContent) typsys.TypeDefContent { return c }
})()

func collectTypes(entry *loader.Module, al AliasRegistry) (TypeRegistry, *source.Error) {
	var reg = make(TypeRegistry)
	var err = registerTypes(entry, reg)
	if err != nil { return nil, err }
	for type_name, def := range reg {
		var _, conflict = al[type_name.Name]
		if conflict {
			return nil, source.MakeError(def.Location, E_TypeConflictWithAlias {
				Which: type_name.String(),
			})
		}
	}
	// TODO
}

func registerTypes(mod *loader.Module, reg TypeRegistry) *source.Error {
	var sb SectionBuffer
	for _, stmt := range mod.AST.Statements {
		var title, is_title = stmt.Statement.(ast.Title)
		if is_title {
			sb.SetFrom(title)
		}
		var decl, is_type_decl = stmt.Statement.(ast.DeclType)
		if !(is_type_decl) { continue }
		var _, err = registerType(&decl, &sb, mod, reg, (typsys.CaseInfo {}))
		if err != nil { return err }
	}
	for _, imported := range mod.ImpMap {
		var err = registerTypes(imported, reg)
		if err != nil { return err }
	}
	return nil
}

func registerType (
	decl  *ast.DeclType,
	sb    *SectionBuffer,
	mod   *loader.Module,
	reg   TypeRegistry,
	ci    typsys.CaseInfo,
) (*typsys.TypeDef, *source.Error) {
	var type_name, type_name_ok = TypeNameFromIdentifier(decl.Name, mod)
	if !(type_name_ok) {
		return nil, source.MakeError(decl.Name.Location, E_InvalidTypeName {
			Name: ast.Id2String(decl.Name),
		})
	}
	var def = new(typsys.TypeDef)
	reg[type_name] = TypeDefItem {
		TypeDef: def,
		AstNode: decl,
	}
	var loc = decl.Location
	var doc = ast.GetDocContent(decl.Docs)
	var section = sb.GetFrom(loc)
	var meta attr.TypeMetadata
	var meta_text = ast.GetMetadataContent(decl.Meta)
	var meta_err = json.Unmarshal(([] byte)(meta_text), &meta)
	if meta_err != nil {
		return nil, source.MakeError(loc, E_InvalidMetadata {
			Reason: meta_err.Error(),
		})
	}
	var attrs = attr.TypeAttrs {
		Attrs: attr.Attrs {
			Location: loc,
			Section:  section,
			Doc:      doc,
		},
		Metadata: meta,
	}
	var params, params_err = (func() ([] typsys.Parameter, *source.Error) {
		if ci.Enum != nil {
			if len(decl.Params) > 0 {
				return nil, source.MakeError(loc, E_TypeParametersOnCaseType {})
			}
			return ci.Enum.Parameters, nil
		} else {
			var params = make([] typsys.Parameter, len(decl.Params))
			for i, p := range decl.Params {
				var n, v, ok = ParameterNameVarianceFromIdentifier(p.Name)
				if !(ok) {
					return nil, source.MakeError(p.Name.Location, E_InvalidTypeName {
						Name: ast.Id2String(p.Name),
					})
				}
				params[i] = typsys.Parameter {
					Name:     n,
					Default:  __DefaultInit,
					Variance: v,
					Bound:    __BoundInit,
				}
			}
			return params, nil
		}
	})()
	if params_err != nil { return nil, params_err }
	*def = typsys.TypeDef {
		TypeAttrs:  attrs,
		Name:       type_name,
		Implements: __ImplInit,
		Parameters: params,
		Content:    __ContentInit,
		CaseInfo:   ci,
	}
	var enum, is_enum = decl.TypeDef.TypeDef.(ast.EnumType)
	var case_defs = make([] *typsys.TypeDef, len(enum.Cases))
	if is_enum {
		for i, c := range enum.Cases {
			var ct, err = registerType(&c, sb, mod, reg, typsys.CaseInfo {
				Enum:      def,
				CaseIndex: uint(i),
			})
			case_defs[i] = ct
			if err != nil { return nil, err }
		}
	}
	return def, nil
}

type TypeConsContext struct {
	Module    *loader.Module
	TypeReg   TypeRegistry
	AliasReg  AliasRegistry
}
func (ctx TypeConsContext) ResolveName(n name.TypeName) (TypeDefItem, string, bool) {
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
func newSpecialType(item_name string) (RawType, bool) {
	switch item_name {
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
func newType(t ast.VariousType, ctx TypeConsContext) (RawType, *source.Error) {
	switch T := t.Type.(type) {
	case ast.TypeRef:
		var n = TypeNameFromTypeRef(T, ctx.Module)
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
		}
		var def, n_desc, exists = ctx.ResolveName(n)
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
			if arg == nil {panic("something went wrong") }
		}
		var ret = &typsys.NestedType {
			Content: typsys.Ref {
				Def:  def.TypeDef,
				Args: args,
			},
		}
		return RawType { Type: ret }, nil
	case ast.TypeLiteral:
		switch R := T.Repr.Repr.(type) {
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
	default:
		panic("impossible branch")
	}
}


