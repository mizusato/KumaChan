package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
)

type TypeRegistry map[loader.Symbol] *GenericType

type TypeExprError interface { TypeExprError() }

func (impl E_ModuleOfTypeRefNotFound) TypeExprError() {}
type E_ModuleOfTypeRefNotFound struct {
	Name  string
	Node  node.Node
}
func (impl E_TypeNotFound) TypeExprError() {}
type E_TypeNotFound struct {
	Name  loader.Symbol
	Node  node.Node
}
func (impl E_NativeTypeNotFound) TypeExprError() {}
type E_NativeTypeNotFound struct {
	Name  string
	Node  node.Node
}
func (impl E_WrongArgumentQuantity) TypeExprError() {}
type E_WrongArgumentQuantity struct {
	TypeName  loader.Symbol
	Required  uint
	Given     uint
	Node      node.Node
}
func (impl E_DuplicateFields) TypeExprError() {}
type E_DuplicateFields struct {
	FieldName  string
	Node       node.Node
}

type TypeDeclError interface { TypeDeclError() }

func (impl E_DuplicateTypeDecl) TypeDeclError() {}
type E_DuplicateTypeDecl struct {
	TypeName  loader.Symbol
	Node      node.Node
}
func (impl E_GenericUnionSubType) TypeDeclError() {}
type E_GenericUnionSubType struct {
	TypeName  loader.Symbol
	Node      node.Node
}
func (impl E_InvalidTypeDecl) TypeDeclError() {}
type E_InvalidTypeDecl struct {
	TypeName   loader.Symbol
	Decl       node.Node
	ExprError  TypeExprError
}

type RawTypeRegistry struct {
	DeclMap       map[loader.Symbol] node.DeclType
	UnionRootMap  map[loader.Symbol] loader.Symbol
}

func MakeRawTypeRegistry() RawTypeRegistry {
	return RawTypeRegistry {
		DeclMap: make(map[loader.Symbol] node.DeclType),
		UnionRootMap: make(map[loader.Symbol] loader.Symbol),
	}
}

func (raw RawTypeRegistry) LookupArity(name loader.Symbol) (bool, uint) {
	var t, exists = raw.DeclMap[name]
	if exists {
		var ur_name, exists = raw.UnionRootMap[name]
		if exists {
			var ur = raw.DeclMap[ur_name]  // the value must exist, thus omit checking
			return true, uint(len(ur.Params))
		} else {
			return true, uint(len(t.Params))
		}
	} else {
		return false, 0
	}
}

func RegisterRawTypes (mod *loader.Module, raw RawTypeRegistry) TypeDeclError {
	var decls = make([]node.DeclType, 0)
	var root_map = make(map[int]int, 0)
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case node.DeclType:
			var cur_union_index = len(decls)
			decls = append(decls, c)
			var root_of_union, root_of_union_exists = root_map[cur_union_index]
			switch u := c.TypeValue.TypeValue.(type) {
			case node.UnionType:
				for _, item := range u.Items {
					var cur_sub_index = len(decls)
					decls = append(decls, item)
					if root_of_union_exists {
						root_map[cur_sub_index] = root_of_union
					} else {
						root_map[cur_sub_index] = cur_union_index
					}
				}
			}
		}
	}
	for i, d := range decls {
		var type_sym = mod.SymbolFromName(d.Name)
		var _, exists = raw.DeclMap[type_sym]
		if exists {
			return E_DuplicateTypeDecl {
				TypeName:  type_sym,
				Node:      d.Node,
			}
		} else {
			raw.DeclMap[type_sym] = d
			var root, exists = root_map[i]
			if exists {
				if len(d.Params) > 0 {
					return E_GenericUnionSubType {
						TypeName: type_sym,
						Node:     d.Name.Node,
					}
				}
				raw.UnionRootMap[type_sym] = mod.SymbolFromName(decls[root].Name)
			}
		}
	}
	for _, imported := range mod.ImpMap {
		var err = RegisterRawTypes(imported, raw)
		if err != nil {
			return err
		}
	}
	return nil
}

type TypeExprContext interface { TypeExprContext() }

func (impl TypeDeclContext) TypeExprContext() {}
type TypeDeclContext struct {
	Module  *loader.Module
	RawReg  RawTypeRegistry
	Params  [] string
}

func RegisterTypes (mod *loader.Module) (TypeRegistry, TypeDeclError) {
	var raw = MakeRawTypeRegistry()
	var err = RegisterRawTypes(mod, raw)
	if err != nil { return nil, err }
	var reg = make(TypeRegistry)
	for name, t := range raw.DeclMap {
		var params = make([]string, len(t.Params))
		var params_t = t
		var root, exists = raw.UnionRootMap[name]
		if exists {
			params_t = raw.DeclMap[root]
		}
		for i, param := range params_t.Params {
			params[i] = loader.Id2String(param)
		}
		var val, err = TypeValFrom(t.TypeValue.TypeValue, TypeDeclContext {
			Module: mod,
			RawReg: raw,
			Params: params,
		})
		if err != nil { return nil, E_InvalidTypeDecl {
			TypeName:  name,
			Decl:      t.Node,
			ExprError: err,
		} }
		reg[name] = &GenericType{
			Arity:     uint(len(params)),
			IsOpaque:  t.IsOpaque,
			Value:     val,
			Node:      t.Node,
		}
	}
	return reg, nil
}

func TypeValFrom (tv node.TypeValue, ctx TypeDeclContext) (TypeVal, TypeExprError) {
	switch v := tv.(type) {
	case node.UnionType:
		var subtypes = make([]loader.Symbol, len(v.Items))
		for i, item := range v.Items {
			subtypes[i] = ctx.Module.SymbolFromName(item.Name)
		}
		return UnionTypeVal {
			SubTypes: subtypes,
		}, nil
	case node.SingleType:
		var expr, err = TypeExprFromRepr(v.Repr.Repr, ctx)
		if err != nil { return nil, err }
		return SingleTypeVal {
			Expr: expr,
		}, nil
	default:
		panic("impossible branch")
	}
}

func TypeExprFrom (type_ node.Type, ctx TypeExprContext) (TypeExpr, TypeExprError) {
	switch t := type_.(type) {
	case node.TypeRef:
		switch c := ctx.(type) {
		case TypeDeclContext:
			var ref_mod = string(t.Ref.Module.Name)
			var ref_name = string(t.Ref.Id.Name)
			if ref_mod == "" {
				for i, param := range c.Params {
					if param == ref_name {
						return ParameterType { Index: uint(i) }, nil
					}
				}
			}
			var sym = c.Module.SymbolFromRef(t.Ref)
			switch s := sym.(type) {
			case loader.Symbol:
				var exists, arity = c.RawReg.LookupArity(s)
				if !exists { return nil, E_TypeNotFound {
					Name: s,
					Node: t.Ref.Node,
				}}
				var given_arity = uint(len(t.Ref.TypeArgs))
				if arity != given_arity { return nil, E_WrongArgumentQuantity {
					TypeName: s,
					Required: arity,
					Given:    given_arity,
					Node:     node.Node{},
				} }
				var args = make([]TypeExpr, arity)
				for i, arg_node := range t.Ref.TypeArgs {
					var arg, err = TypeExprFrom(arg_node.Type, ctx)
					if err != nil { return nil, err }
					args[i] = arg
				}
				return NamedType {
					Name: s,
					Args: args,
				}, nil
			default:
				return nil, E_ModuleOfTypeRefNotFound {
					Name: loader.Id2String(t.Ref.Module),
					Node: t.Ref.Module.Node,
				}
			}
		default:
			panic("unsupported branch")
		}
	case node.TypeLiteral:
		return TypeExprFromRepr(t.Repr.Repr, ctx)
	default:
		panic("impossible branch")
	}
}

func TypeExprFromRepr (repr node.Repr, ctx TypeExprContext) (TypeExpr, TypeExprError) {
	switch r := repr.(type) {
	case node.ReprTuple:
		if len(r.Elements) == 0 {
			return AnonymousType {
				Repr: Nil {},
			}, nil
		} else {
			var elements = make([]TypeExpr, len(r.Elements))
			for i, el := range r.Elements {
				var e, err = TypeExprFrom(el.Type, ctx)
				if err != nil { return nil, err }
				elements[i] = e
			}
			if len(elements) == 1 {
				return elements[0], nil
			} else {
				return AnonymousType {
					Repr: Tuple { Elements:elements },
				}, nil
			}
		}
	case node.ReprBundle:
		if len(r.Fields) == 0 {
			return AnonymousType {
				Repr: Nil {},
			}, nil
		} else {
			var fields = make(map[string]TypeExpr)
			for _, f := range r.Fields {
				var f_name = loader.Id2String(f.Name)
				var _, exists = fields[f_name]
				if exists { return nil, E_DuplicateFields {
					FieldName: f_name,
					Node:      f.Node,
				} }
				var f_type, err = TypeExprFrom(f.Type.Type, ctx)
				if err != nil { return nil, err }
				fields[f_name] = f_type
			}
			return AnonymousType {
				Repr: Bundle { Fields: fields },
			}, nil
		}
	case node.ReprFunc:
		var input, err1 = TypeExprFrom(r.Input.Type, ctx)
		if err1 != nil { return nil, err1 }
		var output, err2 = TypeExprFrom(r.Output.Type, ctx)
		if err2 != nil { return nil, err2 }
		return AnonymousType {
			Repr: Func {
				Input:  input,
				Output: output,
			},
		}, nil
	case node.ReprNative:
		var str_id = string(r.Ref.Id.Value)
		var id, exists = NativeTypes[str_id]
		if !exists { return nil, E_NativeTypeNotFound {
			Name: str_id,
			Node: r.Node,
		} }
		return AnonymousType {
			Repr: NativeType { Id: id },
		}, nil
	default:
		panic("impossible branch")
	}
}