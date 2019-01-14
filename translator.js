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


function iterate_next (tree, next_part) {
    return iterate(tree, t => children_hash(t)[next_part], SyntaxTreeEmpty)
}


function translate_next(tree, current_part, next_part, separator, f) {
    check(translate_next, arguments, {
        current_part: Str, next_part: Str, separator: Str, f: Function
    })
    let link = (x => join(x, separator))
    return link(map_lazy(
        iterate_next(tree, next_part),
        (t, i) => f(children_hash(t)[current_part], i)
    ))
}


function find_parameters (tree) {
    function crush (tree) {
        let Parameter = $(function (tree) {
            return (
                (tree.name == 'SimpleUnit')
                && assert(tree.children[0].is(SyntaxTreeLeaf))
                && tree.children[0].children.is(Token('Parameter'))
            )
        })
        let FuncTree = $(tree => tree.name.is(one_of(
            'FunDef', 'FunExpr', 'HashLambda', 'ListLambda', 'SimpleLambda'
        )))
        let Drop = $(function (x) { return x === this })
        let drop = (t => Drop)
        return (tree.children).transform_by(chain(
            x => map_lazy(x, t => transform(t, [
                { when_it_is: SyntaxTreeEmpty, use: drop },
                { when_it_is: SyntaxTreeLeaf, use: drop },
                { when_it_is: FuncTree, use: drop },
                { when_it_is: Parameter, use: t => [t] },
                { when_it_is: Otherwise, use: t => crush(t) }
            ])),
            x => filter(x, (y => y.is_not(Drop))),
            x => concat(x),
            x => list(x)
        ))
    }
    let linear_list = map(crush(tree), function (simple_unit) {
        let h = children_hash(simple_unit)
        assert(h.has('Identifier'))
        let id_token = h.Identifier.children
        return {
            pos: id_token.position,
            name: id_token.matched.string
        }
    })
    let parameters = linear_list.transform_by(chain(
        x => fold(x, {}, function (item, hash) {
            let name = item.name
            if (hash[name]) {
                let r = Token.pos_compare(item.pos, hash[name].pos)
                if (r == -1) {
                    hash[name] = item
                }
            } else {
                hash[name] = item
            }
            return hash
        }),
        x => map(x, (key, value) => value),
        x => x.sort((x,y) => Token.pos_compare(x.pos, y.pos)),
        x => map(x, item => item.name)
    ))
    return parameters
}


function function_string (body) {
    let head = `var id = Lookup(scope);`
    return `(function (scope) { ${head} ${body} })`
}


function translate_lambda (tree, type, get_val) {
    let parameters = tree.transform_by(chain(
        x => find_parameters(x),
        x => map_lazy(x, name => `'${escape_raw(name)}'`),
        x => join(x, ', ')
    ))
    let parameter_list = `[${parameters}]`
    let value = get_val(tree)
    let is_expr = (type == 'expr')
    let func = function_string(is_expr? `return ${value}`: value)
    return `Lambda(scope, ${parameter_list}, ${func})`
}


let Translate = {
    RawCode: function (leaf) {
        let string = string_of(leaf)
        let content = string.slice(2, string.length-2).trim()
        return `${content}`
    },
    /* ---------------------- */
    Concept: use_first,
    Id: function (tree) {
        let id_id = $(tree => children_hash(tree).Identifier)
        let str_id = $(tree => children_hash(tree).RawString)
        let string = transform(tree, [
            { when_it_is: id_id, use: tree => string_of(tree.children[0]) },
            { when_it_is: str_id, use: function (tree) {
                let wrapped = string_of(tree.children[0])
                let content = wrapped.slice(1, wrapped.length-1)
                return content
            } }
        ])
        let escaped = escape_raw(string)
        return `'${escaped}'`
    },
    Module: use_first,
    Program: function (tree) {
        return translate_next(
            tree, 'Command', 'NextCommand', '; ', cmd => translate(cmd)
        )
    },
    Command: use_first,
    Assign: function (tree) {
        let h = children_hash(tree)
        let LeftVal = children_hash(h.LeftVal)
        let head = translate(LeftVal.Id)
        let members = map_lazy(
            iterate_next(LeftVal.MemberNext, 'MemberNext'),
            t => translate(children_hash(t).Member)
        )
        let keys = map_lazy(
            iterate_next(LeftVal.KeyNext, 'KeyNext'),
            t => translate(children_hash(t).Key)
        )
        let total = list(cat(members, keys))
        let read = (total.length > 0)? total.slice(0, total.length-1): []
        let write = (total.length > 0)? total[total.length-1]: head
        let value = translate(h.Expr)
        return (total.length > 0)? (function() {
            let read_str = fold(read, `id(${head})`, function (key, reduced) {
                return `get(${reduced}, ${key})`
            })
            return `set(${read_str}, ${write}, ${value})`
        })(): `scope.replace(${write}, ${value})`
    },
    Let: function (tree) {
        let h = children_hash(tree)
        return `scope.emplace(${translate(h.Id)}, ${translate(h.Expr)})`
    },
    Return: function (tree) {
        let h = children_hash(tree)
        return `return ${translate(h.Expr)}`
    },
    Hash: function (tree) {
        let h = children_hash(tree)
        let content = (h.has('PairList'))? translate_next(
            h.PairList, 'Pair', 'NextPair', ', ', function (pair) {
                let h = children_hash(pair)
                return `${translate(h.Id)}: ${translate(h.Expr)}`
            }
        ): ''
        return `HashObject({${content}})`
    },
    List: function (tree) {
        let h = children_hash(tree)
        let content = (h.has('ItemList'))? translate_next(
            h.ItemList, 'Item', 'NextItem', ', ',
            (item) => `${use_first(item)}`
        ): ''
        return `ListObject([${content}])`
    },
    Expr: use_first,
    MapOperand: use_first,
    MapOperator: function (tree) {
        let name = string_of(tree.children[0])
        let escaped = escape_raw(name)
        return `(id('operator::${escaped}'))`
    },
    MapExpr: function (tree) {
        let first_operand = use_first(tree)
        let h = children_hash(tree)
        let items = map(iterate_next(tree, 'MapNext'), children_hash)
        let shifted = items.slice(1, items.length)
        return fold(shifted, first_operand, function(h, reduced_string) {
            let operator_string = translate(h.MapOperator)
            let operand = (h.FunExpr)? h.FunExpr: h.MapOperand
            let operand_string = translate(operand)
            return `${operator_string}(${reduced_string}, ${operand_string})`
        })
    },
    /* ---------------------- */
    SimpleLambda: function (tree) {
        return translate_lambda(tree, 'expr', function (tree) {
            let h = children_hash(tree)
            return translate(h.Simple || h.SimpleLambda)
        })
    },
    HashLambda: function (tree) {
        return translate_lambda(tree, 'expr', t => Translate.Hash(t))
    },
    ListLambda: function (tree) {
        return translate_lambda(tree, 'expr', t => Translate.List(t))
    },
    BodyLambda: function (tree) {
        return translate_lambda(tree, 'body', t => Translate.Body(t))
    },
    Body: function (tree) {
        let h = children_hash(tree)
        let program = translate(h.Program)
        return `${program}; return VoidObject`
    },
    ParaList: function (tree) {
        let h = children_hash(tree)
        let parameters = (h.has('Para'))? translate_next(
            tree, 'Para', 'NextPara', ', ', function (para) {
                let h = children_hash(para)
                let name = translate(h.Id)
                let constraint = translate(h.Concept)
                let is_immutable = h.PassFlag.is(SyntaxTreeEmpty)
                let pass_policy = is_immutable? `'immutable'`: `'dirty'`
                return `[${name},${constraint},${pass_policy}]`
            }
        ): ''
        return `[${parameters}]`
    },
    Target: function (tree) {
        let h = children_hash(tree)
        return (h.has('Concept'))? translate(h.Concept): 'AnyConcept'
    },
    FunExpr: function (tree) {
        let h = children_hash(tree)
        let flag = h.FunFlag
        let is_global = (
            flag.is_not(SyntaxTreeEmpty)
            && string_of(flag.children[0]) == 'g'
        )
        let effect = is_global? `'global'`: `'local'`
        let paras = translate(h.ParaList)
        let target = translate(h.Target)
        let func = function_string(translate(h.Body))
        return `FunInst(scope, ${effect}, ${paras}, ${target}, ${func})`
    },
    FunDef: function (tree) {
        let h = children_hash(tree)
        let effect_raw = string_of(h.Effect.children[0])
        let effect = `'${effect_raw}'`
        let name = translate(h.Id)
        let paras = translate(h.ParaList)
        let target = translate(h.Target)
        let func = function_string(translate(h.Body))
        return `define(scope, ${name}, ${effect}, ${paras}, ${target}, ${func})`
    },
    /* ---------------------- */
    Wrapped: function (tree) {
        let h = children_hash(tree)
        return translate(h.Simple)
    },
    Simple: use_first,
    SimpleUnit: function (tree) {
        let args = tree.children.slice(1, tree.children.length)
        let operator = tree.children[0].children
        let op_name = operator.matched.name
        if (op_name == 'Parameter') { return translate(args[0]) }
        let func_name = ({
            Call: 'call',
            Get: 'get',
            Access: 'access'
        })[op_name] || `(apply(id('operator::${op_name.toLowerCase()}')))`
        let arg_str_list = map(args, a => translate(a))
        if (op_name == 'Access') { arg_str_list.push('scope') }
        let arg_list_str = join(arg_str_list, ', ')
        return `${func_name}(${arg_list_str})`
    },
    ArgList: function (tree) {
        let key_list = $(tree => children_hash(tree).has('KeyArg'))
        let index_list = $(tree => children_hash(tree).has('Arg'))
        let empty_list = Otherwise
        let list_str = transform(tree, [
            { when_it_is: key_list, use: tree => (translate_next(
                tree, 'KeyArg', 'NextKeyArg', ', ', function(key_arg) {
                    let h = children_hash(key_arg)
                    return `${translate(h.Id)}: ${translate(h.Simple)}`
                }
            ))},
            { when_it_is: index_list, use: tree => (translate_next(
                tree, 'Arg', 'NextArg', ', ', function(arg, i) {
                    return `'${i}': ${use_first(arg)}`
                }
            ))},
            { when_it_is: empty_list, use: tree => '' }
        ])
        return `{${list_str}}`
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
