class ScanningError extends Error {}


function row_col_info (string) {
    let info = []
    let row = 1
    let column = 0
    let to_str = function() { `line ${this.row}, column ${this.column}` }
    for (let i=0; i<string.length; i++) {
        let char = string[i]
        if (char == '\n') {
            row += 1
            column = 0
        } else if (char != '\r') {
            column += 1
        }
        info[i] = { row: row, column: column, toString: to_str }
    }
    return info
}


function get_tokens (string) {
    check(get_tokens, arguments, { string: Str })
    let tokens = []
    let pos = 0
    let rest = string
    let info = row_col_info(string)
    while (rest != '') {
        let matched = take_if(
            map_lazy( TOKEN_ORDER, type => ({
                type: type,
                object: rest.match(TOKEN[type].pattern)
            })),
            matched => matched.object !== null
        )
        if ( matched !== NA ) {
            if ( matched.type != 'space' ) {
                tokens.push({
                    type: matched.type,
                    info: info[pos],
                    content: matched.object[TOKEN[matched.type].extract]
                })
            }
            let matched_length = matched.object[0].length
            pos = pos + matched_length
            rest = rest.slice(matched_length, string.length)            
        } else {
            throw new ScanningError(`Invalid token at ${info[pos]}`)
        }
    }
    return tokens
}
