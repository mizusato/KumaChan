package node

import "kumachan/parser/scanner"


type Node struct {
    Point  scanner.Point
    Span   scanner.Span
    Info   interface{}
}

type Declaration interface { Declaration() }

// Section: used to group declarations
func (impl Section) Declaration() {}
type Section struct {
    Node
    Name   string
    Decls  [] Declaration
}

// Block: used in function declarations, lambda and guard expressions
type Block struct {
    Node
    Imports   [] Import
    Commands  [] Command
}

// Identifier: used by various nodes
type Identifier struct {
    Node
    Name string
}

var NullIdentifier = Identifier { Name: "" }