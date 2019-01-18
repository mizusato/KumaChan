'use strict';


function normalize_operator_name (name) {
    check(normalize_operator_name, arguments, { name: Str })
    /**
     *  if the operator is SimpleOperator, 'name' is name of token
     *  if the operator is MapOperator, 'name' is matched string of token
     */
    let prefix = is(name[0], Char.Alphabet)? 'operator_': ''
    return (prefix + name.toLowerCase())
}


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
        current_part: Str, next_part: Str, separator: Str, f: Fun
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
                && assert(is(tree.children[0], SyntaxTreeLeaf))
                && is(tree.children[0].children, Token('Parameter'))
            ) || (tree.name == 'LambdaPara')
        })
        let FuncTree = $(tree => is(tree.name, one_of(
            'FunDef', 'FunExpr', 'HashLambda', 'ListLambda', 'SimpleLambda'
        )))
        let Drop = $(function (x) { return x === this })
        let drop = (t => Drop)
        return apply_on(tree.children, chain(
            x => map_lazy(x, t => transform(t, [
                { when_it_is: SyntaxTreeEmpty, use: drop },
                { when_it_is: SyntaxTreeLeaf, use: drop },
                { when_it_is: FuncTree, use: drop },
                { when_it_is: Parameter, use: t => [t] },
                { when_it_is: Otherwise, use: t => crush(t) }
            ])),
            x => filter(x, (y => is_not(y, Drop))),
            x => concat(x),
            x => list(x)
        ))
    }
    let linear_list = map(crush(tree), function (para_unit) {
        let h = children_hash(para_unit)
        assert(has(h, 'Identifier'))
        let id_token = h.Identifier.children
        return {
            pos: id_token.position,
            name: id_token.matched.string
        }
    })
    let parameters = apply_on(linear_list, chain(
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
    let parameters = apply_on(tree, chain(
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
    Constraint: use_first,
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
    Outer: function (tree) {
        let h = children_hash(tree)
        return `assign_outer(scope, ${translate(h.Id)}, ${translate(h.Expr)})`
    },
    Return: function (tree) {
        let h = children_hash(tree)
        return `return ${translate(h.Expr)}`
    },
    Hash: function (tree) {
        let h = children_hash(tree)
        let content = (has(h, 'PairList'))? translate_next(
            h.PairList, 'Pair', 'NextPair', ', ', function (pair) {
                let h = children_hash(pair)
                return `${translate(h.Id)}: ${translate(h.Expr)}`
            }
        ): ''
        return `HashObject({${content}})`
    },
    List: function (tree) {
        let h = children_hash(tree)
        let content = (has(h, 'ItemList'))? translate_next(
            h.ItemList, 'Item', 'NextItem', ', ',
            (item) => `${use_first(item)}`
        ): ''
        return `ListObject([${content}])`
    },
    Expr: use_first,
    MapOperand: use_first,
    MapOperator: function (tree) {
        let name = string_of(tree.children[0])
        let normalized = normalize_operator_name(name)
        let escaped = escape_raw(normalized)
        return `(apply(id('${escaped}')))`
    },
    MapExpr: function (tree) {
        let first_operand = use_first(tree)
        let h = children_hash(tree)
        let items = map(iterate_next(tree, 'MapNext'), children_hash)
        let shifted = items.slice(1, items.length)
        return fold(shifted, first_operand, function(h, reduced_string) {
            let operator_string = translate(h.MapOperator)
            let operand = h.FunExpr || h.BodyLambda || h.MapOperand
            let operand_string = translate(operand)
            return `${operator_string}(${reduced_string}, ${operand_string})`
        })
    },
    Filter: use_first,
    FilterList: function (tree) {
        let conditions = translate_next(
            tree, 'Filter', 'NextFilter', ' && ',
            f => `assert_bool(${translate(f)})`
        )
        return `${conditions}`
    },
    Concept: function (tree) {
        let h = children_hash(tree)
        let parameter = translate(h.Id)
        let filter_list = translate(h.FilterList)
        let f = function_string(`return ${filter_list}`)
        let checker = `Lambda(scope, [${parameter}], ${f})`
        return `Abstract(${checker})`
    },
    Struct: function (tree) {
        let h = children_hash(tree)
        let hash_object = translate(h.Hash)
        return `Structure(${hash_object})`
    },
    ListExprArgList: function (tree) {
        let table_content = translate_next(
            tree, 'ListExprArg', 'NextListExprArg', ', ',
            function (arg) {
                let h = children_hash(arg)
                return `${translate(h.Id)}: ${translate(h.Simple)}`
            }
        )
        let bind_code = translate_next(
            tree, 'ListExprArg', 'NextListExprArg', '; ',
            function (arg) {
                let h = children_hash(arg)
                let name = translate(h.Id)
                assert(name != `'__'`)
                return `scope.emplace(${name}, get(id('__'), ${name}))`
            }
        )
        return {
            bind_code: bind_code,
            iterables: `HashObject({${table_content}})`
        }
    },
    ListExprFilterList: function (tree) {
        let h = children_hash(tree)
        return (h.FilterList)? translate(h.FilterList): 'true'
    },
    IteratorExpr: function (tree) {
        let h = children_hash(tree)
        let expr = use_first(h.MapOperand)
        let arg_list = translate(h.ListExprArgList)
        let parameters = arg_list.parameters
        let bind_code = arg_list.bind_code
        let iterables = arg_list.iterables
        let filter_list = translate(h.ListExprFilterList)
        let f = function_string(`${bind_code}; return ${filter_list}`)
        let filter_lambda = `Lambda(scope, [], ${f})`
        let g = function_string(`${bind_code}; return ${expr}`)
        let mapper = `Lambda(scope, [], ${g})`
        let zipped = `K.zip.apply(${iterables})`
        let filtered = `K.filter.apply(${zipped}, ${filter_lambda})`
        return `K.map.apply(${filtered}, ${mapper})`
    },
    ListExpr: function (tree) {
        let iterator = Translate.IteratorExpr(tree)
        return `K.list.apply(${iterator})`
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
        let parameters = (has(h, 'Para'))? translate_next(
            tree, 'Para', 'NextPara', ', ', function (para) {
                let h = children_hash(para)
                let name = translate(h.Id)
                let constraint = translate(h.Constraint)
                let is_immutable = is(h.PassFlag, SyntaxTreeEmpty)
                let pass_policy = is_immutable? `'immutable'`: ({
                    '&': `'dirty'`,
                    '*': `'natural'`,
                })[string_of(h.PassFlag.children[0])]
                assert(is(pass_policy, Str))
                return `[${name},${constraint},${pass_policy}]`
            }
        ): ''
        return `[${parameters}]`
    },
    Target: function (tree) {
        let h = children_hash(tree)
        return (has(h, 'Constraint'))? translate(h.Constraint): 'AnyConcept'
    },
    FunExpr: function (tree) {
        let h = children_hash(tree)
        let Flag = h.FunFlag
        let effect = Flag? (function() {
            let h = children_hash(Flag)
            let flag = string_of(h.Keyword)
            return ({
                global: `'global'`,
                g: `'global'`,
                upper: `'upper'`,
                h: `'upper'`,
                local: `'local'`,
                f: `'local'`,
            })[flag]
        })(): `'local'`
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
        let normal = (name => escape_raw(normalize_operator_name(name)))
        let func_name = ({
            Call: 'call',
            Get: 'get',
            Access: 'access'
        })[op_name] || `(apply(id('${normal(op_name)}')))`
        let arg_str_list = map(args, a => translate(a))
        if (op_name == 'Access') { arg_str_list.push('scope') }
        let arg_list_str = join(arg_str_list, ', ')
        return `${func_name}(${arg_list_str})`
    },
    ArgList: function (tree) {
        let key_list = $(tree => has(children_hash(tree), 'KeyArg'))
        let index_list = $(tree => has(children_hash(tree), 'Arg'))
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
    assert(!has(tree, 'ok') || tree.ok == true)
    return Translate[tree.name](tree)
}
