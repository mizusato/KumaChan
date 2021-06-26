package ast


type MaybeType interface { Maybe(Type, MaybeType) }
func (impl VariousType) Maybe(Type, MaybeType) {}
type VariousType struct {
    Node         `part:"type"`
    Type  Type   `use:"first"`
}
type Type interface { Type() }

func (impl TypeBlank) Type() {}
type TypeBlank struct {
    Node `part:"type_blank"`
}

type MaybeTypeRef interface { Maybe(TypeRef, MaybeTypeRef) }
func (impl TypeRef) Maybe(TypeRef, MaybeTypeRef) {}
func (impl TypeRef) Type()  {}
type TypeRef struct {
    Node                       `part:"type_ref"`
    Module    Identifier       `part_opt:"module_prefix.name"`
    Item      Identifier       `part:"name"`
    TypeArgs  [] VariousType   `list_more:"type_args" item:"type"`
}

func (impl TypeLiteral) Type()  {}
type TypeLiteral struct {
    Node                `part:"type_literal"`
    Repr  VariousRepr   `part:"repr"`
}
type VariousRepr struct {
    Node         `part:"repr"`
    Repr  Repr   `use:"first"`
}
type Repr interface { Repr() }


func (impl ReprTuple) Repr() {}
type ReprTuple struct {
    Node                       `part:"repr_tuple"`
    Elements  [] VariousType   `list_more:"type_list" item:"type"`
}

func (impl ReprRecord) Repr() {}
type ReprRecord struct {
    Node               `part:"repr_record"`
    Fields  [] Field   `list_more:"field_list" item:"field"`
}
type Field struct {
    Node                `part:"field"`
    Docs  [] Doc        `list_rec:"docs"`
    Meta  Meta          `part:"meta"`
    Name  Identifier    `part:"name"`
    Type  VariousType   `part:"type"`
}

func (impl ReprFunc) Repr() {}
type ReprFunc struct {
    Node                  `part:"repr_func"`
    Input   VariousType   `part:"input_type.type"`
    Output  VariousType   `part:"output_type.type"`
}

