package node


type Module struct {
    Node
    MetaData  ModuleMetaData
    Imports   [] Import
}

type ModuleMetaData struct {
    Node
    Shebang     string
    Exported    map[string] string
    Resolving   map[string] ModuleSource
}

type ModuleSource struct {
    Node
    IsBuiltIn  bool
    Version    string
    URL        string
}

type Import struct {
    Node
    FromModule  string
    Names       [] ImportedName
}

type ImportedName struct {
    Node
    Name   string
    Alias  string
}

