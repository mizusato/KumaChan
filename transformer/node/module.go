package node

// module = header export imports decls commands
// header = shebang resolve
// export? = @export { namelist! }! | @export namelist!
type Module struct {
    Node                       `part:"module"`
    Resolved  [] Resolve       `list_more:"header.resolve" item:"resolve_item"`
    Exported  [] Identifier    `list:"export.namelist"`
    Imports   [] Import        `list_rec:"imports"`
    Decls     [] Declaration   `list_rec:"decls"`
    Commands  [] Command       `list_rec:"commands"`
}

// resolve_item = name =! version string!
// version? = name @of!
type Resolve struct {
    Node                     `part:"resolve_item"`
    Name     Identifier      `part:"name"`
    Version  Identifier      `part_opt:"version.name"`
    Source   StringLiteral   `part:"string"`
}

// import = import_from import_names | import_from alias
// import_from = @import name ::!
// import_names = * | {! alias_list! }!
// alias_list = alias alias_list_tail
type Import struct {
    Node                          `part:"import"`
    FromModule  Identifier        `part:"import_from.name"`
    Names       [] ImportedName   `list:"import_names.alias_list" fallback:"alias"`
}

// alias = name @as alias_name | name
type ImportedName struct {
    Node                `part:"alias"`
    Name   Identifier   `part:"name"`
    Alias  Identifier   `part:"alias_name.name" fallback:"name"`
}
