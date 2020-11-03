package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/kmd"
	"kumachan/parser/ast"
	"kumachan/stdlib"
)


var __KmdPrimitiveTypes = map[loader.Symbol] kmd.TypeKind {
	__Bool:   kmd.Bool,
	__Float:  kmd.Float,
	__Uint32: kmd.Uint32,
	__Int32:  kmd.Int32,
	__Uint64: kmd.Uint64,
	__Int64:  kmd.Int64,
	__Int:    kmd.Int,
	__String: kmd.String,
	__Bytes:  kmd.Binary,
}

type KmdIdMapping  map[loader.Symbol] kmd.TypeId

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
			var _, is_native = g.Value.(*Native)
			if is_native {
				return nil, nil, nil, &KmdError {
					Point:    point,
					Concrete: E_KmdOnNative {},
				}
			} else if len(g.Params) > 0 {
				return nil, nil, nil, &KmdError {
					Point:    point,
					Concrete: E_KmdOnGeneric {},
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
	for sym, id := range mapping {
		var schema, err = GetKmdSchema(sym, nodes, reg, mapping)
		if err != nil { return nil, nil, nil, err }
		sch[id] = schema
	}
	for sym, id := range mapping {
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

func GetFunctionKmdInfo(name string, t Func, mapping KmdIdMapping) FunctionKmdInfo {
	var info FunctionKmdInfo
	if name == KmdAdapterName {
		switch I := t.Input.(type) {
		case *NamedType:
			if len(I.Args) == 0 {
				var in, exists = mapping[I.Name]
				if exists {
					switch O := t.Output.(type) {
					case *NamedType:
						if len(O.Args) == 0 {
							var out, exists = mapping[O.Name]
							if exists {
								info.IsAdapter = true
								info.AdapterId = kmd.AdapterId {
									From: in,
									To:   out,
								}
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
				var in, exists = mapping[I.Name]
				if exists {
					if AreTypesEqualInSameCtx(t.Output, __T_Bool) {
						info.IsValidator = true
						info.ValidatorId = kmd.ValidatorId(in)
					}
				}
			}
		}
	}
	return info
}

func CraftKmdApiFunction (
	sym loader.Symbol,
	reg TypeRegistry,
	nodes TypeDeclNodeInfo,
	id  kmd.TransformerPartId,
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
	var error_t = make_type(stdlib.BinaryError)
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

func IsKmdApiPublic(sym loader.Symbol, reg TypeRegistry) bool {
	var g = reg[sym]
	switch def := g.Value.(type) {
	case *Boxed:
		if def.Opaque || def.Protected {
			return false
		} else {
			return true
		}
	case *Union:
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
	sym      loader.Symbol,
	nodes    TypeDeclNodeInfo,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
) (kmd.Schema, *KmdError) {
	var p = ErrorPointFrom(nodes[sym])
	var g = reg[sym]
	switch def := g.Value.(type) {
	case *Boxed:
		return GetKmdInnerTypeSchema(def.InnerType, p, reg, mapping)
	case *Union:
		var index_map = make(map[kmd.TypeId] uint)
		for i, case_t := range def.CaseTypes {
			var case_id, exists = mapping[case_t.Name]
			if !(exists) { return nil, &KmdError {
				Point:    p,
				Concrete: E_KmdTypeNotSerializable {},
			} }
			index_map[case_id] = uint(i)
		}
		return kmd.UnionSchema {
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
	t        Type,
	p        ErrorPoint,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
) (kmd.Schema, *KmdError) {
	switch T := t.(type) {
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Tuple:
			var elements = make([] *kmd.Type, len(R.Elements))
			for i, el := range R.Elements {
				var el_t, err = GetKmdType(el, p, reg, mapping)
				if err != nil { return nil, &KmdError {
					Point:    p,
					Concrete: E_KmdElementNotSerializable { uint(i) },
				} }
				elements[i] = el_t
			}
			return kmd.TupleSchema { Elements: elements }, nil
		case Bundle:
			var fields = make(map[string] kmd.RecordField)
			for name, field := range R.Fields {
				var field_t, err = GetKmdType(field.Type, p, reg, mapping)
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
	return GetKmdInnerTypeSchema(wrapped, p, reg, mapping)
}

func GetKmdType (
	t        Type,
	p        ErrorPoint,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
) (*kmd.Type, *KmdError) {
	switch T := t.(type) {
	case *NamedType:
		if len(T.Args) == 0 {
			var id, ok = mapping[T.Name]
			if ok {
				var g, exists = reg[T.Name]
				if !(exists) { panic("something went wrong") }
				switch def := g.Value.(type) {
				case *Boxed:
					switch T := def.InnerType.(type) {
					case *AnonymousType:
						switch T.Repr.(type) {
						case Bundle:
							return kmd.AlgebraicType(kmd.Record, id), nil
						}
					}
					return kmd.AlgebraicType(kmd.Tuple, id), nil
				case *Union:
					return kmd.AlgebraicType(kmd.Union, id), nil
				}
			}
			kind, ok := __KmdPrimitiveTypes[T.Name]
			if ok {
				return kmd.PrimitiveType(kind), nil
			}
		} else if len(T.Args) == 1 {
			var arg = T.Args[0]
			var arg_t, err = GetKmdType(arg, p, reg, mapping)
			if err != nil { return nil, err }
			if T.Name == __Array {
				return kmd.ContainerType(kmd.Array, arg_t), nil
			} else if T.Name == __Maybe {
				return kmd.ContainerType(kmd.Optional, arg_t), nil
			}
		}
	}
	return nil, &KmdError {
		Point:    p,
		Concrete: E_KmdTypeNotSerializable {},
	}
}

