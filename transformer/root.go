package transformer

import (
    . "kumachan/transformer/node"
)

// module_header = shebang module_metadata
func TransformModuleHeader (tree Tree) Module {
    var root_ptr = 0
    return Module {
        Node:      GetNode(tree, root_ptr, nil),
        FileName:  GetFileName(tree),
        MetaData:  module_metadata(tree, root_ptr),
        Decls:     nil,
        Commands:  nil,
    }
}

// module = shebang module_metadata imports decls commands
func TransformModule (tree Tree) Module {
    var root_ptr = 0
    return Module {
        Node:      GetNode(tree, root_ptr, nil),
        FileName:  GetFileName(tree),
        MetaData:  module_metadata(tree, root_ptr),
        Imports:   imports(tree, root_ptr),
        Decls:     decls(tree, root_ptr),
        Commands:  commands(tree, root_ptr),
    }
}

// module_metadata = export resolve
func module_metadata (tree Tree, parent Pointer) ModuleMetaData {
    var ptr = GetChildPointer(tree, parent)
    return ModuleMetaData {
        Node: GetNode(tree, ptr, nil),
        Exported: export(tree, ptr),
        Resolving: resolve(tree, ptr),
    }
}

// export? = @export { namelist! }! | @export namelist!
func export (tree Tree, parent Pointer) []Identifier {
    var ptr = GetChildPointer(tree, parent)
    return namelist(tree, ptr)
}

// namelist = name namelist_tail
func namelist (tree Tree, parent Pointer) []Identifier {
    var ptr = GetChildPointer(tree, parent)
    var name_ptrs = FlatSubTree(tree, ptr, "name", "namelist_tail")
    var names = make([]Identifier, len(name_ptrs))
    for i, name_ptr := range name_ptrs {
        names[i] = name(tree, name_ptr)
    }
    return names
}

// name = Name
func name (tree Tree, parent Pointer) Identifier {
    var ptr = GetChildPointer(tree, parent)
    return Identifier {
        Node: GetNode(tree, ptr, nil),
        Name: GetTokenContent(tree, ptr),
    }
}

// resolve? = @resolve { resolve_item more_resolve_items }!
func resolve (tree Tree, parent Pointer) []ModuleSource {
    var ptr = GetChildPointer(tree, parent)
    var item_ptrs = FlatSubTree (
        tree, ptr, "resolve_item", "more_resolve_items",
    )
    var items = make([]ModuleSource, len(item_ptrs))
    for i, item_ptr := range item_ptrs {
        items[i] = resolve_item(tree, item_ptr)
    }
    return items
}

// resolve_item = name =! resolve_detail String!
func resolve_item (tree Tree, parent Pointer) ModuleSource {
    var ptr = GetChildPointer(tree, parent)
    return ModuleSource {
        Node: GetNode(tree, ptr, nil),
        Alias: name(tree, ptr),
        URL: String(tree, ptr),
        Detail: resolve_detail(tree, ptr),
    }
}

// resolve_detail? = name @in! | ( name! mod_version )! @in!
func resolve_detail (tree Tree, parent Pointer) ModuleDetail {
    var ptr = GetChildPointer(tree, parent)
    var node = GetNode(tree, ptr, nil)
    if Empty(tree, ptr) {
        return ModuleDetail {
            Node: node,
            Name: NullIdentifier,
            Version: NullIdentifier,
        }
    } else {
        return ModuleDetail {
            Node: node,
            Name: name(tree, ptr),
            Version: mod_version(tree, ptr),
        }
    }
}

// mod_version? = , name!
func mod_version (tree Tree, parent Pointer) Identifier {
    var ptr = GetChildPointer(tree, parent)
    if Empty(tree, ptr) {
        return NullIdentifier
    } else {
        return name(tree, ptr)
    }
}

// imports? = import imports
func imports (tree Tree, parent Pointer) []Import {
    var ptr = GetChildPointer(tree, parent)
    if Empty(tree, ptr) {
        return []Import {}
    } else {
        var import_ptrs = FlatSubTree(tree, ptr, "import" , "imports")
        var result = make([]Import, len(import_ptrs))
        for i, import_ptr := range import_ptrs {
            result[i] = import_(tree, import_ptr)
        }
        return result
    }
}

// import = @import name ::! imported_names
func import_ (tree Tree, parent Pointer) Import {
    var ptr = GetChildPointer(tree, parent)
    return Import {
        Node: GetNode(tree, ptr, nil),
        FromModule: name(tree, ptr),
        Names: imported_names(tree, ptr),
    }
}

// imported_names = name | * | {! alias_list! }!
func imported_names (tree Tree, parent Pointer) []ImportedName {
    var ptr = GetChildPointer(tree, parent)
    if HasChild("name", tree, ptr) {
        var n = name(tree, ptr)
        return [] ImportedName {
            {
                Node: GetNode(tree, ptr, nil),
                Name: n,
                Alias: n,
            },
        }
    } else if HasChild("*", tree, ptr) {
        return [] ImportedName {}
    } else {
        return alias_list(tree, ptr)
    }
}

// alias_list = alias alias_list_tail
func alias_list (tree Tree, parent Pointer) []ImportedName {
    var ptr = GetChildPointer(tree, parent)
    var alias_ptrs = FlatSubTree(tree, ptr, "alias", "alias_list_tail")
    var result = make([]ImportedName, len(alias_ptrs))
    for i, alias_ptr := range alias_ptrs {
        result[i] = alias(tree, alias_ptr)
    }
    return result
}

// alias = name @as name! | name
func alias (tree Tree, parent Pointer) ImportedName {
    var ptr = GetChildPointer(tree, parent)
    if HasChild("@as", tree, ptr) {
        var first, last = FirstLastChild(tree, ptr)
        return ImportedName {
            Node: GetNode(tree, ptr, nil),
            Name: name(tree, first),
            Alias: name(tree, last),
        }
    } else {
        var n = name(tree, ptr)
        return ImportedName {
            Node: GetNode(tree, ptr, nil),
            Name: n,
            Alias: n,
        }
    }
}



