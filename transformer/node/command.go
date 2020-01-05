package node


type VariousCommand struct {
    Node                `part:"command"`
    Command   Command   `use:"first"`
}
type Command interface { Command() }

func (impl Import) Command() {}
type Import struct {
    Node                  `part:"import"`
    Name  Identifier      `part:"name"`
    Path  StringLiteral   `part:"string"`
}
type Identifier struct {
    Node           `part:"name"`
    Name [] rune   `content:"Name"`
}

func (impl DeclConst) Command() {}
type DeclConst struct {
    Node                `part:"decl_const"`
    Name   Identifier   `part:"name"`
    Value  Expr         `part:"expr"`
}

func (impl Do) Command() {}
type Do struct {
    Node `part:"do"`
    Effect  Expr   `part:"expr"`
}

func (impl DeclFunction) Command() {}
type DeclFunction struct {
    Node                      `part:"decl_func"`
    IsGlobal  bool            `option:"scope.@global"`
    Name      Identifier      `part:"name"`
    Params    [] Identifier   `list_more:"type_params" item:"name"`
    Repr      ReprFunc        `part:"repr_func"`
    Body      VariousBody     `part:"body"`
}
type VariousBody struct {
    Node         `part:"body"`
    Body  Body   `part:"lambda" fallback:"native"`
}
type Body interface { Body() }
func (impl NativeRef) Body() {}
type NativeRef struct {
    Node                `part:"native"`
    Id  StringLiteral   `part:"string"`
}

func (impl DeclType) Command() {}
type DeclType struct {
    Node                          `part:"decl_type"`
    IsOpaque   bool               `option:"opaque_opt.@opaque"`
    Name       Identifier         `part:"name"`
    Params     [] Identifier      `list_more:"type_params" item:"name"`
    TypeValue  VariousTypeValue   `part:"type_value"`
}
type VariousTypeValue struct {
    Node                   `part:"type_value"`
    TypeValue  TypeValue   `use:"first"`
}
type TypeValue interface { TypeValue() }
func (impl SingleType) TypeValue() {}
type SingleType struct {
    Node                `part:"single_type"`
    Repr  VariousRepr   `part:"repr"`
}
func (impl UnionType) TypeValue() {}
type UnionType struct {
    Node                 `part:"union_type"`
    Items  [] DeclType   `list_more:"" item:"decl_type"`
}
