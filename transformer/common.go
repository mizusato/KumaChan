package transformer

import ."kumachan/transformer/node"

// decls? = decl decls
func decls (tree Tree, parent Pointer) []Declaration {
    // TODO
    return []Declaration{}
}

// section = @section name { decls }!
func section (tree Tree, parent Pointer) Section {
    var ptr = GetChildPointer(tree, parent)
    return Section {
        Node: GetNode(tree, ptr, nil),
        Name: name(tree, ptr),
        Decls: decls(tree, ptr),
    }
}

// commands? = command commands
func commands (tree Tree, parent Pointer) []Command {
    var ptr = GetChildPointer(tree, parent)
    if Empty(tree, ptr) {
        return []Command {}
    } else {
        // TODO
        return []Command {}
    }
}

// block = { imports commands }!
func block (tree Tree, parent Pointer) Block {
    var ptr = GetChildPointer(tree, parent)
    return Block {
        Node: GetNode(tree, ptr, nil),
        Imports: imports(tree, ptr),
        Commands: commands(tree, ptr),
    }
}