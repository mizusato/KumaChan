package ast

import "kumachan/rpc/kmd"


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

func (impl Do) Statement() {}
type Do struct {
    Node   `part:"do"`
    Effect Expr `part:"expr"`
}

func (impl DeclConst) Statement() {}
type DeclConst struct {
    Node                          `part:"decl_const"`
    Docs      [] Doc              `list_rec:"docs"`
    Public    bool                `option:"scope.@public"`
    Name      Identifier          `part:"name"`
    Type      VariousType         `part:"type"`
    Value     VariousConstValue   `part_opt:"const_value"`
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
    Docs      [] Doc          `list_rec:"docs"`
    Tags      [] Tag          `list_rec:"tags"`
    Public    bool            `option:"scope.@public"`
    Name      Identifier      `part:"name"`
    Params    [] TypeParam    `list_more:"type_params" item:"type_param"`
    Implicit  [] VariousType  `list_more:"signature.implicit_input.type_args" item:"type"`
    Repr      ReprFunc        `part:"signature.repr_func"`
    Body      VariousBody     `part_opt:"body"`
}
type VariousBody struct {
    Node         `part:"body"`
    Body  Body   `use:"first"`
}
type Body interface { Body() }
func (impl KmdApiFuncBody) Body() {}
type KmdApiFuncBody struct {
    Id  kmd.TransformerPartId
}
func (impl ServiceMethodFuncBody) Body() {}
type ServiceMethodFuncBody struct {
    Name  string
}
func (impl ServiceCreateFuncBody) Body() {}
type ServiceCreateFuncBody struct {}

func (impl DeclType) Statement() {}
type DeclType struct {
    Node                        `part:"decl_type"`
    Docs       [] Doc           `list_rec:"docs"`
    Tags       [] Tag           `list_rec:"tags"`
    Name       Identifier       `part:"name"`
    Params     [] TypeParam     `list_more:"type_params" item:"type_param"`
    TypeDef    VariousTypeDef   `part:"type_def"`
}
type Doc struct {
    Node                  `part:"doc"`
    RawContent  [] rune   `content:"Doc"`
}
type Tag struct {
    Node                  `part:"tag"`
    RawContent  [] rune   `content:"Pragma"`
}
type TypeParam struct {
    Node                        `part:"type_param"`
    Name     Identifier         `part:"name"`
    Bound    VariousTypeBound   `part_opt:"type_bound"`
    Default  TypeParamDefault   `part_opt:"type_param_default"`
}
type TypeParamDefault struct {
    Node                    `part:"type_param_default"`
    HasValue  bool          `option:"type"`
    Value     VariousType   `part:"type"`
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
    Node               `part:"type_def"`
    TypeDef  TypeDef   `use:"first"`
}
type TypeDef interface { TypeDef() }
func (impl NativeType) TypeDef() {}
type NativeType struct {
    Node   `part:"t_native"`
}
func (impl EnumType) TypeDef() {}
type EnumType struct {
    Node                 `part:"t_enum"`
    Cases  [] DeclType   `list_more:"" item:"decl_type"`
}
func (impl ImplicitType) TypeDef() {}
type ImplicitType struct {
    Node               `part:"t_implicit"`
    Repr  ReprBundle   `part:"repr_bundle"`
}
func (impl BoxedType) TypeDef() {}
type BoxedType struct {
    Node                   `part:"t_boxed"`
    Protected  bool        `option:"box_option.@protected"`
    Opaque     bool        `option:"box_option.@opaque"`
    Weak       bool        `option:"match_option.@weak"`
    Inner      MaybeType   `part_opt:"inner_type.type"`
}

