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


type TypeRegistry (map[name.TypeName] *typsys.TypeDef)

type typeDraftPair struct {
	draft  *typsys.TypeDef
	decl   *ast.DeclType
}
type _typeDraftMap ([] typeDraftPair)
type typeDraftMap *_typeDraftMap

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
		return name.TypeName{}, false
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
	var ref_mod = ast.Id2String(ref.Module)
	var ref_item = ast.Id2String(ref.Item)
	if ref_mod == "" {
		var _, is_core_type = coreTypes[ref_item]
		if is_core_type {
			return name.MakeTypeName(stdlib.Mod_core, ref_item)
		} else {
			return name.MakeTypeName(mod.Name, ref_item)
		}
	} else {
		var imported, found = mod.ImpMap[ref_mod]
		if found {
			return name.MakeTypeName(imported.Name, ref_item)
		} else {
			return name.MakeTypeName("", ref_item)
		}
	}
}

func TypeNameListFrom(ref_list ([] ast.TypeRef), mod *loader.Module) ([] name.TypeName) {
	var list = make([] name.TypeName, len(ref_list))
	for i, ref := range ref_list {
		list[i] = TypeNameFromTypeRef(ref, mod)
	}
	return list
}

func collectTypes(mod *loader.Module, reg TypeRegistry) *source.Error {
	var err = registerTypes(mod, reg)
	if err != nil { return err }
	// TODO: alias handling
	// TODO
}

func registerTypes(mod *loader.Module, reg TypeRegistry) *source.Error {
	var dm = typeDraftMap(new(_typeDraftMap))
	var sb SectionBuffer
	for _, stmt := range mod.AST.Statements {
		var title, is_title = stmt.Statement.(ast.Title)
		if is_title {
			sb.SetFrom(title)
		}
		var decl, is_type_decl = stmt.Statement.(ast.DeclType)
		if !(is_type_decl) { continue }
		var _, err = registerType(&decl, dm, &sb, mod, reg, (typsys.CaseInfo {}))
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
	dm    typeDraftMap,
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
	var draft = new(typsys.TypeDef)
	reg[type_name] = draft
	*dm = append(*dm, typeDraftPair { draft: draft, decl: decl })
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
					Default:  nil, // TODO
					Variance: v,
					Bound:    typsys.Bound {}, // TODO
				}
			}
			return params, nil
		}
	})()
	if params_err != nil { return nil, params_err }
	*draft = typsys.TypeDef {
		TypeAttrs:  attrs,
		Name:       type_name,
		Implements: nil, // TODO
		Parameters: params,
		Content:    nil, // TODO
		CaseInfo:   ci,
	}
	var enum, is_enum = decl.TypeDef.TypeDef.(ast.EnumType)
	var case_defs = make([] *typsys.TypeDef, len(enum.Cases))
	if is_enum {
		for i, c := range enum.Cases {
			var ct, err = registerType(&c, dm, sb, mod, reg, typsys.CaseInfo {
				Enum:      draft,
				CaseIndex: uint(i),
			})
			case_defs[i] = ct
			if err != nil { return nil, err }
		}
	}
	return draft, nil
}


