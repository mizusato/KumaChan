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
type MaybeIdentifier interface { Maybe(Identifier,MaybeIdentifier) }
func (impl Identifier) Maybe(Identifier,MaybeIdentifier) {}
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
    Params    [] TypeParam    `list_more:"type_params" item:"type_param"`
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
    Node                        `part:"decl_type"`
    Tags       [] TypeTag       `list_rec:"tags"`
    Name       Identifier       `part:"name"`
    Params     [] TypeParam     `list_more:"type_params" item:"type_param"`
    TypeValue  VariousTypeDef   `part:"type_def"`
}
type TypeTag struct {
    Node                  `part:"tag"`
    RawContent  [] rune   `content:"Pragma"`
}
type TypeParam struct {
    Node                        `part:"type_param"`
    Name    Identifier          `part:"name"`
    Bound   VariousTypeBound    `part_opt:"type_bound"`
}
type VariousTypeBound struct {
    Node                   `part:"type_bound"`
    TypeBound  TypeBound   `use:"first"`
}
type TypeBound interface { TypeBound() }
func (impl TypeLowerBound) TypeBound() {}
type TypeLowerBound struct {
    Node                     `part:"type_lo_bound"`
    BoundType  VariousType   `part:"type"`
}
func (impl TypeHigherBound) TypeBound() {}
type TypeHigherBound struct {
    Node                     `part:"type_hi_bound"`
    BoundType  VariousType   `part:"type"`
}
type VariousTypeDef struct {
    Node                 `part:"type_def"`
    TypeValue  TypeDef   `use:"first"`
}
type TypeDef interface { TypeDef() }
func (impl NativeType) TypeDef() {}
type NativeType struct {
    Node   `part:"t_native"`
}
func (impl UnionType) TypeDef() {}
type UnionType struct {
    Node                 `part:"t_union"`
    Cases  [] DeclType   `list_more:"" item:"decl_type"`
}
func (impl InterfaceType) TypeDef() {}
type InterfaceType struct {
    Node                   `part:"t_interface"`
    Params  [] TypeParam   `list_more:"interface_params" item:"type_param"`
    Inner   MaybeType      `part:"interface_inner_type.type"`
}
func (impl ImplicitType) TypeDef() {}
type ImplicitType struct {
    Node               `part:"t_implicit"`
    Repr  ReprBundle   `part:"repr_bundle"`
}
func (impl BoxedType) TypeDef() {}
type BoxedType struct {
    Node                   `part:"t_boxed"`
    Weak       bool        `option:"box_option.@weak"`
    Protected  bool        `option:"box_option.@protected"`
    Opaque     bool        `option:"box_option.@opaque"`
    Inner      MaybeType   `part_opt:"inner_type.type"`
}
