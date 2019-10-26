package node


type Module struct {
    Node
    FileName  string
    MetaData  ModuleMetaData
    Imports   [] Import
    Decls     [] Declaration
    Commands  [] Command
}

type ModuleMetaData struct {
    Node
    Exported    [] Identifier
    Resolving   [] ModuleSource
}

type ModuleSource struct {
    Node
    Alias      Identifier
    URL        StringLiteral
    Detail     ModuleDetail
}

type ModuleDetail struct {
    Node
    Version    Identifier
    Name       Identifier
}

type Import struct {
    Node
    FromModule  Identifier
    Names       [] ImportedName
}

type ImportedName struct {
    Node
    Name   Identifier
    Alias  Identifier
}

