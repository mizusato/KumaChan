package node

// decl_type = singleton | union | schema | class | interface
func (impl TypeDeclaration) DeclContent() {}
type TypeDeclaration struct {
    Node                       `part:"decl_type"`
    Content  TypeDeclContent   `use:"first"`
}
type TypeDeclContent interface { TypeDeclContent() }

// singleton = @singleton namelist
func (impl SingletonTypes) TypeDeclContent() {}
type SingletonTypes struct {
    Node                   `part:"singleton"`
    Names  [] Identifier   `list:"namelist"`
}

func (impl UnionType) TypeDeclContent() {}
type UnionType struct {
    Node
    Name      string
    TP        [] TypeParameter
    Elements  [] TypeExpr
}

func (impl SchemaType) TypeDeclContent() {}
type SchemaType struct {
    Node
    Name    string
    Attrs   SchemaAttrs
    TP      [] TypeParameter
    Bases   [] TypeExpr
    Fields  [] SchemaField
}

func (impl ClassType) TypeDeclContent() {}
type ClassType struct {
    Node
    Name     string
    Attrs    ClassAttrs
    TP       [] TypeParameter
    Bases    [] TypeExpr
    Impls    [] TypeExpr
    Init     ClassInit
    PF       map[string] FunctionItem
    Methods  map[string] FunctionItem
}

func (impl InterfaceType) TypeDeclContent() {}
type InterfaceType struct {
    Node
    Name     string
    Attrs    InterfaceAttrs
    TP       [] TypeParameter
    Methods  map[string] MethodSignature
}


type SchemaAttrs struct {
    Mutable     bool
    Extensible  bool
}

type SchemaField struct {
    Node
    Name    string
    Type    TypeExpr
    Default Expr
}

type ClassAttrs struct {
    Extensible  bool
}

type ClassInit struct {
    Node
    Parameters  [] Parameter
    Body        FunctionBody
}

type InterfaceAttrs struct {
    IsNative  bool
}

type MethodSignature struct {
    Node
    Parameters  [] Parameter
    ReturnValue TypeExpr
}
