package node


type TypeDeclaration interface {
    Declaration
    TypeDeclaration()
}

func (impl SingletonTypes) Declaration() {}
func (impl SingletonTypes) TypeDeclaration() {}
type SingletonTypes struct {
    Node
    Names  [] string
}

func (impl UnionType) Declaration() {}
func (impl UnionType) TypeDeclaration() {}
type UnionType struct {
    Node
    Name      string
    TP        [] TypeParameter
    Elements  [] TypeExpr
}

func (impl SchemaType) Declaration() {}
func (impl SchemaType) TypeDeclaration() {}
type SchemaType struct {
    Node
    Name    string
    Attrs   SchemaAttrs
    TP      [] TypeParameter
    Bases   [] TypeExpr
    Fields  [] SchemaField
}

func (impl ClassType) Declaration() {}
func (impl ClassType) TypeDeclaration() {}
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

func (impl InterfaceType) Declaration() {}
func (impl InterfaceType) TypeDeclaration() {}
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
