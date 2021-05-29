package checker

import (
	"errors"
	"kumachan/misc/rpc"
	"kumachan/misc/rpc/kmd"
	"kumachan/lang"
	"kumachan/lang/parser/ast"
	"kumachan/compiler/loader"
	"kumachan/stdlib"
	. "kumachan/misc/util/error"
)


var __KmdPrimitiveTypes = map[lang.Symbol] kmd.TypeKind {
	__Bool:          kmd.Bool,
	__Integer:       kmd.Integer,
	__NormalFloat:   kmd.Float,
	__NormalComplex: kmd.Complex,
	__String:        kmd.String,
	__Bytes:         kmd.Binary,
}

type KmdIdMapping  map[lang.Symbol] kmd.TypeId

type KmdStmtInjection  map[string] []ast.VariousStatement


func CollectKmdApi (
	reg        TypeRegistry,
	nodes      TypeDeclNodeInfo,
	raw_index  loader.Index,
) (KmdIdMapping, kmd.SchemaTable, KmdStmtInjection, *KmdError) {
	var mapping = make(KmdIdMapping)
	var sch = make(kmd.SchemaTable)
	var inj = make(KmdStmtInjection)
	for sym, g := range reg {
		var point = ErrorPointFrom(nodes[sym])
		var conf = g.Tags.DataConfig
		var labelled_serializable = (conf.Name != "")
		if labelled_serializable {
			var _, is_native = g.Definition.(*Native)
			if is_native {
				return nil, nil, nil, &KmdError {
					Point:    point,
					Concrete: E_KmdOnNative {},
				}
			} else {
				var raw_mod = raw_index[sym.ModuleName]
				mapping[sym] = kmd.TypeId {
					TypeIdFuzzy: kmd.TypeIdFuzzy {
						Vendor:  raw_mod.Vendor,
						Project: raw_mod.Project,
						Name:    conf.Name,
					},
					Version: conf.Version,
				}
			}
		}
	}
	var mono_mapping = make(map[kmd.TypeId] ([] kmd.TypeId))
	var mono_args = make(map[kmd.TypeId] ([] Type))
	var add_mono = func(id kmd.TypeId, mono kmd.TypeId) {
		existing, exists := mono_mapping[id]
		if exists {
			mono_mapping[id] = append(existing, mono)
		} else {
			mono_mapping[id] = [] kmd.TypeId { mono }
		}
	}
	for sym, id := range mapping {
		var g = reg[sym]
		if len(g.Params) == 0 {
			switch def := g.Definition.(type) {
			case *Boxed:
				switch T := def.InnerType.(type) {
				case *NamedType:
					var inner_id, inner_has_id = mapping[T.Name]
					if inner_has_id && len(T.Args) > 0 {
						var inner_g, exists = reg[T.Name]
						if !(exists) { panic("something went wrong") }
						if len(T.Args) != len(inner_g.Params) { panic("something went wrong") }
						mono_args[id] = T.Args
						add_mono(inner_id, id)
					}
				}
			}
		}
	}
	for sym, id := range mapping {
		var point = ErrorPointFrom(nodes[sym])
		var g = reg[sym]
		if len(g.Params) > 0 {
			var mono_types = make([] [] kmd.TypeId, 0)
			var args_maps = make([] [] uint, 0)
			var m, exists = mono_mapping[id]
			if exists {
				mono_types = append(mono_types, m)
				var args_map = make([] uint, len(g.Params))
				for i, _ := range g.Params { args_map[i] = uint(i) }
				args_maps = append(args_maps, args_map)
			} else {
				var cur = g
				for cur.CaseInfo.IsCaseType {
					var case_info = cur.CaseInfo
					var enum_sym = case_info.EnumName
					var enum_id = mapping[enum_sym]
					var m, exists = mono_mapping[enum_id]
					if exists {
						mono_types = append(mono_types, m)
						args_maps = append(args_maps, case_info.CaseParams)
					}
					cur = reg[enum_sym]
				}
			}
			for i, _ := range mono_types {
				var group = mono_types[i]
				var args_map = args_maps[i]
				for _, mono_id := range group {
					var decorated_id = id.Decorate(mono_id)
					var args = make([] Type, len(args_map))
					for i := 0; i < len(args_map); i += 1 {
						args[i] = mono_args[mono_id][args_map[i]]
					}
					var schema, err = GetKmdSchema (
						decorated_id, sym, nodes, reg, mapping,
						args, mono_id,
					)
					if err != nil { return nil, nil, nil, err }
					_, exists := sch[decorated_id]
					if exists { return nil, nil, nil, &KmdError {
						Point:    point,
						Concrete: E_KmdDuplicateType {
							Id: decorated_id.String(),
						},
					} }
					sch[decorated_id] = schema
				}
			}
		} else {
			var schema, err = GetKmdSchema (
				id, sym, nodes, reg, mapping,
				nil, kmd.TypeId {},
			)
			if err != nil { return nil, nil, nil, err }
			_, exists := sch[id]
			if exists { return nil, nil, nil, &KmdError {
				Point:    point,
				Concrete: E_KmdDuplicateType {
					Id: id.String(),
				},
			} }
			sch[id] = schema
		}
	}
	for sym, id := range mapping {
		var g = reg[sym]
		if len(g.Params) > 0 {
			continue
		}
		var mod = sym.ModuleName
		var serializer = ast.VariousStatement {
			Node:      nodes[sym],
			Statement: CraftKmdApiFunction(sym, reg, nodes,
				kmd.SerializerId { TypeId: id }),
		}
		var deserializer = ast.VariousStatement {
			Node: nodes[sym],
			Statement: CraftKmdApiFunction(sym, reg, nodes,
				kmd.DeserializerId { TypeId: id }),
		}
		var _, exists = inj[mod]
		if exists {
			inj[mod] = append(inj[mod], serializer, deserializer)
		} else {
			inj[mod] = [] ast.VariousStatement { serializer, deserializer }
		}
	}
	return mapping, sch, inj, nil
}

func EnforceGoodKmdFunctions(types TypeRegistry, idx Index) ([] E) {
	type AdapterKey struct {
		In   lang.Symbol
		Out  lang.Symbol
	}
	var errs = make([] E, 0)
	var add_error = func(p ErrorPoint, e ConcreteKmdError) {
		errs = append(errs, &KmdError {
			Point:    p,
			Concrete: e,
		})
	}
	var has_adapter = make(map[AdapterKey] bool)
	var has_validator = make(map[lang.Symbol] bool)
	for mod_name, mod := range idx {
		for _, group := range mod.Functions {
			for _, item := range group {
				if item.IsAdapter {
					var key = AdapterKey {
						In:  item.KmdIn,
						Out: item.KmdOut,
					}
					if has_adapter[key] {
						add_error(item.Point, E_KmdDuplicateAdapter {})
					} else {
						has_adapter[key] = true
					}
				}
				if item.IsValidator {
					var type_name = item.KmdIn
					if has_validator[type_name] {
						add_error(item.Point, E_KmdDuplicateValidator {})
					} else {
						if type_name.ModuleName != mod_name {
							add_error(item.Point, E_KmdValidatorNotInSameModule {})
						} else {
							has_validator[type_name] = true
						}
					}
				}
			}
		}
	}
	for type_name, t := range types {
		if t.Tags.DeclaredSerializable() {
			var boxed, is_boxed = t.Definition.(*Boxed)
			if is_boxed {
				if boxed.Protected || boxed.Opaque {
					if !(has_validator[type_name]) {
						add_error(ErrorPointFrom(t.Node), E_KmdMissingValidator {})
					}
				} else {
					if has_validator[type_name] {
						add_error(ErrorPointFrom(t.Node), E_KmdSuspiciousValidator {})
					}
				}
			}
		}
	}
	if len(errs) > 0 {
		return errs
	} else {
		return nil
	}
}

func GetFunctionKmdInfo(name string, t Func, mapping KmdIdMapping) FunctionKmdInfo {
	var info FunctionKmdInfo
	if name == KmdAdapterName {
		switch I := t.Input.(type) {
		case *NamedType:
			if len(I.Args) == 0 {
				var in_name = I.Name
				var in, exists = mapping[in_name]
				if exists {
					switch O := t.Output.(type) {
					case *NamedType:
						if len(O.Args) == 0 {
							var out_name = O.Name
							var out, exists = mapping[out_name]
							if exists {
								info.IsAdapter = true
								info.AdapterId = kmd.AdapterId {
									From: in,
									To:   out,
								}
								info.KmdIn = in_name
								info.KmdOut = out_name
							}
						}
					}
				}
			}
		}
	} else if name == KmdValidatorName {
		switch I := t.Input.(type) {
		case *NamedType:
			if len(I.Args) == 0 {
				var in_name = I.Name
				var in, exists = mapping[in_name]
				if exists {
					if TypeEqualWithoutContext(t.Output, __T_Bool) {
						info.IsValidator = true
						info.ValidatorId = kmd.ValidatorId(in)
						info.KmdIn = in_name
					}
				}
			}
		}
	}
	return info
}

func CraftKmdApiFunction (
	sym   lang.Symbol,
	reg   TypeRegistry,
	nodes TypeDeclNodeInfo,
	id    kmd.TransformerPartId,
) ast.DeclFunction {
	var node = nodes[sym]
	var make_type = func(name string) ast.VariousType {
		return ast.VariousType {
			Node: node,
			Type: ast.TypeRef {
				Node: node,
				Id: ast.Identifier {
					Node: node,
					Name: ([] rune)(name),
				},
				TypeArgs: make([] ast.VariousType, 0),
			},
		}
	}
	var binary_t = make_type(stdlib.Bytes)
	var object_t = make_type(sym.SymbolName)
	var error_t = make_type(stdlib.Error)
	var name string
	var sig ast.ReprFunc
	switch id.(type) {
	case kmd.SerializerId:
		name = KmdSerializerName
		sig = ast.ReprFunc {
			Node:   node,
			Input:  object_t,
			Output: binary_t,
		}
	case kmd.DeserializerId:
		name = KmdDeserializerName
		sig = ast.ReprFunc {
			Node:   node,
			Input:  binary_t,
			Output: ast.VariousType {
				Node: node,
				Type: ast.TypeRef {
					Node:     node,
					Id:       ast.Identifier {
						Node: node,
						Name: ([] rune)(stdlib.Result),
					},
					TypeArgs: [] ast.VariousType { object_t, error_t },
				},
			},
		}
	default:
		panic("impossible branch")
	}
	return ast.DeclFunction {
		Node:     node,
		Public:   IsKmdApiPublic(sym, reg),
		Name:     ast.Identifier {
			Node: node,
			Name: ([] rune)(name),
		},
		Params:   make([] ast.TypeParam, 0),
		Implicit: make([] ast.VariousType, 0),
		Repr:     sig,
		Body:     ast.VariousBody {
			Node: node,
			Body: ast.KmdApiFuncBody { Id: id },
		},
	}
}

func IsKmdApiPublic(sym lang.Symbol, reg TypeRegistry) bool {
	var g = reg[sym]
	switch def := g.Definition.(type) {
	case *Boxed:
		if def.Opaque || def.Protected {
			return false
		} else {
			return true
		}
	case *Enum:
		for _, case_t := range def.CaseTypes {
			var case_public = IsKmdApiPublic(case_t.Name, reg)
			if !(case_public) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

func GetKmdSchema (
	id       kmd.TypeId,
	sym      lang.Symbol,
	nodes    TypeDeclNodeInfo,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
	args     [] Type,
	mono_id  kmd.TypeId,
) (kmd.Schema, *KmdError) {
	var p = ErrorPointFrom(nodes[sym])
	var g = reg[sym]
	switch def := g.Definition.(type) {
	case *Boxed:
		var inner Type
		if len(g.Params) > 0 {
			inner = FillTypeArgsWithDefaults(def.InnerType, args, g.Defaults)
		} else {
			inner = def.InnerType
		}
		var generic = len(g.Params) > 0
		return GetKmdInnerTypeSchema(id, generic, inner, p, reg, mapping)
	case *Enum:
		var index_map = make(map[kmd.TypeId] uint)
		for i, case_t := range def.CaseTypes {
			var case_id, exists = mapping[case_t.Name]
			if len(g.Params) > 0 && len(case_t.Params) > 0 {
				case_id = case_id.Decorate(mono_id)
			}
			if !(exists) { return nil, &KmdError {
				Point:    p,
				Concrete: E_KmdTypeNotSerializable {},
			} }
			index_map[case_id] = uint(i)
		}
		return kmd.EnumSchema {
			CaseIndexMap: index_map,
		}, nil
	default:
		return nil, &KmdError {
			Point:    p,
			Concrete: E_KmdTypeNotSerializable {},
		}
	}
}

func GetKmdInnerTypeSchema (
	id       kmd.TypeId,
	generic  bool,
	t        Type,
	p        ErrorPoint,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
) (kmd.Schema, *KmdError) {
	switch T := t.(type) {
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Tuple:
			var length = len(R.Elements)
			if length == 1 {
				switch E := R.Elements[0].(type) {
				case *AnonymousType:
					switch E.Repr.(type) {
					case Unit:
						var zero_tuple =
							&AnonymousType { Tuple { Elements: [] Type {} } }
						return GetKmdInnerTypeSchema(
							id, generic, zero_tuple, p, reg, mapping)
					}
				}
			}
			var elements = make([] *kmd.Type, length)
			for i, el := range R.Elements {
				var mono_ok bool
				var mono_id kmd.TypeId
				if length == 1 {
					// length = 1, subtyping (must be a non-generic subtype)
					if generic { return nil, &KmdError {
						Point:    p,
						Concrete: E_KmdElementNotSerializable { uint(i) },
					} }
					mono_ok = true
					mono_id = id
				} else {
					// length > 1, tuple
					mono_ok = false
				}
				var el_t, err = GetKmdType(el, p, reg, mapping, mono_ok, mono_id)
				if err != nil { return nil, &KmdError {
					Point:    p,
					Concrete: E_KmdElementNotSerializable { uint(i) },
				} }
				elements[i] = el_t
			}
			return kmd.TupleSchema { Elements: elements }, nil
		case Record:
			var fields = make(map[string] kmd.RecordField)
			for name, field := range R.Fields {
				var field_t, err = GetKmdType(field.Type, p, reg, mapping, false, kmd.TypeId {})
				if err != nil { return nil, &KmdError {
					Point:    p,
					Concrete: E_KmdFieldNotSerializable { name },
				} }
				fields[name] = kmd.RecordField {
					Type:  field_t,
					Index: field.Index,
				}
			}
			return kmd.RecordSchema { Fields: fields }, nil
		}
	}
	var wrapped = &AnonymousType { Tuple { Elements: []Type { t } } }
	return GetKmdInnerTypeSchema(id, generic, wrapped, p, reg, mapping)
}

func GetKmdType (
	t         Type,
	p         ErrorPoint,
	reg       TypeRegistry,
	mapping   KmdIdMapping,
	mono_ok   bool,
	mono_id   kmd.TypeId,
) (*kmd.Type, *KmdError) {
	switch T := t.(type) {
	case *NamedType:
		kind, ok := __KmdPrimitiveTypes[T.Name]
		if ok {
			if len(T.Args) > 0 { panic("something went wrong") }
			return kmd.PrimitiveType(kind), nil
		}
		if len(T.Args) == 1 {
			var arg = T.Args[0]
			var arg_t, err = GetKmdType(arg, p, reg, mapping, false, kmd.TypeId {})
			if err != nil { return nil, err }
			if T.Name == __List {
				return kmd.ContainerType(kmd.Array, arg_t), nil
			} else if T.Name == __Maybe {
				return kmd.ContainerType(kmd.Optional, arg_t), nil
			}
		}
		id, ok := mapping[T.Name]
		if ok {
			var g, exists = reg[T.Name]
			if !(exists) { panic("something went wrong") }
			if len(g.Params) != len(T.Args) { panic("something went wrong") }
			if len(g.Params) > 0 {
				if mono_ok {
					id = id.Decorate(mono_id)
				} else {
					goto error
				}
			}
			switch def := g.Definition.(type) {
			case *Boxed:
				switch T := def.InnerType.(type) {
				case *AnonymousType:
					switch T.Repr.(type) {
					case Record:
						return kmd.AlgebraicType(kmd.Record, id), nil
					}
				}
				return kmd.AlgebraicType(kmd.Tuple, id), nil
			case *Enum:
				return kmd.AlgebraicType(kmd.Enum, id), nil
			}
		}
		error:
	}
	return nil, &KmdError {
		Point:    p,
		Concrete: E_KmdTypeNotSerializable {},
	}
}


type ServiceMethodSignature struct {
	Input       kmd.TypeId
	Output      kmd.TypeId
	MultiValue  bool
}

func CollectServices (
	index      loader.Index,
	functions  FunctionStore,
	reg        TypeRegistry,
	table      kmd.SchemaTable,
	mapping    KmdIdMapping,
) (rpc.ServiceIndex, *ServiceError) {
	var services = make(rpc.ServiceIndex)
	for mod_name, mod := range index {
		if !(mod.IsService) { continue }
		var id = mod.ServiceIdentifier
		var arg_type_sym = lang.MakeSymbol(mod_name, mod.ServiceArgTypeName)
		_, exists := reg[arg_type_sym]
		if !(exists) { panic("something went wrong") }
		arg_type_id, exists := mapping[arg_type_sym]
		if !(exists) { panic("something went wrong") }
		var arg_type = table.GetTypeFromId(arg_type_id)
		var ctor = rpc.ServiceConstructorInterface { ArgType: arg_type }
		var methods = make(map[string] rpc.ServiceMethodInterface)
		for _, name := range mod.ServiceMethodNames {
			var group, exists = functions[mod_name][name]
			if !(exists) { panic("something went wrong") }
			var filtered = make([] FunctionReference, 0)
			for _, f := range group {
				if !(f.IsImported) && f.Function.Tags.IsServiceMethod {
					filtered = append(filtered, f)
				}
			}
			if len(filtered) != 1 { panic("something went wrong") }
			var f = filtered[0]
			var t = f.Function.DeclaredType
			var sig, err = GetServiceMethodSignature(t, mod_name, reg, table, mapping)
			if err != nil { return nil, &ServiceError {
				Point:    ErrorPointFrom(f.Function.Node),
				Concrete: E_ServiceMethodInvalidSignature {
					Reason: err.Error(),
				},
			} }
			methods[name] = sig
		}
		_, exists = services[id]
		if exists { return nil, &ServiceError {
			Point:    ErrorPointFrom(mod.AST.Node),
			Concrete: E_ServiceDuplicateModule {
				Id: rpc.DescribeServiceIdentifier(id),
			},
		} }
		services[id] = rpc.ServiceInterface {
			ServiceIdentifier: id,
			Constructor:       ctor,
			Methods:           methods,
		}
	}
	return services, nil
}

func GetServiceMethodSignature (
	sig      Func,
	mod      string,
	reg      TypeRegistry,
	table    kmd.SchemaTable,
	mapping  KmdIdMapping,
) (rpc.ServiceMethodInterface, error) {
	var throw = func(err_desc string) (rpc.ServiceMethodInterface, error) {
		return rpc.ServiceMethodInterface{}, errors.New(err_desc)
	}
	var get_type = func(sym lang.Symbol) *kmd.Type {
		var type_id, exists = mapping[sym]
		if !(exists) { panic("something went wrong") }
		return table.GetTypeFromId(type_id)
	}
	var result = rpc.ServiceMethodInterface {}
	var instance_t = ServiceInstanceType(mod)
	var in = sig.Input
	var out = sig.Output
	switch T := in.(type) {
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Tuple:
			if len(R.Elements) != 2 {
				return throw("wrong arity")
			}
			var args = R.Elements
			if !(TypeEqualWithoutContext(args[0], instance_t)) {
				return throw("first input should be service instance type")
			}
			switch req_t := args[1].(type) {
			default:
				return throw("second input is invalid")
			case *NamedType:
				if len(req_t.Args) != 0 {
					return throw("second input is invalid")
				}
				var g = reg[req_t.Name]
				if !(g.Tags.DeclaredSerializable()) {
					return throw("second input should be serializable")
				}
				result.ArgType = get_type(req_t.Name)
				goto in_ok
			}
		}
	}
	return throw("invalid input type")
in_ok:
	switch T := out.(type) {
	case *NamedType:
		if T.Name != __Async && T.Name != __Observable {
			return throw("output should be of the Async or Observable type")
		}
		result.MultiValue = (T.Name == __Observable)
		if len(T.Args) != 2 {
			return throw("invalid output type")
		}
		if !(TypeEqualWithoutContext(T.Args[1], __ErrorType)) {
			return throw("invalid exception type")
		}
		switch res_t := T.Args[0].(type) {
		case *NamedType:
			if len(res_t.Args) != 0 {
				return throw("invalid output type")
			}
			var g = reg[res_t.Name]
			if !(g.Tags.DeclaredSerializable()) {
				return throw("output should be serializable")
			}
			result.RetType = get_type(res_t.Name)
			goto out_ok
		}
	}
	return throw("invalid output type")
out_ok:
	return result, nil
}

