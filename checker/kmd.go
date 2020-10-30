package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/kmd"
	"kumachan/parser/ast"
	"kumachan/stdlib"
)


const KmdSerializerName = "data-serialize"
const KmdDeserializerName = "data-deserialize"

type KmdIdMapping  map[loader.Symbol] kmd.TypeId

type KmdStmtInjection  map[string] []ast.VariousStatement


func CollectKmdApi (
	reg        TypeRegistry,
	nodes      TypeDeclNodeInfo,
	raw_index  loader.Index,
) (KmdIdMapping, KmdStmtInjection, *KmdError) {
	var mapping = make(KmdIdMapping)
	var inj = make(KmdStmtInjection)
	for sym, g := range reg {
		var point = ErrorPointFrom(nodes[sym])
		var conf = g.Tags.DataConfig
		var labelled_serializable = (conf.Name != "")
		if labelled_serializable {
			var _, is_native = g.Value.(*Native)
			if is_native {
				return nil, nil, &KmdError {
					Point:    point,
					Concrete: E_KmdOnNative {},
				}
			} else if len(g.Params) > 0 {
				return nil, nil, &KmdError {
					Point:    point,
					Concrete: E_KmdOnGeneric {},
				}
			} else {
				mapping[sym] = kmd.TypeId {
					TypeIdFuzzy: kmd.TypeIdFuzzy {
						Vendor: raw_index[sym.ModuleName].Vendor,
						Name:   conf.Name,
					},
					Version: conf.Version,
				}
			}
		}
	}
	for sym, _ := range mapping {
		var err = CheckKmdSerializable(sym, nodes, reg, mapping)
		if err != nil { return nil, nil, err }
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
	return mapping, inj, nil
}

func CraftKmdApiFunction (
	sym loader.Symbol,
	reg TypeRegistry,
	nodes TypeDeclNodeInfo,
	id  kmd.TransformerPartId,
) ast.DeclFunction {
	var node = nodes[sym]
	var binary_t = ast.VariousType {
		Node: node,
		Type: ast.TypeRef {
			Node: node,
			Id: ast.Identifier {
				Node: node,
				Name: ([] rune)(stdlib.Bytes),
			},
			TypeArgs: make([] ast.VariousType, 0),
		},
	}
	var object_t = ast.VariousType {
		Node: node,
		Type: ast.TypeRef {
			Node:     node,
			Id:       ast.Identifier {
				Node: node,
				Name: ([] rune)(sym.SymbolName),
			},
			TypeArgs: make([] ast.VariousType, 0),
		},
	}
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
			Output: object_t,
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

func CheckKmdSerializable (
	sym      loader.Symbol,
	nodes    TypeDeclNodeInfo,
	reg      TypeRegistry,
	mapping  KmdIdMapping,
) *KmdError {
	var p = ErrorPointFrom(nodes[sym])
	var g = reg[sym]
	switch def := g.Value.(type) {
	case *Boxed:
		return CheckKmdInnerTypeSerializable(def.InnerType, p, mapping)
	case *Union:
		for _, case_t := range def.CaseTypes {
			var err = CheckKmdSerializable(case_t.Name, nodes, reg, mapping)
			if err != nil { return err }
		}
		return nil
	default:
		return &KmdError {
			Point:    p,
			Concrete: E_KmdTypeNotSerializable {},
		}
	}
}

func CheckKmdInnerTypeSerializable(t Type, p ErrorPoint, mapping KmdIdMapping) *KmdError {
	switch T := t.(type) {
	case *AnonymousType:
		switch R := T.Repr.(type) {
		case Tuple:
			for i, el := range R.Elements {
				var err = CheckKmdInnerTypeSerializable(el, p, mapping)
				if err != nil {
					return &KmdError {
						Point:    p,
						Concrete: E_KmdElementNotSerializable { uint(i) },
					}
				}
			}
			return nil
		case Bundle:
			for name, field := range R.Fields {
				var err = CheckKmdInnerTypeSerializable(field.Type, p, mapping)
				if err != nil {
					return &KmdError {
						Point:    p,
						Concrete: E_KmdFieldNotSerializable { name },
					}
				}
			}
			return nil
		}
	case *NamedType:
		var _, ok = mapping[T.Name]
		if ok {
			return nil
		}
	}
	return &KmdError {
		Point:    p,
		Concrete: E_KmdTypeNotSerializable {},
	}
}