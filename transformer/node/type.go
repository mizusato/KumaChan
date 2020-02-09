package node


type MaybeType interface { MaybeType() }
func (impl VariousType) MaybeType() {}
type VariousType struct {
    Node         `part:"type"`
    Type  Type   `use:"first"`
}
type Type interface { Type() }

func (impl TypeRef) Type() {}
type TypeRef struct {
    Node        `part:"type_ref"`
    Ref   Ref   `part:"ref"`
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

func (impl ReprBundle) Repr() {}
type ReprBundle struct {
    Node               `part:"repr_bundle"`
    Fields  [] Field   `list_more:"field_list" item:"field"`
}
type Field struct {
    Node                `part:"field"`
    Name  Identifier    `part:"name"`
    Type  VariousType   `part:"type"`
}

func (impl ReprFunc) Repr() {}
type ReprFunc struct {
    Node                  `part:"repr_func"`
    Input   VariousType   `part:"input_type.type"`
    Output  VariousType   `part:"output_type.type"`
}
