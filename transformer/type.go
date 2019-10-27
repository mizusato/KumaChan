package transformer

import ."kumachan/transformer/node"

// type = type_ordinary | type_attached | type_trait | type_misc
func type_ (tree Tree, parent Pointer) TypeExpr {
    var ptr = GetChildPointer(tree, parent)
    return TypeExpr {
        Node: GetNode(tree, ptr, nil),

    }
}

// type_ordinary = module_prefix name type_args
func type_ordinary (tree Tree, parent Pointer) OrdinaryTypeExpr {
    var ptr = GetChildPointer(tree, parent)
    return OrdinaryTypeExpr {
        Node: GetNode(tree, ptr, nil),
        Module: module_prefix(tree, ptr),
        Name: name(tree, ptr),
        Args: type_args(tree, ptr),
    }
}

// module_prefix? = name ::
func module_prefix (tree Tree, parent Pointer) Identifier {
    var ptr = GetChildPointer(tree, parent)
    if Empty(tree, ptr) {
        return NullIdentifier
    } else {
        return Identifier {
            Node: GetNode(tree, ptr, nil),
            Name: GetTokenContent(tree, ptr),
        }
    }
}

// type_args? = NoLF [ typelist! ]!
func type_args (tree Tree, parent Pointer) []TypeExpr {
    var ptr = GetChildPointer(tree, parent)
    if Empty(tree, ptr) {
        return []TypeExpr {}
    } else {
        return typelist(tree, ptr)
    }
}

// typelist = type typelist_tail
func typelist (tree Tree, parent Pointer) []TypeExpr {
    var ptr = GetChildPointer(tree, parent)
    var type_ptrs = FlatSubTree(tree, ptr, "type", "typelist_tail")
    var result = make([]TypeExpr, len(type_ptrs))
    for i, type_ptr := range type_ptrs {
        result[i] = type_(tree, type_ptr)
    }
    return result
}

// type_attached = attached_name
func type_attached (tree Tree, parent Pointer) AttachedTypeExpr {
    var ptr = GetChildPointer(tree, parent)
    return AttachedTypeExpr {
        Node: GetNode(tree, ptr, nil),
        AttachedExpr: attached_name(tree, ptr),
    }
}

// attached_name = _at type! :! name!
func attached_name (tree Tree, parent Pointer) AttachedExpr {
    var ptr = GetChildPointer(tree, parent)
    return AttachedExpr {
        Node: GetNode(tree, ptr, nil),
        Type: type_(tree, ptr),
        AttachedName: name(tree, ptr),
    }
}