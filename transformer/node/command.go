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
type MaybeIdentifier interface { MaybeIdentifier() }
func (impl Identifier) MaybeIdentifier() {}
type Identifier struct {
    Node           `part:"name"`
    Name [] rune   `content:"Name"`
}

func (impl DeclConst) Command() {}
type DeclConst struct {
    Node                          `part:"decl_const"`
    IsPublic  bool                `option:"scope.@public"`
    Name      Identifier          `part:"name"`
    Type      VariousType         `part:"type"`
    Value     VariousConstValue   `part:"const_value"`
}
type VariousConstValue struct {
    Node                     `part:"const_value"`
    ConstValue  ConstValue   `use:"first"`
}
type ConstValue interface { ConstValue() }
func (impl NativeRef) ConstValue() {}
func (impl NativeRef) Body() {}
type NativeRef struct {
    Node                `part:"native"`
    Id  StringLiteral   `part:"string"`
}

func (impl Do) Command() {}
type Do struct {
    Node           `part:"do"`
    Effect  Expr   `part:"expr"`
}

func (impl DeclFunction) Command() {}
type DeclFunction struct {
    Node                    `part:"decl_func"`
    Public  bool            `option:"scope.@public"`
    Name    Identifier      `part:"name"`
    Params  [] Identifier   `list_more:"type_params" item:"name"`
    Repr    ReprFunc        `part:"signature.repr_func"`
    Body    VariousBody     `part:"body"`
}
type VariousBody struct {
    Node         `part:"body"`
    Body  Body   `part:"lambda" fallback:"native"`
}
type Body interface { Body() }

func (impl DeclType) Command() {}
type DeclType struct {
    Node                          `part:"decl_type"`
    Name       Identifier         `part:"name"`
    Params     [] Identifier      `list_more:"type_params" item:"name"`
    TypeValue  VariousTypeValue   `part:"type_value"`
}
type VariousTypeValue struct {
    Node                   `part:"type_value"`
    TypeValue  TypeValue   `use:"first"`
}
type TypeValue interface { TypeValue() }
func (impl NativeType) TypeValue() {}
type NativeType struct {
    Node   `part:"native_type"`
}
func (impl BoxedType) TypeValue() {}
type BoxedType struct {
    Node                     `part:"boxed_type"`
    Protected  bool          `option:"box_option.@protected"`
    Opaque     bool          `option:"box_option.@opaque"`
    Inner      MaybeType     `part_opt:"inner_type.type"`
}
func (impl UnionType) TypeValue() {}
type UnionType struct {
    Node                 `part:"union_type"`
    Items  [] DeclType   `list_more:"" item:"decl_type"`
}
