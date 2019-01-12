'use strict';


function escape_raw (string) {
    check(escape_raw, arguments, { string: Str })
    function f (char) {
        if (char == '\\') {
            return '\\\\'
        } else if (char == "'") {
            return "\\'"
        } else {
            return char
        }
    }
    return join(map_lazy(string, f), '')
}


function id_reference (string) {
    check(id_reference, arguments, { string: Str })
    let escaped = escape_raw(string)
    return `id('${escaped}')`
}


function string_of (leaf) {
    check(string_of, arguments, { leaf: SyntaxTreeLeaf })
    let token = leaf.children
    return token.matched.string
}


function use_string (leaf) {
    return string_of(leaf)
}


function use_first (tree) {
    assert(tree.children.length > 0)
    return translate(tree.children[0])
}


function children_hash (tree) {
    return fold(tree.children, {}, (child, h) => (h[child.name] = child, h))
}


function translate_next(tree, current_part, next_part, separator, f, depth=0) {
    check(translate_next, arguments, {
        current_part: Str, next_part: Str, separator: Str, f: Function
    })
    let HasNext = (next_part => $(function (tree) {
        let h = children_hash(tree)
        return h[next_part] && h[next_part].is_not(SyntaxTreeEmpty)
    }))
    let Last = (next_part => $(function (tree) {
        let h = children_hash(tree)
        return h[next_part] && h[next_part].is(SyntaxTreeEmpty)
    }))
    let Empty = $(tree => tree.is(SyntaxTreeEmpty))
    let has_next = HasNext(next_part)
    let last = Last(next_part)
    let empty = Empty
    let h = children_hash(tree)
    let call_next = (
        (tree, depth) => translate_next(
            tree, current_part, next_part, separator, f, depth
        )
    )
    return transform(tree, [
        { when_it_is: has_next,
          use: () => `${f(h[current_part], depth)}` + separator
                    + `${call_next(h[next_part], depth+1)}` },
        { when_it_is: last,
          use: () => `${f(h[current_part], depth)}` },
        { when_it_is: empty,
          use: () => '' }
    ])
}


let Translate = {
    Simple: use_first,
    SimpleUnit: function (tree) {
        let args = tree.children.slice(1, tree.children.length)
        let operator = tree.children[0].children
        let op_name = operator.matched.name
        let func_name = ({
            Call: 'call',
            Get: 'get',
            Access: 'access'
        })[op_name] || `(id('operator::${op_name.toLowerCase()}'))`
        let arg_str_list = map(args, a => translate(a))
        if (op_name == 'Access') { arg_str_list.push('scope') }
        let arg_list_str = join(arg_str_list, ', ')
        return `${func_name}(${arg_list_str})`
    },
    ArgList: function (tree) {
        let list = translate_next(
            tree, 'Arg', 'NextArg', ', ', function(t, i) {
                return `'${i}': ${use_first(t)}`
            }
        )
        return `{ ${list} }`
    },
    Key: function (tree) {
        let h = children_hash(tree)
        return translate(h.Simple)
    },
    Identifier: function (leaf) {
        return id_reference(string_of(leaf))
    },
    Member: function (leaf) {
        let string = string_of(leaf)
        let escaped = escape_raw(string)
        return `'${escaped}'`
    },
    RawString: function (leaf) {
        let string = string_of(leaf)
        let content = string.slice(1, string.length-1)
        let escaped = escape_raw(content)
        return `'${escaped}'`
    },
    FormatString: function (leaf) {
        let string = string_of(leaf)
        let content = string.slice(1, string.length-1)
        let escaped = escape_raw(content)
        return `FormatString('${escaped}', id)`
    },
    Integer: use_string,
    Float: use_string,
    Exponent: use_string
}


function translate (tree) {
    return Translate[tree.name](tree)
}
