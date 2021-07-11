package ast

import (
    "strings"
    "kumachan/standalone/rpc/kmd"
)


type VariousStatement struct {
    Node                   `part:"stmt"`
    Statement  Statement   `use:"first"`
}
type Statement interface { Statement() }

func (impl Title) Statement() {}
type Title struct {
    Node               `part:"title"`
    Content  [] rune   `content:"Title"`
}

func (impl Import) Statement() {}
type Import struct {
    Node               `part:"import"`
    Name  Identifier   `part:"name"`
    Path  StringText   `part:"string_text"`
}

func (impl Do) Statement() {}
type Do struct {
    Node           `part:"do"`
    Effect  Expr   `part:"expr"`
}

func (impl Alias) Statement() {}
type Alias struct {
    Node                 `part:"alias"`
    Name    Identifier   `part:"name"`
    Module  Identifier   `part_opt:"alias_target.module_prefix.name"`
    Item    Identifier   `part:"alias_target.name"`
}

func (impl DeclConst) Statement() {}
type DeclConst struct {
    Node                          `part:"decl_const"`
    Docs      [] Doc              `list_rec:"docs"`
    Meta      Meta                `part:"meta"`
    Public    bool                `option:"scope.@export"`
    Name      Identifier          `part:"name"`
    Type      VariousType         `part:"type"`
    Value     VariousConstValue   `part_opt:"const_def.const_value"`
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

func (DeclMethod) Statement() {}
type DeclMethod struct {
    Node                    `part:"decl_method"`
    Docs      [] Doc        `list_rec:"docs"`
    Meta      Meta          `part:"meta"`
    Public    bool          `option:"scope.@export"`
    Name      Identifier    `part:"name"`
    Receiver  Identifier    `part:"receiver.name"`
    Type      VariousType   `part:"type"`
    Body      VariousBody   `part:"body"`
}

func (impl DeclFunction) Statement() {}
type DeclFunction struct {
    Node                         `part:"decl_func"`
    Docs      [] Doc             `list_rec:"docs"`
    Meta      Meta               `part:"meta"`
    Public    bool               `option:"scope.@export"`
    Name      Identifier         `part:"name"`
    Params    [] FuncTypeParam   `list_more:"sig.func_type_params" item:"func_type_param"`
    Implicit  ReprRecord         `part_opt:"sig.implicit.repr_record"`
    InOut     ReprFunc           `part:"sig.repr_func"`
    Body      VariousBody        `part_opt:"body"`
    Kind      FuncKind
}
type FuncKind int

const (
    _            FuncKind  =  iota
    FK_Constant  // desugared from a ConstDecl
    FK_Method    // desugared from a MethodDecl
)
type FuncTypeParam struct {
    Node                        `part:"func_type_param"`
    Name     Identifier         `part:"name"`
    Bound    VariousTypeBound   `part_opt:"type_bound"`
}
type VariousTypeBound struct {
    Node           `part:"bound"`
    Bound  Bound   `use:"first"`
}
type Bound interface { Bound() }
func (impl TypeLowerBound) Bound() {}
type TypeLowerBound struct {
    Node                 `part:"lo_bound"`
    Value  VariousType   `part:"type"`
}
func (impl TypeHigherBound) Bound() {}
type TypeHigherBound struct {
    Node                 `part:"hi_bound"`
    Value  VariousType   `part:"type"`
}
type VariousBody struct {
    Node         `part:"body"`
    Body  Body   `use:"first"`
}
type Body interface { Body() }
func (impl PredefinedThunk) Body() {}
type PredefinedThunk struct {
    Value  interface {}
}
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
    Meta       Meta             `part:"meta"`
    Name       Identifier       `part:"name"`
    Params     [] TypeParam     `list_more:"type_params" item:"type_param"`
    Impl       [] TypeDeclRef   `list_more:"impl" item:"type_decl_ref"`
    TypeDef    MaybeTypeDef     `part_opt:"type_def"`
}
type Doc struct {
    Node                  `part:"doc"`
    RawContent  [] rune   `content:"Doc"`
}
type Meta struct {
    Node                 `part:"meta"`
    Items  [] MetaItem   `list_rec:"meta_items"`
}
type MetaItem struct {
    Node                  `part:"meta_item"`
    RawContent  [] rune   `content:"Meta"`
}
type TypeParam struct {
    Node                        `part:"type_param"`
    Name     Identifier         `part:"name"`
    In       bool               `option:"variance.@in"`
    Out      bool               `option:"variance.@out"`
    Default  MaybeType          `part_opt:"default_type.type"`
}
type TypeDeclRef struct {
    Node                 `part:"type_decl_ref"`
    Module  Identifier   `part_opt:"module_prefix.name"`
    Item    Identifier   `part:"name"`
}
type MaybeTypeDef interface { Maybe(VariousTypeDef, MaybeTypeDef) }
func (impl VariousTypeDef) Maybe(VariousTypeDef, MaybeTypeDef) {}
type VariousTypeDef struct {
    Node               `part:"type_def"`
    TypeDef  TypeDef   `use:"first"`
}
type TypeDef interface { TypeDef() }
func (impl NativeType) TypeDef() {}
type NativeType struct {
    Node `part:"t_native"`
}
func (impl EnumType) TypeDef() {}
type EnumType struct {
    Node                 `part:"t_enum"`
    Cases  [] DeclType   `list_more:"" item:"decl_type"`
}
func (impl InterfaceType) TypeDef() {}
type InterfaceType struct {
    Node                  `part:"t_interface"`
    Methods  ReprRecord   `part:"repr_record"`
}
func (impl BoxedType) TypeDef() {}
type BoxedType struct {
    Node                     `part:"t_boxed"`
    Protected  bool          `option:"box_option.@protected"`
    Opaque     bool          `option:"box_option.@opaque"`
    Weak       bool          `option:"match_option.@weak"`
    Inner      VariousType   `part:"inner_type.type"`
}

func GetDocContent(raw ([] Doc)) string {
    var buf strings.Builder
    for _, line := range raw {
        var t = string(line.RawContent)
        t = strings.TrimPrefix(t, "///")
        t = strings.TrimPrefix(t, " ")
        t = strings.TrimRight(t, " \r")
        buf.WriteString(t)
        buf.WriteRune('\n')
    }
    return buf.String()
}

func GetMetadataContent(meta Meta) string {
    var buf strings.Builder
    for _, item := range meta.Items {
        var t = string(item.RawContent)
        t = strings.TrimSuffix(t, "\r")
        t = strings.TrimPrefix(t, "#")
        t = strings.Trim(t, " \t")
        buf.WriteString(t)
        buf.WriteRune('\n')
    }
    return buf.String()
}


