class ParserError extends Error {}


function parser_error (element, info) {
    check(parser_error, arguments, {
        element: Any, // $u(Token.Valid, SyntaxTreeNode),
        info: Str 
    })
    function get_message () {
        function get_token (tree) {
            assert(tree.is_not(SyntaxTreeEmpty))
            return transform(tree, [
                { when_it_is: SyntaxTreeLeaf, use: t => t.children },
                { when_it_is: Otherwise, use: t => get_token(t.children[0]) } 
            ])
        }
        let token = transform(element, [
            { when_it_is: Token.Valid, use: x => x },
            { when_it_is: Otherwise, use: x => get_token(x) }
        ])
        let { row, col } = token.position
        return `row ${row}, column ${col}: ${token.matched.name}: ${info}`
    }
    return {
        assert: function (condition) {
            this.if(!condition)
        },
        if: function (condition) {
            if (condition) { this.throw() }
        },
        throw: function () {
            throw ( new ParserError(get_message()) )
        }
    }
}


function get_tokens (string) {
    check(get_tokens, arguments, { string: Str })
    function assert_valid (tokens) {
        return map_lazy(tokens, t => (assert(t.is(Token.Valid)) && t))
    }
    function remove_comment (tokens) {
        /* remove ordinary comment and inline comment */
        let Current = name => $(look => look.current.is(Token(name)))
        let Left = name => $(look => look.left.is(Token(name)))
        let InlineCommentElement = $u(
            Current('..'),
            $n(Left('..'), Current('Name'))
        )
        return tokens.transform_by(chain(
            x => filter_lazy(x, t => t.is_not(Token('Comment'))),
            x => lookaside(x, Token.Null),
            x => filter_lazy(x, look => look.is_not(InlineCommentElement)),
            x => map_lazy(x, look => look.current)
        ))
    }
    function remove_space (tokens) {
        /**
         *  note: linefeed won't be removed by this function.
         *        therefore, the effect of add_call_operator()
         *        and add_get_operator() cannot cross lines.
         */
        return filter_lazy(tokens, token => token.matched.name != 'Space')
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
    function remove_linefeed (tokens) {
        return filter_lazy(tokens, token => token.matched.name != 'Linefeed')
    }
    function eliminate_ambiguity (tokens) {
        function transform (item) {
            let { left, current, right } = item
            let token = current
            let mapper = {
                '.': function () {
                    let left_is_name = left.is(Token('Name'))
                    let name = left_is_name? 'Access': 'Parameter'
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
    let raw = CodeScanner(string).match(Matcher(Tokens))
    return raw.transform_by(chain(
        assert_valid,
        remove_comment,
        remove_space,
        add_call_operator,
        add_get_operator,
        remove_linefeed,
        eliminate_ambiguity,
        assert_valid,
        list
    ))
}


const SyntaxTreeEmpty = Struct({
    name: Str,
    children: $n(Array, $(a => a.length == 0))
})
const SyntaxTreeLeaf = Struct({
    name: Str,
    children: Token.Valid
})
const SyntaxTreeNode = $u(SyntaxTreeLeaf, Struct({
    name: Str,
    children: ArrayOf( $u(SyntaxTreeLeaf, $(x => x.is(SyntaxTreeNode))) )
}))
const SyntaxTreeRoot = $n(SyntaxTreeNode, Struct({
    ok: Bool,
    amount: Int
}))


function print_tree (tree, deepth = 0, is_last = [true]) {
    let indent_increment = 2
    let repeat = function (string, n, f = (i => '')) {
        return join(map_lazy(count(n), i => f(i) || string), '')
    }
    let base = repeat(' ', indent_increment)
    let indent = repeat(base, deepth,
        i => (
            is_last[i]?
            repeat(' ', indent_increment):
            ('│'+repeat(' ', indent_increment-1))
        )
    )
    let pointer = (
        ((is_last[deepth])? '└': '├')
        + repeat('─', indent_increment-1)
        + ((tree.children.length > 0)? '┬': '─')
        + '─'
    )
    let node_name = `${tree.name}`
    let node_children = transform(tree, [
        { when_it_is: SyntaxTreeEmpty, use: t => (' ' + '[]') },
        { when_it_is: SyntaxTreeLeaf, use: function (tree) {
            let string = tree.children.matched.string
            return (string == tree.name)? '': `: ${string}`
        } },
        { when_it_is: Otherwise, use: function (tree) {
            let last_index = tree.children.length-1
            let subtree_string = tree.children.transform_by(chain(
                x => map_lazy(x, (child, index) => print_tree(
                    child, deepth+1,
                    is_last.added(index == last_index)
                )),
                x => join(x, LF)
            ))
            return (LF + subtree_string)
        } }
    ])
    return (indent + pointer + node_name + node_children)
}


function build_leaf (token) {
    check(build_leaf, arguments, { token: Token.Valid })
    return { amount: 1, name: token.matched.name, children: token }
}


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
    //assert(match.is(SyntaxTreeRoot))
    return match
}


function parse_simple (syntax, tokens, pos) {
    /**
     *  parse simple experssion using shunting yard algorithm
     *
     *  1. invoke converge() to parse argument list and key expression
     *  2. define operations on output & operator stack 
     *  3. run the shunting yard algorithm
     *  4. get final state of stacks and return result
     */
    function* converge () {
        let offset = 0
        while (pos+offset < tokens.length) {
            let p = pos + offset
            let token = tokens[p]
            let left = (p-1 >= 0)? tokens[p-1]: Token.Null
            let is_args = ( token.is(Token('(')) && left.is(Token('Call')) )
            let is_key = ( token.is(Token('[')) && left.is(Token('Get')) )
            let is_par = ( token.is(Token('(')) && left.is_not(Token('Call')) )
            if ( is_args || is_key ) {
                let type = is_args? 'args': 'key'
                let syntax_item = (
                    { args: 'Arguments', key: 'Key' }
                )[type]
                let err_msg = (
                    { args: 'bad argument list', key: 'bad key' }
                )[type]
                let match = build_tree(syntax, syntax_item, tokens, p)
                let err = parser_error(token, err_msg)
                err.assert(match.ok)
                yield match
                offset += match.amount
            } else if ( is_par ) {
                let syntax_item = 'WrappedSimple'
                let err_msg = 'missing )'
                let match = build_tree(syntax, syntax_item, tokens, p)
                let err = parser_error(token, err_msg)
                err.assert(match.ok)
                yield match
                offset += match.amount
            } else {
                let left_is_operand = left.is($u(SimpleOperand, Token(')')))
                let this_is_operand = token.is($u(SimpleOperand, Token('(')))
                if ( left_is_operand && this_is_operand ) {
                    break
                } else {
                    yield build_leaf(token)
                    offset += 1
                }
            }
        }
    }
    let sentinel = {
        position: { row: -1, col: -1 },
        matched: { category: 'Operator', name: 'Sentinel', string: '' }
    }
    let input = cat(converge(), [{ name: 'Sentinel', children: sentinel }])
    let info = (operator => SimpleOperator[operator.matched.name])
    function empty (state) {
        let operators = state.operators
        return operators.length == 0
    }
    function top (state) {
        let operators = state.operators
        assert(operators.length > 0)
        return operators[operators.length-1]
    }
    function pop (state) {
        let operators = state.operators
        let output = state.output
        let operator = top(state)
        let err = parser_error(operator, 'missing operand')
        let type = info(operator).type
        let count = ({ prefix: 1, infix: 2 })[type]
        err.assert(output.length >= count)
        let take_out = output.slice(output.length-count, output.length)
        let remaining = output.slice(0, output.length-count)
        let poped = operators.slice(0, operators.length-1)
        let children = take_out.added_front(build_leaf(operator))
        let reduced = {
            name: 'SimpleUnit',
            children: children,
            amount: fold(children, 0, (tree, sum) => sum+tree.amount)
        }
        return {
            output: remaining.added(reduced),
            operators: poped
        }
    }
    function push (state, operator) {
        let operators = state.operators
        let output = state.output
        return {
            output: output,
            operators: operators.added(operator)
        }
    }
    function append (state, operand) {
        let operators = state.operators
        let output = state.output
        let element = operand.is(Token)? build_leaf(operand): operand
        return {
            output: output.added(element),
            operators: operators
        }
    }
    function put (state, operator) {
        let operators = state.operators
        let output = state.output
        return (empty(state))? push(state, operator): (function() {
            let input = operator
            let stack = top(state)
            let input_info = info(input)
            let stack_info = info(stack)
            let assoc = input_info.assoc
            let should_pop = ({
                left: input_info.priority <= stack_info.priority,
                right: input_info.priority < stack_info.priority
            })[assoc]
            return should_pop? put(pop(state), operator): push(state, input)
        })()
    }
    function get_result (state) {
        let operators = state.operators
        let output = state.output
        let output_first = output[0] || null
        let err = parser_error(output_first, 'missing operator')
        err.if(output.length > 1)  // 1 = final result
        assert(output.length == 1 && operators.length == 1)
        return output_first
    }
    let initial = { output: [], operators: [] }
    let final = fold(input, initial, function (element, state) {
        console.log(state)
        let TreeElement = $_(SyntaxTreeLeaf)
        let LeafElement = SyntaxTreeLeaf
        let add_to_output = (operand => append(state, operand))
        let put_operator = (operator => put(state, operator))
        let terminate = (t => put(state, sentinel))
        let should_break = (!empty(state) && top(state) == sentinel)
        return should_break? Break: transform(element, [
            { when_it_is: TreeElement, use: add_to_output },
            { when_it_is: LeafElement, use: l => transform(l.children, [
                { when_it_is: SimpleOperand, use: add_to_output },
                { when_it_is: SimpleOperator, use: put_operator },
                { when_it_is: Otherwise, use: terminate }
            ]) }
        ])
    })
    let EmptyState = Struct({
        output: $(x => x.length == 0),
        operators: $(x => x.length == 1)
    })
    return (final.is(EmptyState))? { ok: false }: (function() {
        let result = get_result(final)
        return {
            ok: true,
            amount: result.amount,
            name: 'Simple',
            children: [result]
        }
    })()
}
