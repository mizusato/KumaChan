package node


type Declaration interface { Declaration() }

// Section: used to group declarations
func (impl Section) Declaration() {}
type Section struct {
    Node
    Name   Identifier
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
    Node           `part:"name"`
    Name [] rune   `content:"Name"`
}

var NullIdentifier = Identifier { Name: []rune{} }
