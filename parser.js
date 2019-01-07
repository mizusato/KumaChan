class ParserError extends Error {
    assert (condition) {
        this.if(!condition)
    }
    if (condition) {
        if (condition) { this.throw() }
    }
    throw () {
        throw this
    }
}


function parser_error (token, info) {
    check(parser_error, arguments, { token: Token.Valid, info: Str })
    let p = token.position
    return new ParserError(
        `row ${p.row}, column ${p.col}: ${info}`
    )
}


function get_tokens (string) {
    check(get_tokens, arguments, { string: Str })
    let raw = CodeScanner(string).match(Matcher(Tokens))
    return raw.transform_by(chain(
        x => map_lazy(x, t => assert(t.is(Token.Valid)) && t),
        x => remove_comment(x),
        x => remove_space(x),
        x => add_call_operator(x),
        x => add_get_operator(x),
        x => remove_linefeed(x),
        x => eliminate_ambiguity(x),
        x => map_lazy(x, t => assert(t.is(Token.Valid)) && t),
        x => list(x)
    ))
}


function remove_comment (tokens) {
    return filter_lazy(tokens, token => token.matched.name != 'Comment')
}


function remove_space (tokens) {
    return filter_lazy(tokens, token => token.matched.name != 'Space')
}


function remove_linefeed (tokens) {
    return filter_lazy(tokens, token => token.matched.name != 'Linefeed')
}


function add_call_operator (tokens) {
    return insert(tokens, Token.Null, function (token, next) {
        let this_is_id = token.is(Token('Name'))
        let this_is_rp = token.is(Token(')'))
        let next_is_lp = next.is(Token('('))
        let need_insert = (this_is_id || this_is_rp) && next_is_lp
        return (
            need_insert?
            Token.create_from(token, 'Operator', 'Call', '@'): Nothing
        )
    })
}


function add_get_operator (tokens) {
    return insert(tokens, Token.Null, function (token, next) {
        let this_is_id = token.is(Token('Name'))
        let next_is_lb = next.is(Token('['))
        let need_insert = (this_is_id && next_is_lb)
        return (
            need_insert?
            Token.create_from(token, 'Operator', 'Get', '#'): Nothing
        )
    })
}


function eliminate_ambiguity (tokens) {
    function transform (item) {
        let { left, current, right } = item
        let token = current
        let mapper = {
            '.': function () {
                let left_is_id = left.is(Token('Name'))
                let name = left_is_id? 'Access': 'Parameter'
                return Token.create_from(token, 'Operator', name)
            },
            Name: function () {
                let left_is_dot = left.is(Token('.'))
                let name = left_is_dot? 'Member': 'Identifier'
                return Token.create_from(token, 'Name', name)
            }
        }
        let name = token.matched.name
        return mapper.has(name)? mapper[name](): token
    }
    return map_lazy(lookaside(tokens, Token.Null), transform)
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


function build_tree (syntax, root, tokens, pos = 0) {
    function match_part (part, pos) {
        let SyntaxPart = $(part => syntax.has(part))
        let TokenPart = $(part => true)
        let token = (pos < tokens.length)? tokens[pos]: ''
        return transform(part, [
            { when_it_is: SyntaxPart,
              use: part => match_item(part, pos) },
            { when_it_is: TokenPart,
              use: part => (token.is(Token(part)) && {
                  ok: true,
                  amount: 1,
                  name: part,
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
                    name: match.name,
                    children: match.children
                }: Nothing
                return {
                    ok: match.ok,
                    amount: amount + proceed,
                    children: children.added(new_child)
                }
            })(): Break
        })
    }
    function match_item (item_name, pos) {
        let ReduceItem = $(x => x.has('reducers'))
        let DeriveItem = $(x => x.has('derivations'))
        let item = syntax[item_name]
        let matches = transform(item, [
            { when_it_is: ReduceItem, use: item => (
                map_lazy(item.reducers, r => (r())(syntax, tokens, pos))) },
            { when_it_is: DeriveItem, use: item => (
                map_lazy(item.derivations, d => match_derivation(d, pos))) }
        ])
        let match = find(matches, match => match.ok)
        match = (match != NotFound)? match: { ok: false }
        match.name = item_name
        return match
    }
    let match = match_item(root, pos)
    let finish = (match.amount == tokens.length)
    let tree = finish? match: { ok: false, amount: 0, name: root, children: [] }
    assert(tree.is(SyntaxTreeRoot))
    return tree
}


function parse_simple (syntax, tokens, pos) {
    function* converge () {
        let offset = 0
        while (pos+offset < tokens.length) {
            let p = pos + offset
            let token = tokens[p]
            let left = (p-1 >= 0)? tokens[p-1]: { name: null }
            let name = token.matched.name
            let left_name = left.matched.name
            let is_args = (name == '(' && left_name == 'Call')
            let is_key = (name == '[' && left_name == 'Get')
            if ( is_args || is_key ) {
                let type = is_args? 'args': 'key'
                let syntax_item = (
                    { arg: 'Arguments', key: 'Key' }
                )[type]
                let err_msg = (
                    { arg: 'bad argument list', key: 'bad key' }
                )[type]
                let match = build_tree(syntax, syntax_item, tokens, p)
                let err = parser_error(token, err_msg)
                err.assert(match.ok)
                yield { type: 'tree', value: match }
                offset += match.amount
            } else {
                yield { type: 'token', value: token }
                offset += 1
            }
        }
    }
    let initial = { output: [], operators: [] }
    fold(converge(), initial, function (element, state) {
        let info = (operator => SimpleOperator[operator.matched.name])
        let output = state.output
        let operators = state.operators
        let empty = () => (operators.length == 0)
        let top = () => (
            assert(operators.length > 0)
            && operators[operators.length-1]
        )
        let add_to_output = (operand => ({
            output: output.added(operand),
            operators: operators
        }))
        let push = (operator => ({
            output: output,
            operators: operators.added(operator)
        }))
        let pop = () => pop_and_push(Nothing)
        let pop_and_push = function (new_operator) {
            let err = parser_error(by_operator, 'missing operand')
            let operator = top()
            let type = info(operator).type
            let count = ({ prefix: 1, infix: 2 })[type]
            let check = err.assert(output.length >= count)
            let take_out = output.slice(output.length-count, output.length)
            let remaining = output.slice(0, output.length-count)
            let poped = operators.slice(0, operators.length-1)
            let reduced = {
                name: 'ExprUnit',
                children: take_out.added_front(operator)
            }
            return {
                output: remaining.added(reduced),
                operators: poped.added(new_operator)
            }
        }
        let put_operator = function (operator) {
            return (empty())? push(operator): (function() {
                let input = operator
                let stack = top()
                let input_info = info(input)
                let stack_info = info(stack)
                let assoc = input_info.assoc
                let should_pop = ({
                    left: input_info.priority <= stack_info.priority,
                    right: input_info.priority < stack_info.priority
                })[assoc]
                return should_pop? pop_and_push(input): push(input)
            })()
        }
        let panic = token => parser_error(token, 'bad operator').throw()
        let TreeElement = $(x => x.type == 'tree')
        let TokenElement = $(x => x.type == 'token')
        return transform(element, [
            { when_it_is: TreeElement, use: add_to_output },
            { when_it_is: TokenElement, use: token => transform(token, [
                { when_it_is: SimpleOperand, use: add_to_output },
                { when_it_is: SimpleOperator, use: put_operator },
                { when_it_is: Otherwise, use: panic }
            ]) }
        ])
    })
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
    Name
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
