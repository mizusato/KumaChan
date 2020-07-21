package ast


type VariousStatement struct {
    Node                   `part:"stmt"`
    Statement  Statement   `use:"first"`
}
type Statement interface { Statement() }

func (impl Import) Statement() {}
type Import struct {
    Node               `part:"import"`
    Name  Identifier   `part:"name"`
    Path  StringText   `part:"string_text"`
}
type MaybeIdentifier interface { MaybeIdentifier() }
func (impl Identifier) MaybeIdentifier() {}
type Identifier struct {
    Node           `part:"name"`
    Name [] rune   `content:"Name"`
}

func (impl Do) Statement() {}
type Do struct {
    Node   `part:"do"`
    Effect Expr `part:"expr"`
}

func (impl DeclConst) Statement() {}
type DeclConst struct {
    Node                          `part:"decl_const"`
    Public    bool                `option:"scope.@public"`
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
    Node               `part:"native"`
    Id    StringText   `part:"string_text"`
}
func (impl PredefinedValue) ConstValue() {}
type PredefinedValue struct {
    Value  interface {}
}

func (impl DeclFunction) Statement() {}
type DeclFunction struct {
    Node                      `part:"decl_func"`
    Public    bool            `option:"scope.@public"`
    Name      Identifier      `part:"name"`
    Params    [] Identifier   `list_more:"type_params.namelist" item:"name"`
    Implicit  [] VariousType  `list_more:"signature.implicit_input.type_args" item:"type"`
    Repr      ReprFunc        `part:"signature.repr_func"`
    Body      VariousBody     `part:"body"`
}
type VariousBody struct {
    Node         `part:"body"`
    Body  Body   `part:"lambda" fallback:"native"`
}
type Body interface { Body() }

func (impl DeclType) Statement() {}
type DeclType struct {
    Node                          `part:"decl_type"`
    Name       Identifier         `part:"name"`
    Params     [] Identifier      `list_more:"type_params.namelist" item:"name"`
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
func (impl ImplicitType) TypeValue() {}
type ImplicitType struct {
    Node               `part:"implicit_type"`
    Repr  ReprBundle   `part:"repr_bundle"`
}
func (impl BoxedType) TypeValue() {}
type BoxedType struct {
    Node                     `part:"boxed_type"`
    AsIs       bool          `option:"box_option.@as"`
    Protected  bool          `option:"box_option.@protected"`
    Opaque     bool          `option:"box_option.@opaque"`
    Inner      MaybeType     `part_opt:"inner_type.type"`
}
func (impl UnionType) TypeValue() {}
type UnionType struct {
    Node                 `part:"union_type"`
    Cases  [] DeclType   `list_more:"" item:"decl_type"`
}
