function get_tokens (string) {
    check(get_tokens, arguments, { string: Str })
    let raw = CodeScanner(string).match(Matcher(Tokens))
    return raw.transform_by(chain(
        x => map_lazy(x, t => assert(t.is(Token.Valid)) && t),
        x => remove_space(x),
        x => add_call_operator(x),
        x => add_get_operator(x),
        x => remove_linefeed(x),
        x => map_lazy(x, t => assert(t.is(Token.Valid)) && t),
        x => list(x)
    ))
}


function remove_space (tokens) {
    return filter_lazy(tokens, token => token.matched.name != 'Space')
}


function remove_linefeed (tokens) {
    return filter_lazy(tokens, token => token.matched.name != 'Linefeed')
}


function add_call_operator (tokens) {
    return insert(tokens, Token.Null, function (token, next) {
        let this_is_id = token.is(Token('Identifier'))
        let this_is_rp = token.is(Token(')'))
        let next_is_lp = next.is(Token('('))
        let need_insert = (this_is_id || this_is_rp) && next_is_lp
        return need_insert? {
            position: token.position,
            matched: {
                category: 'Operator',
                name: 'Call',
                string: '@'
            }
        }: Nothing
    })
}


function add_get_operator (tokens) {
    return insert(tokens, Token.Null, function (token, next) {
        let this_is_id = token.is(Token('Identifier'))
        let next_is_lb = next.is(Token('['))
        let need_insert = (this_is_id && next_is_lb)
        return need_insert? {
            position: token.position,
            matched: {
                category: 'Operator',
                name: 'Get',
                string: '#'
            }
        }: Nothing
    })
}


const EmptySyntaxTree = Struct({ ok: $1(false) })
const SyntaxTreeNode = $u(EmptySyntaxTree, Struct({
    name: Str,
    children: $u(Token.Valid, ArrayOf($(x => x.is(SyntaxTreeNode))))
}))
const SyntaxTreeRoot = $n(SyntaxTreeNode, Struct({
    ok: Bool,
    amount: Int
}))


function build_tree (syntax, root, tokens) {
    function iterator_at (pos) {
        return function* iterator () {
            for (let i=pos; i<tokens.length; i++) {
                yield tokens[i]
            }
        }
    }
    function match_part (part, pos) {
        let SyntaxPart = $(part => syntax.has(part))
        let TokenPart = $(part => true)
        let token = (pos < tokens.length)? tokens[pos]: ''
        return transform(part, [
            { when_it_is: SyntaxPart,
              use: part => match_item(syntax[part], pos) },
            { when_it_is: TokenPart,
              use: part => (token.is(Token(part)) && {
                  ok: true,
                  amount: 1,
                  children: token
              } || { ok: false }) }
        ])
    }
    function match_derivation (derivation, pos) {
        let initial = { amount: 0, children: [], ok: true }
        return fold(derivation, initial, function (part, state) {
            let { ok, amount, children } = state
            return ok? (function () {
                let p = pos + amount
                let match = match_part(part, p)
                let proceed = match.ok? match.amount: 0
                let new_child = match.ok? {
                    name: part,
                    children: match.children
                }: Nothing
                return {
                    ok: match.ok,
                    amount: amount + proceed,
                    children: children.added(new_child)
                }
            })(): state
        })
    }
    function match_item (item, pos) {
        let ReduceItem = $(x => x.has('reducers'))
        let DeriveItem = $(x => x.has('derivations'))
        let iterator = iterator_at(pos)
        let matches = transform(item, [
            { when_it_is: ReduceItem, use: item => (
                map_lazy(item.reducers, reducer => reducer(iterator))) },
            { when_it_is: DeriveItem, use: item => (
                map_lazy(item.derivations, d => match_derivation(d, pos))) }
        ])
        let match = find(matches, match => match.ok)
        return (match != NotFound)? match: { ok: false }
    }
    let match = match_item(syntax[root], 0)
    let finish = (match.amount == tokens.length)
    let tree = finish? match: { ok: false, amount: 0, children: [] } 
    tree.name = root
    assert(tree.is(SyntaxTreeRoot))
    return tree
}


function parse_simple (iterator) {
    let input = lookahead(iterator())
    let output = []
    let operators = []    
    let item = input.next()
    while (!item.done) {
        let token = item.current.value
        
        item = input.next()
    }
}


/**

列表 >> 去重 ->  { .每個元素.数值*3 }
数值 >> 平方 >> 開立方 >> { f(.x) + g(.x) }

数值 >> { __ + 1 } >> { __ * 5 }

local 模長 (向量 v) {
    return [v.x, v.y] -> 平方 >> 求和 >> 開方
}

local 模長 (向量 v) {
    return [v.x, v.y] -> { __^2 } >> sum >> sqrt
}

列表 -> { create({ tag: 'div', text: .元素.內容 }) }

9 >> { __*3 }

f(x, y) >> { g(u, v, __) }

函数(甲, 乙) >> { 函数(丙, 丁, __) }

Expression:
    Identifier
    Literal String
    Number
    Calculation Operator
    Simple Expression
    List
    Hash
    Structure Expression
    Lambda
    Map Operator
    Chain Expression
    Function
*/
