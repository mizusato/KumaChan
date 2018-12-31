function get_tokens (string) {
    check(get_tokens, arguments, { string: Str })
    return filter(
        CodeScanner(string).match(Matcher(Tokens)),
        token => token.matched.name != 'Space'
    )
}
