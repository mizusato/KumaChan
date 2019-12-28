package node


type VariousCommand struct {
    Node               `part:"command"`
    Content  Command   `use:"first"`
}
type Command interface { Command() }

func (impl Import) Command() {}
type Import struct {
    Node                  `part:"import"`
    Name  Identifier      `part:"name"`
    Path  StringLiteral   `part:"string"`
}


func (impl DeclFunction) Command() {}
type DeclFunction struct {
    Node `part:"decl_func"`
}

func (impl DeclConst) Command() {}
type DeclConst struct {
    Node `part:"decl_const"`
}

func (impl Do) Command() {}
type Do struct {
    Node `part:"do"`
}

func (impl DeclType) Command() {}
type DeclType struct {
    Node                    `part:"decl_type"`
    Options   TypeOptions   `part:"type_opts"`
    TypeDecl  TypeDecl      `part:"single_type" fallback:"union_type"`
}

type TypeDecl interface { TypeDecl() }

type SingleType struct {
    Node                  `part:"single_type"`
    Name  Identifier      `part:"name"`
    Para  [] Identifier   `list_more:"type_params" item:"name"`
    Repr  VariousRepr     `part:"repr"`
}

type UnionType struct {
    Node                   `part:"union_type"`
    Name   Identifier      `part:"name"`
    Para   [] Identifier   `list_more:"type_params" item:"name"`
    Items  [] DeclType     `list_more:"" item:"decl_type"`
}

type TypeOptions struct {
    Node               `part:"type_opts"`
    IsExported  bool   `option:"export_opt.@export"`
    IsOpaque    bool   `option:"opaque_opt.@opaque"`
}