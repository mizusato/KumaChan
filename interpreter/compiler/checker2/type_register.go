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
)


type TypeRegistry (map[name.TypeName] TypeDef)
type TypeList ([] TypeDefWithModule)
type TypeDefWithModule struct {
	TypeDef
	Module  *loader.Module
}
type TypeDef struct {
	*typsys.TypeDef
	AstNode  *ast.DeclType
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

var defaultInit, defaultWrite = (func() (typsys.Type, func(typsys.Type)(typsys.Type)) {
	return nil,
		func(t typsys.Type) typsys.Type { return t }
})()
var contentInit, contentWrite = (func() (typsys.TypeDefContent, func(typsys.TypeDefContent)(typsys.TypeDefContent)) {
	return nil,
		func(c typsys.TypeDefContent) typsys.TypeDefContent { return c }
})()
var tableInit, tableWrite = (func() (([] typsys.DispatchTable), func([] typsys.DispatchTable)([] typsys.DispatchTable)) {
	return nil,
		func(d ([] typsys.DispatchTable)) ([] typsys.DispatchTable) { return d }
})()

func collectTypes(entry *loader.Module, idx loader.Index, al AliasRegistry) (TypeRegistry, source.Errors) {
	var step_ = func(types TypeList) (func(func(TypeDefWithModule)(*source.Error)) source.Errors) {
		return func(f func(TypeDefWithModule) *source.Error) source.Errors {
			var errs source.Errors
			for _, def := range types {
				source.ErrorsJoin(&errs, f(def))
			}
			return errs
		}
	}
	var check = func() (struct{}, func(func()(*source.Error)) source.Errors) {
		return struct{}{}, func(f func()(*source.Error)) source.Errors {
			return source.ErrorsFrom(f())
		}
	}
	// ************************
	// --- register types ---
	var reg = make(TypeRegistry)
	var err = registerTypes(entry, reg)
	if err != nil { return nil, source.ErrorsFrom(err) }
	// --- create an ordered type definition list ---
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
	// ************************
	// --- main steps ---
	var step = step_(types)
	// 1. check for name conflicts with alias
	//   (*) a type cannot have a name that is identical to an alias.
	var step1_check_alias = step
	// 2. fill implemented interface list
	//   (1) a type cannot have two or more identical implemented types.
	//   (2) an implemented type must be an interface type.
	//   (3) compatible parameters are required for implemented interfaces.
	var step2_fill_impl = step
	// 3. construct default parameter types
	//   (1) construct the default type for each type parameter.
	var step3_construct_default = step
	// 4. construct contents
	//   (1) construct the inner type for each box type.
	//   (2) construct the methods record for each interface type.
	//   (3) contents of enum types should have been constructed.
	//   (4) construct a unique content for each native type.
	var step4_construct_content = step
	// --- additional checks ---
	// 1. check for circular definition
	//   (1) a box must NOT unbox to a ref to itself directly or indirectly.
	//   (2) an interface must NOT include itself directly or indirectly.
	var must_check_circular_box, check1_circular_box = check()
	var must_check_circular_interface, check2_circular_interface = check()
	// 2. check for variance validity
	//   (1) variance defined on parameters of a box must be valid.
	//   (2) variance defined on parameters of an interface must be valid.
	var must_check_boxed_variance, check3_boxed_variance = check()
	var must_check_interface_variance, check4_interface_variance = check()
	// ************************
	{ var err = step1_check_alias(func(def TypeDefWithModule) *source.Error {
		var _, conflict = al[def.Name.Name]
		if conflict {
			return source.MakeError(def.Location,
				E_TypeConflictWithAlias { Which: def.Name.String() })
		} else {
			return nil
		}
	})
	if err != nil { return nil, err } }
	// ---------
	{ var err = step2_fill_impl(func(def TypeDefWithModule) *source.Error {
		var impl_names = make([] name.TypeName, len(def.AstNode.Impl))
		for i, ref := range def.AstNode.Impl {
			impl_names[i] = name.TypeName {
				Name: NameFrom(ref.Module, ref.Item, def.Module),
			}
		}
		var impl_defs = make([] *typsys.TypeDef, len(impl_names))
		var occurred_names = make(map[name.TypeName] struct{})
		for i, n := range impl_names {
			var loc = def.AstNode.Impl[i].Location
			var _, occurred = occurred_names[n]
			if occurred {
				return source.MakeError(loc,
					E_DuplicateImplemented { Which: n.String() })
			}
			occurred_names[n] = struct{}{}
			var impl_def, exists = reg[n]
			if !(exists) {
				return source.MakeError(loc,
					E_TypeNotFound { Which: n.String() })
			}
			var ok = (func() bool {
				var ast_content, specified = impl_def.AstNode.TypeDef.(ast.VariousTypeDef)
				if !(specified) { return false }
				var _, ok = ast_content.TypeDef.(ast.InterfaceType)
				return ok
			})()
			if !(ok) {
				return source.MakeError(loc,
					E_BadImplemented { Which: n.String() })
			}
			var impl_params = impl_def.Parameters
			if len(impl_params) != len(def.Parameters) {
				return source.MakeError(loc,
					E_ImplementedIncompatibleParameters {
						TypeName:      def.Name.String(),
						InterfaceName: n.String(),
					})
			}
			for j := range impl_params {
				var vi = impl_params[j].Variance
				var v = def.Parameters[j].Variance
				if vi != typsys.Invariant && v != vi {
					return source.MakeError(loc,
						E_ImplementedIncompatibleParameters {
							TypeName:      def.Name.String(),
							InterfaceName: n.String(),
						})
				}
			}
			impl_defs[i] = impl_def.TypeDef
		}
		def.Implements = impl_defs
		return nil
	})
	if err != nil { return nil, err } }
	// ---------
	{ var err = step3_construct_default(func(def TypeDefWithModule) *source.Error {
		var ctx = TypeConsContext {
			Module:   def.Module,
			TypeReg:  reg,
			AliasReg: al,
		}
		var defaults = make([] struct { *typsys.Parameter; typsys.Type }, 0)
		var err = def.ForEachParameter(func(i uint, p *typsys.Parameter) *source.Error {
			var p_node = &(def.AstNode.Params[i])
			var default_, has_default = p_node.Default.(ast.VariousType)
			if has_default {
				var raw, err = newType(default_, ctx)
				if err != nil { return err }
				defaults = append(defaults, struct { *typsys.Parameter; typsys.Type } {
					Parameter: p,
					Type:      raw.Type,
				})
			}
			return nil
		})
		if err != nil { return err }
		// note: write all at once (avoid interference)
		for _, item := range defaults {
			item.Parameter.Default = defaultWrite(item.Type)
		}
		return nil
	})
	if err != nil { return nil, err } }
	// ---------
	{ var err = step4_construct_content(func(def TypeDefWithModule) *source.Error {
		var ctx = TypeConsContext {
			Module:   def.Module,
			TypeReg:  reg,
			AliasReg: al,
			ParamVec: def.Parameters,
		}
		var ast_content, specified = def.AstNode.TypeDef.(ast.VariousTypeDef)
		if !(specified) {
			return source.MakeError(def.AstNode.Name.Location,
				E_BlankTypeDefinition {})
		}
		switch content := ast_content.TypeDef.(type) {
		case ast.BoxedType:
			var kind = typsys.Isomorphic
			if content.Protected { kind = typsys.Protected }
			if content.Opaque { kind = typsys.Opaque }
			var weak = content.Weak
			var raw, err = newType(content.Inner, ctx)
			if err != nil { return err }
			var inner = raw.Type
			def.Content = contentWrite(&typsys.Box {
				BoxKind:      kind,
				WeakWrapping: weak,
				InnerType:    inner,
			})
			_ = must_check_circular_box
			_ = must_check_boxed_variance
		case ast.InterfaceType:
			var raw, err = newTypeFromRepr(content.Methods, ctx)
			if err != nil { return err }
			var methods = raw.Type.(*typsys.NestedType).Content.(typsys.Record)
			var included = make([] *typsys.Interface, len(def.Implements))
			for i, impl_def := range def.Implements {
				included[i] = impl_def.Content.(*typsys.Interface)
			}
			def.Content = contentWrite(&typsys.Interface {
				Included: included,
				Methods:  methods,
			})
			_ = must_check_circular_interface
			_ = must_check_interface_variance
		case ast.EnumType:
			if def.Content == nil { panic("something went wrong") }
			// content already generated
		case ast.NativeType:
			def.Content = contentWrite(&typsys.Native {})
		default:
			panic("impossible branch")
		}
		return nil
	})
	if err != nil { return nil, err } }
	// ************************
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
	// ---------
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
	// ---------
	{ var err = check2_circular_interface(func() *source.Error {
		var bad = check_circular(func(def *typsys.TypeDef) ([] *typsys.TypeDef) {
			var _, is_interface = def.Content.(*typsys.Interface)
			if is_interface {
				return def.Implements
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
	// ---------
	{ var err = check3_boxed_variance(func() *source.Error {
		for _, def := range types {
			var box, is_box = def.Content.(*typsys.Box)
			if is_box {
				var v = typsys.GetVariance(box.InnerType, def.Parameters)
				var ok, invalid = typsys.MatchVariance(def.Parameters, v)
				if !(ok) {
					var loc = def.AstNode.Name.Location
					return source.MakeError(loc,
						E_InvalidVarianceOnParameters { Which: invalid })
				}
			}
		}
		return nil
	})
	if err != nil { return nil, err } }
	// ---------
	{ var err = check4_interface_variance(func() *source.Error {
		for _, def := range types {
			var interface_, is_interface = def.Content.(*typsys.Interface)
			if is_interface {
				var t = &typsys.NestedType { Content: interface_.Methods }
				var v = typsys.GetVariance(t, def.Parameters)
				var ok, invalid = typsys.MatchVariance(def.Parameters, v)
				if !(ok) {
					var loc = def.AstNode.Name.Location
					return source.MakeError(loc,
						E_InvalidVarianceOnParameters { Which: invalid })
				}
			}
		}
		return nil
	})
	if err != nil { return nil, err } }
	// ---------
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
	var type_item_name = ast.Id2String(decl.Name)
	if !(isValidTypeItemName(type_item_name)) {
		return nil, source.MakeError(decl.Name.Location,
			E_InvalidTypeName { Name: type_item_name })
	}
	var type_name = name.MakeTypeName(mod.Name, type_item_name)
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
		return nil, source.MakeError(loc,
			E_InvalidMetadata { Reason: meta_err.Error() })
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
				return nil, source.MakeError(loc,
					E_TypeParametersOnCaseType {})
			}
			return ci.Enum.Parameters, nil
		} else {
			var arity = len(decl.Params)
			if arity > MaxTypeParameters {
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
				var p_name = ast.Id2String(p.Name)
				if !(isValidTypeItemName(p_name)) {
					return nil, source.MakeError(p.Name.Location,
						E_InvalidTypeName { Name: p_name })
				}
				params[i] = typsys.Parameter {
					Name:     p_name,
					Default:  defaultInit,
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
		Tables:     tableInit,
		Parameters: params,
		Content:    contentInit,
		CaseInfo:   ci,
	}
	var enum, is_enum = (func() (ast.EnumType, bool) {
		var ast_content, specified = decl.TypeDef.(ast.VariousTypeDef)
		if !(specified) { return ast.EnumType {}, false }
		var enum, is_enum = ast_content.TypeDef.(ast.EnumType)
		return enum, is_enum
	})()
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

