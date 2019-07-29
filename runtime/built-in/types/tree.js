Types.TreeNodeInflater = create_fun_sig (
    [Types.Hash, Types.List], Types.TypePlaceholder
)

function inflate_tree_node (inflater, props, children) {
    ensure(is(inflater, Types.TreeNodeInflater), 'invalid_tree_node_inflater')
    assert(is(props, Types.Hash))
    assert(is(children, Types.List))
    return call(inflater, [props, children])
}
