package checker2

import (
	"sort"
	"encoding/json"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
	"kumachan/stdlib"
)


type TypeRegistry (map[name.TypeName] TypeDef)
type TypeList ([] TypeDefWithModule)
type TypeDefWithModule struct {
	TypeDef
	Module  *loader.Module
}
func (l TypeList) Less(i int, j int) bool {
	var u = l[i].Name
	var v = l[j].Name
	if u.ModuleName < v.ModuleName {
		return true
	} else if u.ModuleName == v.ModuleName {
		return (u.ItemName < v.ItemName)
	} else {
		return false
	}
}
func (l TypeList) Len() int {
	return len(l)
}
func (l TypeList) Swap(i int, j int) {
	var I = &(l[i])
	var J = &(l[j])
	var t = *I
	*I = *J
	*J = t
}

type TypeDef struct {
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

func TypeNameFromTypeRef(ref ast.TypeRef, mod *loader.Module) name.TypeName {
	return name.TypeName { Name: NameFrom(ref.Module, ref.Item, mod) }
}

func TypeNameListFrom(ref_list ([] ast.TypeDeclRef), mod *loader.Module) ([] name.TypeName) {
	var list = make([] name.TypeName, len(ref_list))
	for i, ref := range ref_list {
		list[i] = name.TypeName { Name: NameFrom(ref.Module, ref.Item, mod) }
	}
	return list
}

var __DefaultInit, defaultWrite = (func() (typsys.Type, func(typsys.Type)(typsys.Type)) {
	return nil,
		func(t typsys.Type) typsys.Type { return t }
})()
var __ContentInit, contentWrite = (func() (typsys.TypeDefContent, func(typsys.TypeDefContent)(typsys.TypeDefContent)) {
	return nil,
		func(c typsys.TypeDefContent) typsys.TypeDefContent { return c }
})()
var __TableInit, tableWrite = (func() (([] typsys.DispatchTable), func([] typsys.DispatchTable)([] typsys.DispatchTable)) {
	return nil,
		func(d ([] typsys.DispatchTable)) ([] typsys.DispatchTable) { return d }
})()

func collectTypes(entry *loader.Module, idx loader.Index, al AliasRegistry) (TypeRegistry, *source.Error) {
	var reminder = func() (struct{}, func(func()(*source.Error)) *source.Error) {
		return struct{}{},
			func(f func()(*source.Error)) *source.Error { return f() }
	}
	// *** Postponed Checks ***
	// 1. circular:
	//   (1) a box must NOT unbox to a ref to itself directly or indirectly.
	//   (2) an interface must NOT include itself directly or indirectly.
	var must_check_circular_box, check1_circular_box = reminder()
	var must_check_circular_interface, check2_circular_interface = reminder()
	// 2. variance:
	//   (1) variance defined on parameters of a box must be valid.
	//   (2) variance defined on parameters of an interface must be valid.
	var must_check_boxed_variance, check3_boxed_variance = reminder()
	var must_check_interface_variance, check4_interface_variance = reminder()
	// *************
	// Register types
	var reg = make(TypeRegistry)
	var err = registerTypes(entry, reg)
	if err != nil { return nil, err }
	// Create an ordered type definition list
	var types = make(TypeList, 0, len(reg))
	for _, def := range reg {
		var mod, exists = idx[def.Name.ModuleName]
		if !(exists) { panic("something went wrong") }
		types = append(types, TypeDefWithModule {
			TypeDef: def,
			Module:  mod,
		})
	}
	sort.Sort(types)
	// Check for conflicts with alias
	for _, def := range types {
		var _, conflict = al[def.Name.Name]
		if conflict {
			return nil, source.MakeError(def.Location, E_TypeConflictWithAlias {
				Which: def.Name.String(),
			})
		}
	}
	// Check implemented types
	for _, def := range types {
		var impl_names = TypeNameListFrom(def.AstNode.Impl, def.Module)
		var occurred_names = make(map[name.TypeName] struct{})
		for i, n := range impl_names {
			var loc = def.AstNode.Impl[i].Location
			var _, occurred = occurred_names[n]
			if occurred {
				return nil, source.MakeError(loc, E_DuplicateImplemented {
					Which: n.String(),
				})
			}
			occurred_names[n] = struct{}{}
			var impl_def, exists = reg[n]
			if !(exists) {
				return nil, source.MakeError(loc, E_TypeNotFound {
					Which: n.String(),
				})
			}
			var _, ok = impl_def.AstNode.TypeDef.TypeDef.(ast.InterfaceType)
			if !(ok) {
				return nil, source.MakeError(loc, E_BadImplemented {
					Which: n.String(),
				})
			}
			var impl_params = impl_def.Parameters
			if len(impl_params) != len(def.Parameters) {
				return nil, source.MakeError(loc, E_ImplementedIncompatibleParameters{
					TypeName:      def.Name.String(),
					InterfaceName: n.String(),
				})
			}
			for j := range impl_params {
				var vi = impl_params[j].Variance
				var v = def.Parameters[j].Variance
				if vi != typsys.Invariant && v != vi {
					return nil, source.MakeError(loc, E_ImplementedIncompatibleParameters{
						TypeName:      def.Name.String(),
						InterfaceName: n.String(),
					})
				}
			}
		}
	}
	// Construct default parameter types
	for _, def := range types {
		var ctx = TypeConsContext {
			Module:   def.Module,
			TypeReg:  reg,
			AliasReg: al,
		}
		var err = def.ForEachParameter(func(i uint, p *typsys.Parameter) *source.Error {
			var p_node = &(def.AstNode.Params[i])
			var default_, has_default = p_node.Default.(ast.VariousType)
			if has_default {
				var raw, err = newType(default_, ctx)
				if err != nil { return err }
				p.Default = defaultWrite(raw.Type)
			}
			return nil
		})
		if err != nil { return nil, err }
	}
	// Construct contents
	for _, def := range types {
		var ctx = TypeConsContext {
			Module:   def.Module,
			TypeReg:  reg,
			AliasReg: al,
			ParamVec: def.Parameters,
		}
		switch content := def.AstNode.TypeDef.TypeDef.(type) {
		case ast.BoxedType:
			var kind = typsys.Isomorphic
			if content.Protected { kind = typsys.Protected }
			if content.Opaque { kind = typsys.Opaque }
			var weak = content.Weak
			var inner, err = (func() (typsys.Type, *source.Error) {
				var inner_node, exists = content.Inner.(ast.VariousType)
				if exists {
					var raw, err = newType(inner_node, ctx)
					if err != nil { return nil, err }
					return raw.Type, nil
				} else {
					return typsys.UnitType {}, nil
				}
			})()
			if err != nil { return nil, err }
			def.Content = contentWrite(&typsys.Box {
				BoxKind:      kind,
				WeakWrapping: weak,
				InnerType:    inner,
			})
			_ = must_check_circular_box
			_ = must_check_boxed_variance
		case ast.EnumType:
			if def.Content == nil { panic("something went wrong") }
			// content already generated
		case ast.InterfaceType:
			var raw, err = newTypeFromRepr(content.Methods, ctx)
			if err != nil { return nil, err }
			var methods = raw.Type.(*typsys.NestedType).Content.(typsys.Record)
			var impl_names = TypeNameListFrom(def.AstNode.Impl, def.Module)
			var included = make([] *typsys.TypeDef, len(impl_names))
			for i, n := range impl_names {
				var def, exists = reg[n]
				if !(exists) {
					var loc = def.AstNode.Impl[i].Location
					return nil, source.MakeError(loc, E_TypeNotFound {
						Which: n.String(),
					})
				}
				included[i] = def.TypeDef
				_ = must_check_circular_interface
				_ = must_check_interface_variance
			}
			def.Content = contentWrite(&typsys.Interface {
				Included: included,
				Methods:  methods,
			})
		case ast.NativeType:
			def.Content = contentWrite(&typsys.Native {})
		}
	}
	// Perform all postponed checks
	var check_circular = func(get_deps func(*typsys.TypeDef)([] *typsys.TypeDef)) ([] *typsys.TypeDef) {
		var in = make(map[*typsys.TypeDef] uint)
		var q = make([] *typsys.TypeDef, 0)
		for _, def := range types {
			var deps = get_deps(def.TypeDef.TypeDef)
			for _, dep := range deps {
				in[dep] += 1
			}
		}
		for def, n := range in {
			if n == 0 {
				q = append(q, def)
			}
		}
		for len(q) > 0 {
			var head = q[0]
			q = q[1:]
			var deps = get_deps(head)
			for _, dep := range deps {
				var current = in[dep]
				if !(current >= 1) { panic("something went wrong") }
				var updated = (current - 1)
				in[dep] = updated
				if updated == 0 {
					q = append(q, dep)
				}
			}
		}
		var bad = make([] *typsys.TypeDef, 0)
		for def, n := range in {
			if n > 0 {
				bad = append(bad, def)
			}
		}
		return bad
	}
	var defs_to_strings = func(defs ([] *typsys.TypeDef)) ([] string) {
		var result = make([] string, len(defs))
		for i, def := range defs {
			result[i] = def.Name.String()
		}
		return result
	}
	{ var err = check1_circular_box(func() *source.Error {
		var bad = check_circular(func(def *typsys.TypeDef) ([] *typsys.TypeDef) {
			var box, is_box = def.Content.(*typsys.Box)
			if is_box {
				var nested, is_nested = box.InnerType.(*typsys.NestedType)
				if is_nested {
					var ref, is_ref = nested.Content.(typsys.Ref)
					if is_ref {
						return [] *typsys.TypeDef { ref.Def }
					}
				}
			}
			return nil
		})
		if len(bad) > 0 {
			return source.MakeError(bad[0].Location, E_CircularSubtypingDefinition {
				Which: defs_to_strings(bad),
			})
		} else {
			return nil
		}
	})
	if err != nil { return nil, err } }
	{ var err = check2_circular_interface(func() *source.Error {
		var bad = check_circular(func(def *typsys.TypeDef) ([] *typsys.TypeDef) {
			var interface_, is_interface = def.Content.(*typsys.Interface)
			if is_interface {
				return interface_.Included
			} else {
				return nil
			}
		})
		if len(bad) > 0 {
			return source.MakeError(bad[0].Location, E_CircularInterfaceDefinition {
				Which: defs_to_strings(bad),
			})
		} else {
			return nil
		}
	})
	if err != nil { return nil, err } }
	{ var err = check3_boxed_variance(func() *source.Error {
		for _, def := range types {
			var box, is_box = def.Content.(*typsys.Box)
			if is_box {
				var v = typsys.GetVariance(box.InnerType, def.Parameters)
				var ok, invalid = typsys.MatchVariance(def.Parameters, v)
				if !(ok) {
					var loc = def.AstNode.Name.Location
					return source.MakeError(loc, E_InvalidVarianceOnParameters {
						Which: invalid,
					})
				}
			}
		}
		return nil
	})
	if err != nil { return nil, err } }
	{ var err = check4_interface_variance(func() *source.Error {
		for _, def := range types {
			var interface_, is_interface = def.Content.(*typsys.Interface)
			if is_interface {
				var t = &typsys.NestedType { Content: interface_.Methods }
				var v = typsys.GetVariance(t, def.Parameters)
				var ok, invalid = typsys.MatchVariance(def.Parameters, v)
				if !(ok) {
					var loc = def.AstNode.Name.Location
					return source.MakeError(loc, E_InvalidVarianceOnParameters {
						Which: invalid,
					})
				}
			}
		}
		return nil
	})
	if err != nil { return nil, err } }
	return reg, nil
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
	reg[type_name] = TypeDef {
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
			var arity = len(decl.Params)
			if arity > MAX_TYPE_PARAMETERS {
				return nil, source.MakeError(decl.Name.Location,
					E_TooManyTypeParameters { TypeName: type_name.String() })
			}
			var params = make([] typsys.Parameter, arity)
			for i, p := range decl.Params {
				if p.In && p.Out { panic("something went wrong") }
				var v = typsys.Invariant
				if p.In {
					v = typsys.Contravariant
				}
				if p.Out {
					v = typsys.Covariant
				}
				var n = ast.Id2String(p.Name)
				if !(CheckTypeName(n)) {
					return nil, source.MakeError(p.Name.Location, E_InvalidTypeName {
						Name: ast.Id2String(p.Name),
					})
				}
				params[i] = typsys.Parameter {
					Name:     n,
					Default:  __DefaultInit,
					Variance: v,
				}
			}
			return params, nil
		}
	})()
	if params_err != nil { return nil, params_err }
	*def = typsys.TypeDef {
		TypeAttrs:  attrs,
		Name:       type_name,
		Implements: __TableInit,
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
		def.Content = contentWrite(&typsys.Enum {
			CaseTypes: case_defs,
		})
	}
	return def, nil
}

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


